package create_order

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/loloneme/pulse-flow/internal/usecase/create_order"
)

type Handler struct {
	createOrderService CreateOrderService
}

func New(createOrderService CreateOrderService) *Handler {
	return &Handler{createOrderService: createOrderService}
}

func (h *Handler) CreateOrder(c echo.Context) error {
	httpReq := new(CreateOrderRequest)
	if err := c.Bind(httpReq); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	usecaseReq := &create_order.CreateOrderRequest{
		UserID:    httpReq.UserID,
		ProductID: httpReq.ProductID,
		Amount:    httpReq.Amount,
	}

	if err := h.createOrderService.CreateOrder(c.Request().Context(), usecaseReq); err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "Order created successfully"})
}
