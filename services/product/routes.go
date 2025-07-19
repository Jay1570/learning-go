package product

import (
	"fmt"
	"net/http"

	"github.com/Jay1570/learning-go/types"
	"github.com/Jay1570/learning-go/utils"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	store types.ProductStore
}

func NewHandler(store types.ProductStore) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(router *http.ServeMux) {
	router.HandleFunc("/products", h.handleProducts)
	// router.HandleFunc("/products", h.handleRegister)
}

func (h *Handler) handleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetProducts(w)

	case http.MethodPost:
		h.handleCreateProduct(w, r)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (h *Handler) handleGetProducts(w http.ResponseWriter) {
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
