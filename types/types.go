package types

import (
	"time"
)

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) error
}

type ProductStore interface {
	GetProducts() ([]Product, error)
	CreateProduct(Product) error
}

type User struct {
	ID        int       `json:"id" db:"id" insert:"-"`
	FirstName string    `json:"firstName" db:"firstName" insert:"firstName"`
	LastName  string    `json:"lastName" db:"lastName" insert:"lastName"`
	Email     string    `json:"email" db:"email" insert:"email"`
	Password  string    `json:"-" db:"password" insert:"password"`
	CreatedAt time.Time `json:"createdAt" db:"createdAt" insert:"-"`
}

type Product struct {
	ID          int       `json:"id" db:"id" insert:"-"`
	Name        string    `json:"name" db:"name" insert:"name"`
	Description string    `json:"description" db:"description" insert:"description"`
	Image       string    `json:"image" db:"image" insert:"image"`
	Price       float64   `json:"price" db:"price" insert:"price"`
	Quantity    int       `json:"quantity" db:"quantity" insert:"quantity"`
	CreatedAt   time.Time `json:"createdAt" db:"createdAt" insert:"-"`
}

type RegisterUserPayload struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=3,max=130"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateProductPayload struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Image       string  `json:"image"`
	Price       float64 `json:"price" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required"`
}
