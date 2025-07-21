package rest

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"mts/internal/domain"
)

type orderHandler struct {
	orderAppService domain.OrderAppService
}

func newOrderHandler(orderAppService domain.OrderAppService) *orderHandler {
	return &orderHandler{
		orderAppService: orderAppService,
	}
}

// createOrder creates a new order in the system
// @Summary Create new order
// @Description Create a new order with multiple items, automatically handles stock reservation
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order creation data"
// @Success 201 {object} Order "Order created successfully"
// @Failure 400 {object} ErrorResponse "Bad request - validation failed or insufficient stock"
// @Failure 404 {object} ErrorResponse "Not found - user or product not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/orders [post]
func (h *orderHandler) createOrder(c fiber.Ctx) error {
	var req CreateOrderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	order, err := h.orderAppService.CreateOrder(c.Context(), req.ToDomain())
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, domain.ErrOrderValidation) {
			status = fiber.StatusBadRequest
		} else if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrProductNotFound) {
			status = fiber.StatusNotFound
		} else if errors.Is(err, domain.ErrInsufficientStock) {
			status = fiber.StatusBadRequest
		}
		return fiber.NewError(status, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(NewOrder(order))
}

// getOrders retrieves a paginated list of orders
// @Summary Get orders list
// @Description Retrieve a paginated list of all orders in the system
// @Tags Orders
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1) minimum(1)
// @Param size query int false "Number of items per page" default(10) minimum(1) maximum(100)
// @Param user_id query string false "Filter orders by user ID" format(uuid)
// @Success 200 {object} OrdersResponse "Orders retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid pagination parameters or user ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/orders [get]
func (h *orderHandler) getOrders(c fiber.Ctx) error {
	pagination := NewPaginationFromRequest(c)

	req := &domain.GetOrdersRequest{}

	// Parse optional user_id filter
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		userId, err := uuid.Parse(userIdStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid user_id format")
		}
		req.UserIds = []uuid.UUID{userId}
	}

	orders, err := h.orderAppService.Orders(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Note: For simplicity, not implementing count for orders in this example
	// In production, you would add CountOrders method to the app service
	pagination.Total = len(orders)
	pagination.CalculateTotalPages()

	return c.JSON(NewOrdersResponse(orders, *pagination))
}

// getOrder retrieves a specific order by ID
// @Summary Get order by ID
// @Description Retrieve detailed information about a specific order using its unique identifier
// @Tags Orders
// @Accept json
// @Produce json
// @Param order_id path string true "Order unique identifier" format(uuid)
// @Success 200 {object} Order "Order information retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid order ID format"
// @Failure 404 {object} ErrorResponse "Not found - order with specified ID does not exist"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/orders/{order_id} [get]
func (h *orderHandler) getOrder(c fiber.Ctx) error {
	orderId, err := uuid.Parse(c.Params("order_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid order ID format")
	}

	orders, err := h.orderAppService.Orders(c.Context(), &domain.GetOrdersRequest{
		Ids: []uuid.UUID{orderId},
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if len(orders) == 0 {
		return fiber.NewError(fiber.StatusNotFound, domain.ErrOrderNotFound.Error())
	}

	return c.JSON(NewOrder(orders[0]))
}

// updateOrder updates an existing order status
// @Summary Update order
// @Description Update an existing order's status or other mutable fields
// @Tags Orders
// @Accept json
// @Produce json
// @Param order_id path string true "Order unique identifier" format(uuid)
// @Param request body UpdateOrderRequest true "Order update data"
// @Success 200 {object} Order "Order updated successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid order ID format or validation failed"
// @Failure 404 {object} ErrorResponse "Not found - order with specified ID does not exist"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/orders/{order_id} [put]
func (h *orderHandler) updateOrder(c fiber.Ctx) error {
	orderId, err := uuid.Parse(c.Params("order_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid order ID format")
	}

	var req UpdateOrderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updateReq := req.ToDomain(orderId)

	order, err := h.orderAppService.UpdateOrder(c.Context(), updateReq)
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, domain.ErrOrderNotFound) {
			status = fiber.StatusNotFound
		} else if errors.Is(err, domain.ErrOrderValidation) {
			status = fiber.StatusBadRequest
		}
		return fiber.NewError(status, err.Error())
	}

	return c.JSON(NewOrder(order))
}

// cancelOrder cancels an order and restores product quantities
// @Summary Cancel order
// @Description Cancel an order and restore product quantities back to inventory
// @Tags Orders
// @Accept json
// @Produce json
// @Param order_id path string true "Order unique identifier" format(uuid)
// @Success 200 {object} Order "Order cancelled successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid order ID format or order cannot be cancelled"
// @Failure 404 {object} ErrorResponse "Not found - order with specified ID does not exist"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/orders/{order_id}/cancel [post]
func (h *orderHandler) cancelOrder(c fiber.Ctx) error {
	orderId, err := uuid.Parse(c.Params("order_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid order ID format")
	}

	updateReq := &domain.UpdateOrderRequest{
		Id:     orderId,
		Status: domain.OrderStatusCancelled,
	}

	order, err := h.orderAppService.UpdateOrder(c.Context(), updateReq)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if errors.Is(err, domain.ErrOrderNotFound) {
			statusCode = fiber.StatusNotFound
		} else if errors.Is(err, domain.ErrOrderValidation) {
			statusCode = fiber.StatusBadRequest
		}
		return fiber.NewError(statusCode, err.Error())
	}

	return c.JSON(NewOrder(order))
}
