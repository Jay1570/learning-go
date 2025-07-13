package user

import (
	"database/sql"
	"fmt"

	"github.com/Jay1570/learning-go/db"
	"github.com/Jay1570/learning-go/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	user, err := db.FindOne[types.User](s.db, "users", &db.QueryOptions{
		Where:     "email = ?",
		WhereArgs: []interface{}{email},
	})

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByID(id int) (*types.User, error) {
	return nil, nil
}

func (s *Store) CreateUser(user types.User) error {
	_, err := db.InsertOne[types.User](s.db, "users", user)
	return err
}
