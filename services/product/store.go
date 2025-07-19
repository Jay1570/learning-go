package product

import (
	"database/sql"

	"github.com/Jay1570/learning-go/db"
	"github.com/Jay1570/learning-go/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetProducts() ([]types.Product, error) {
	products, err := db.FindAll[types.Product](s.db, "products", &db.QueryOptions{})
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (s *Store) CreateProduct(product types.Product) error {
	_, err := db.InsertOne[types.Product](s.db, "products", product)
	return err
}
