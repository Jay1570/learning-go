package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Jay1570/learning-go/services/logging"
	"github.com/Jay1570/learning-go/services/product"
	"github.com/Jay1570/learning-go/services/user"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := http.NewServeMux()
	subrouter := http.NewServeMux()

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	productStore := product.NewStore(s.db)
	productHandler := product.NewHandler(productStore, userStore)
	productHandler.RegisterRoutes(subrouter)

	router.Handle("/api/", http.StripPrefix("/api/v1", subrouter))

	log.Println("Listening on", s.addr)

	return http.ListenAndServe(s.addr, logging.Logging(router))
}
