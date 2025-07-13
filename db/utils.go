package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

type CountResult[T any] struct {
	Data  []T `json:"data"`
	Count int `json:"count"`
}

type QueryOptions struct {
	Where     string        `json:"where,omitempty"`
	WhereArgs []interface{} `json:"whereArgs,omitempty"`
	OrderBy   string        `json:"orderBy,omitempty"`
	Limit     int           `json:"limit,omitempty"`
	Offset    int           `json:"offset,omitempty"`
}

func FindAllAndCount[T any](db *sql.DB, tableName string, options *QueryOptions) (*CountResult[T], error) {
	var result CountResult[T]

	whereClause, args := buildWhereClause(options)

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", tableName, whereClause)
	err := db.QueryRow(countQuery, args...).Scan(&result.Count)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	selectQuery := buildSelectQuery(tableName, options, whereClause)
	rows, err := db.Query(selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	result.Data, err = scanRows[T](rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan records: %w", err)
	}

	return &result, nil
}

func FindAll[T any](db *sql.DB, tableName string, options *QueryOptions) ([]T, error) {
	whereClause, args := buildWhereClause(options)
	query := buildSelectQuery(tableName, options, whereClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	return scanRows[T](rows)
}

func FindOne[T any](db *sql.DB, tableName string, options *QueryOptions) (*T, error) {
	if options == nil {
		options = &QueryOptions{}
	}
	options.Limit = 1

	records, err := FindAll[T](db, tableName, options)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, sql.ErrNoRows
	}

	return &records[0], nil
}

func FindByPK[T any](db *sql.DB, tableName string, pk interface{}) (*T, error) {
	options := &QueryOptions{
		Where:     "id = ?",
		WhereArgs: []interface{}{pk},
	}

	return FindOne[T](db, tableName, options)
}

func InsertOne[T any](db *sql.DB, tableName string, payload interface{}) (int64, error) {
	columns, placeholders, values := buildInsertData(payload)

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	result, err := db.Exec(query, values...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert record: %w", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return lastID, nil
}

func BulkInsert[T any](db *sql.DB, tableName string, payloads []interface{}) (bool, error) {
	if len(payloads) == 0 {
		return true, nil
	}

	tx, err := db.Begin()
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, payload := range payloads {
		columns, placeholders, values := buildInsertData(payload)

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

		_, err := tx.Exec(query, values...)
		if err != nil {
			return false, fmt.Errorf("failed to insert record: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return true, nil
}

func UpdateData[T any](db *sql.DB, tableName string, payload interface{}, options *QueryOptions) ([]T, error) {
	setClause, setArgs := buildSetClause(payload)
	whereClause, whereArgs := buildWhereClause(options)

	args := append(setArgs, whereArgs...)

	query := fmt.Sprintf("UPDATE %s SET %s%s RETURNING *", tableName, setClause, whereClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update records: %w", err)
	}
	defer rows.Close()

	return scanRows[T](rows)
}

func DeleteData[T any](db *sql.DB, tableName string, options *QueryOptions) ([]T, error) {
	whereClause, args := buildWhereClause(options)

	query := fmt.Sprintf("DELETE FROM %s%s RETURNING *", tableName, whereClause)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to delete records: %w", err)
	}
	defer rows.Close()

	return scanRows[T](rows)
}

func buildWhereClause(options *QueryOptions) (string, []interface{}) {
	if options == nil || options.Where == "" {
		return "", nil
	}
	return " WHERE " + options.Where, options.WhereArgs
}

func buildSelectQuery(tableName string, options *QueryOptions, whereClause string) string {
	query := fmt.Sprintf("SELECT * FROM %s%s", tableName, whereClause)

	if options != nil {
		if options.OrderBy != "" {
			query += " ORDER BY " + options.OrderBy
		}

		if options.Limit > 0 {
			query += fmt.Sprintf(" LIMIT %d", options.Limit)
		}

		if options.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", options.Offset)
		}
	}

	return query
}

func buildInsertData(payload interface{}) ([]string, []string, []interface{}) {
	v := reflect.ValueOf(payload)
	t := reflect.TypeOf(payload)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	var columns []string
	var placeholders []string
	var values []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		columnName := strings.Split(dbTag, ",")[0]
		if columnName == "" {
			continue
		}

		if columnName == "id" || columnName == "createdAt" {
			continue
		}

		columns = append(columns, columnName)
		placeholders = append(placeholders, "?")
		values = append(values, field.Interface())
	}

	return columns, placeholders, values
}

func buildSetClause(payload interface{}) (string, []interface{}) {
	v := reflect.ValueOf(payload)
	t := reflect.TypeOf(payload)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	var setParts []string
	var values []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		dbTag := fieldType.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		columnName := strings.Split(dbTag, ",")[0]
		if columnName == "" {
			continue
		}

		if columnName == "id" || columnName == "createdAt" {
			continue
		}

		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		if field.Kind() == reflect.String && field.String() == "" {
			continue
		}

		setParts = append(setParts, fmt.Sprintf("%s = ?", columnName))
		values = append(values, field.Interface())
	}

	return strings.Join(setParts, ", "), values
}

func scanRows[T any](rows *sql.Rows) ([]T, error) {
	var results []T

	for rows.Next() {
		var item T
		err := scanRow(rows, &item)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, rows.Err()
}

func scanRow(scanner interface{}, dest interface{}) error {
	v := reflect.ValueOf(dest).Elem()

	fieldCount := v.NumField()
	scanArgs := make([]interface{}, fieldCount)

	for i := 0; i < fieldCount; i++ {
		field := v.Field(i)
		if field.CanAddr() {
			scanArgs[i] = field.Addr().Interface()
		} else {

			temp := reflect.New(field.Type())
			scanArgs[i] = temp.Interface()
		}
	}

	var err error
	switch s := scanner.(type) {
	case *sql.Row:
		err = s.Scan(scanArgs...)
	case *sql.Rows:
		err = s.Scan(scanArgs...)
	default:
		return fmt.Errorf("unsupported scanner type")
	}

	if err != nil {
		return err
	}

	for i := 0; i < fieldCount; i++ {
		field := v.Field(i)
		if !field.CanAddr() {
			field.Set(reflect.ValueOf(scanArgs[i]).Elem())
		}
	}

	return nil
}
