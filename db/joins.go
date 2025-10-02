package db

import (
	"database/sql"
	"fmt"
)

// JoinType represents the type of SQL join
type JoinType string

const (
	InnerJoin JoinType = "INNER JOIN"
	LeftJoin  JoinType = "LEFT JOIN"
	RightJoin JoinType = "RIGHT JOIN"
	FullJoin  JoinType = "FULL OUTER JOIN"
)

// JoinClause represents a single join operation
type JoinClause struct {
	Type      JoinType // Type of join (INNER, LEFT, RIGHT, FULL)
	Table     string   // Table to join
	Condition string   // Join condition (e.g., "users.id = orders.user_id")
}

// QueryOptionsWithJoins extends QueryOptions to support joins
type QueryOptionsWithJoins struct {
	Joins     []JoinClause  `json:"joins,omitempty"`
	Where     string        `json:"where,omitempty"`
	WhereArgs []interface{} `json:"whereArgs,omitempty"`
	OrderBy   string        `json:"orderBy,omitempty"`
	Limit     int           `json:"limit,omitempty"`
	Offset    int           `json:"offset,omitempty"`
	Select    string        `json:"select,omitempty"` // Custom SELECT clause
}

// FindAllWithJoins performs a query with joins
func FindAllWithJoins[T any](db *sql.DB, tableName string, options *QueryOptionsWithJoins) ([]T, error) {
	query := buildJoinQuery(tableName, options)

	args := []interface{}{}
	if options != nil && options.WhereArgs != nil {
		args = options.WhereArgs
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query records with joins: %w", err)
	}
	defer rows.Close()

	return scanRows[T](rows)
}

// FindAllAndCountWithJoins performs a count and query with joins
func FindAllAndCountWithJoins[T any](db *sql.DB, tableName string, options *QueryOptionsWithJoins) (*CountResult[T], error) {
	var result CountResult[T]

	// Build count query
	countQuery := buildCountQueryWithJoins(tableName, options)

	args := []interface{}{}
	if options != nil && options.WhereArgs != nil {
		args = options.WhereArgs
	}

	err := db.QueryRow(countQuery, args...).Scan(&result.Count)
	if err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Build select query
	selectQuery := buildJoinQuery(tableName, options)
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

// FindOneWithJoins finds a single record with joins
func FindOneWithJoins[T any](db *sql.DB, tableName string, options *QueryOptionsWithJoins) (*T, error) {
	if options == nil {
		options = &QueryOptionsWithJoins{}
	}
	options.Limit = 1

	records, err := FindAllWithJoins[T](db, tableName, options)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, sql.ErrNoRows
	}

	return &records[0], nil
}

// buildJoinQuery constructs a SELECT query with joins
func buildJoinQuery(tableName string, options *QueryOptionsWithJoins) string {
	selectClause := "*"
	if options != nil && options.Select != "" {
		selectClause = options.Select
	}

	query := fmt.Sprintf("SELECT %s FROM %s", selectClause, tableName)

	// Add joins
	if options != nil && len(options.Joins) > 0 {
		for _, join := range options.Joins {
			query += fmt.Sprintf(" %s %s ON %s", join.Type, join.Table, join.Condition)
		}
	}

	// Add WHERE clause
	if options != nil && options.Where != "" {
		query += " WHERE " + options.Where
	}

	// Add ORDER BY
	if options != nil && options.OrderBy != "" {
		query += " ORDER BY " + options.OrderBy
	}

	// Add LIMIT
	if options != nil && options.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", options.Limit)
	}

	// Add OFFSET
	if options != nil && options.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", options.Offset)
	}

	return query
}

// buildCountQueryWithJoins constructs a COUNT query with joins
func buildCountQueryWithJoins(tableName string, options *QueryOptionsWithJoins) string {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)

	// Add joins
	if options != nil && len(options.Joins) > 0 {
		for _, join := range options.Joins {
			query += fmt.Sprintf(" %s %s ON %s", join.Type, join.Table, join.Condition)
		}
	}

	// Add WHERE clause
	if options != nil && options.Where != "" {
		query += " WHERE " + options.Where
	}

	return query
}

// Helper functions for building joins programmatically

// NewInnerJoin creates an INNER JOIN clause
func NewInnerJoin(table, condition string) JoinClause {
	return JoinClause{
		Type:      InnerJoin,
		Table:     table,
		Condition: condition,
	}
}

// NewLeftJoin creates a LEFT JOIN clause
func NewLeftJoin(table, condition string) JoinClause {
	return JoinClause{
		Type:      LeftJoin,
		Table:     table,
		Condition: condition,
	}
}

// NewRightJoin creates a RIGHT JOIN clause
func NewRightJoin(table, condition string) JoinClause {
	return JoinClause{
		Type:      RightJoin,
		Table:     table,
		Condition: condition,
	}
}

// NewFullJoin creates a FULL OUTER JOIN clause
func NewFullJoin(table, condition string) JoinClause {
	return JoinClause{
		Type:      FullJoin,
		Table:     table,
		Condition: condition,
	}
}

// JoinBuilder provides a fluent interface for building complex joins
type JoinBuilder struct {
	tableName string
	options   *QueryOptionsWithJoins
}

// NewJoinBuilder creates a new JoinBuilder
func NewJoinBuilder(tableName string) *JoinBuilder {
	return &JoinBuilder{
		tableName: tableName,
		options: &QueryOptionsWithJoins{
			Joins: []JoinClause{},
		},
	}
}

// InnerJoin adds an INNER JOIN
func (jb *JoinBuilder) InnerJoin(table, condition string) *JoinBuilder {
	jb.options.Joins = append(jb.options.Joins, NewInnerJoin(table, condition))
	return jb
}

// LeftJoin adds a LEFT JOIN
func (jb *JoinBuilder) LeftJoin(table, condition string) *JoinBuilder {
	jb.options.Joins = append(jb.options.Joins, NewLeftJoin(table, condition))
	return jb
}

// RightJoin adds a RIGHT JOIN
func (jb *JoinBuilder) RightJoin(table, condition string) *JoinBuilder {
	jb.options.Joins = append(jb.options.Joins, NewRightJoin(table, condition))
	return jb
}

// FullJoin adds a FULL OUTER JOIN
func (jb *JoinBuilder) FullJoin(table, condition string) *JoinBuilder {
	jb.options.Joins = append(jb.options.Joins, NewFullJoin(table, condition))
	return jb
}

// Select sets custom SELECT clause
func (jb *JoinBuilder) Select(fields string) *JoinBuilder {
	jb.options.Select = fields
	return jb
}

// Where sets the WHERE clause
func (jb *JoinBuilder) Where(condition string, args ...interface{}) *JoinBuilder {
	jb.options.Where = condition
	jb.options.WhereArgs = args
	return jb
}

// OrderBy sets the ORDER BY clause
func (jb *JoinBuilder) OrderBy(orderBy string) *JoinBuilder {
	jb.options.OrderBy = orderBy
	return jb
}

// Limit sets the LIMIT
func (jb *JoinBuilder) Limit(limit int) *JoinBuilder {
	jb.options.Limit = limit
	return jb
}

// Offset sets the OFFSET
func (jb *JoinBuilder) Offset(offset int) *JoinBuilder {
	jb.options.Offset = offset
	return jb
}

// Build returns the built query options
func (jb *JoinBuilder) Build() *QueryOptionsWithJoins {
	return jb.options
}

// GetQuery returns the built SQL query string (useful for debugging)
func (jb *JoinBuilder) GetQuery() string {
	return buildJoinQuery(jb.tableName, jb.options)
}

// GetTableName returns the base table name
func (jb *JoinBuilder) GetTableName() string {
	return jb.tableName
}

// GetOptions returns the query options
func (jb *JoinBuilder) GetOptions() *QueryOptionsWithJoins {
	return jb.options
}

// Execute executes a join builder and returns results
func Execute[T any](db *sql.DB, builder *JoinBuilder) ([]T, error) {
	return FindAllWithJoins[T](db, builder.GetTableName(), builder.GetOptions())
}

// ExecuteOne executes a join builder and returns a single result
func ExecuteOne[T any](db *sql.DB, builder *JoinBuilder) (*T, error) {
	return FindOneWithJoins[T](db, builder.GetTableName(), builder.GetOptions())
}

// ExecuteWithCount executes a join builder with count
func ExecuteWithCount[T any](db *sql.DB, builder *JoinBuilder) (*CountResult[T], error) {
	return FindAllAndCountWithJoins[T](db, builder.GetTableName(), builder.GetOptions())
}
