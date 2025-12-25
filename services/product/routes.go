package product

import (
	"fmt"
	"net/http"

	"github.com/Jay1570/learning-go/services/auth"
	"github.com/Jay1570/learning-go/types"
	"github.com/Jay1570/learning-go/utils"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	store     types.ProductStore
	userStore types.UserStore
}

func NewHandler(store types.ProductStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	productRouter := http.NewServeMux()

	productRouter.HandleFunc("GET /products", h.handleGetProducts)
	productRouter.HandleFunc("POST /products", h.handleCreateProduct)

	router.Handle("/", auth.WithJWTAuth(productRouter, h.userStore))
	// router.HandleFunc("/products", h.handleRegister)
}

func (h *Handler) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.store.GetProducts()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]any{
		"status":   http.StatusOK,
		"products": products,
	}
	utils.WriteJSON(w, response["status"].(int), response)
}

func (h *Handler) handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	var payload types.CreateProductPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	err := h.store.CreateProduct(types.Product{
		Name:        payload.Name,
		Description: payload.Description,
		Image:       payload.Image,
		Price:       payload.Price,
		Quantity:    payload.Quantity,
	})
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]any{
		"status":  http.StatusCreated,
		"message": "Product successfully created",
	}
	utils.WriteJSON(w, response["status"].(int), response)
}
