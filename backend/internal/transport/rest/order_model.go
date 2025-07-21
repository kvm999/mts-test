package rest

import (
	"time"

	"github.com/google/uuid"

	"mts/internal/domain"
)

// OrderItem represents an item in an order
// @Description Order item with historical product information
type OrderItem struct {
	// Item ID
	// @Description Unique identifier for the order item
	// @Example 789e0123-e45f-67g8-h901-234567890123
	Id uuid.UUID `json:"id" example:"789e0123-e45f-67g8-h901-234567890123" swaggertype:"string"`

	// Product ID
	// @Description ID of the product (current)
	// @Example 456e7890-e12b-34d5-a678-901234567890
	ProductId uuid.UUID `json:"product_id" example:"456e7890-e12b-34d5-a678-901234567890" swaggertype:"string"`

	// Quantity
	// @Description Quantity of the product in the order
	// @Example 2
	Quantity int `json:"quantity" example:"2"`

	// Product snapshot
	// @Description Historical product information at order time
	ProductSnapshot ProductSnapshot `json:"product_snapshot"`

	// Created at
	// @Description When the order item was created
	// @Example 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
} // @name OrderItem

// ProductSnapshot represents historical product information
// @Description Historical product data captured at order time
type ProductSnapshot struct {
	// Description
	// @Description Product description at order time
	// @Example "High-quality smartphone"
	Description string `json:"description" example:"High-quality smartphone"`

	// Tags
	// @Description Product tags at order time
	// @Example ["electronics", "mobile"]
	Tags []string `json:"tags" example:"electronics,mobile"`
} // @name ProductSnapshot

// Order represents an order in the API
// @Description Order information with items
type Order struct {
	// Order ID
	// @Description Unique identifier for the order
	// @Example 987e6543-e21d-12c3-b456-426614174000
	Id uuid.UUID `json:"id" example:"987e6543-e21d-12c3-b456-426614174000" swaggertype:"string"`

	// User ID
	// @Description ID of the user who created the order
	// @Example 123e4567-e89b-12d3-a456-426614174000
	UserId uuid.UUID `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000" swaggertype:"string"`

	// Status
	// @Description Current order status
	// @Example "pending"
	Status string `json:"status" example:"pending" enum:"pending,confirmed,cancelled,completed"`

	// Items
	// @Description List of items in the order
	Items []*OrderItem `json:"items"`

	// Total quantity
	// @Description Total quantity of all items in the order
	// @Example 5
	TotalQuantity int `json:"total_quantity" example:"5"`

	// Created at
	// @Description When the order was created
	// @Example 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Updated at
	// @Description When the order was last updated
	// @Example 2024-01-15T10:30:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
} // @name Order

// CreateOrderItemRequest represents request to add an item to order
// @Description Request item for creating an order
type CreateOrderItemRequest struct {
	// Product ID
	// @Description ID of the product to order
	// @Example 456e7890-e12b-34d5-a678-901234567890
	ProductId uuid.UUID `json:"product_id" binding:"required" validate:"required" example:"456e7890-e12b-34d5-a678-901234567890" swaggertype:"string"`

	// Quantity
	// @Description Quantity to order
	// @Example 2
	Quantity int `json:"quantity" binding:"required" validate:"required,gt=0" example:"2"`
} // @name CreateOrderItemRequest

// CreateOrderRequest represents request to create a new order
// @Description Request payload for creating an order
type CreateOrderRequest struct {
	// User ID
	// @Description ID of the user creating the order
	// @Example 123e4567-e89b-12d3-a456-426614174000
	UserId uuid.UUID `json:"user_id" binding:"required" validate:"required" example:"123e4567-e89b-12d3-a456-426614174000" swaggertype:"string"`

	// Items
	// @Description List of items to order (at least one required)
	Items []CreateOrderItemRequest `json:"items" binding:"required" validate:"required,min=1"`
} // @name CreateOrderRequest

func (req *CreateOrderRequest) ToDomain() *domain.CreateOrderRequest {
	items := make([]domain.CreateOrderItemRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.CreateOrderItemRequest{
			ProductId: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	return &domain.CreateOrderRequest{
		UserId: req.UserId,
		Items:  items,
	}
}

// UpdateOrderRequest represents request to update order status
// @Description Request payload for updating order status
type UpdateOrderRequest struct {
	// Status
	// @Description New order status
	// @Example "confirmed"
	Status string `json:"status" binding:"required" validate:"required,oneof=pending confirmed cancelled completed" example:"confirmed"`
} // @name UpdateOrderRequest

func (req *UpdateOrderRequest) ToDomain(orderId uuid.UUID) *domain.UpdateOrderRequest {
	var status domain.OrderStatus
	switch req.Status {
	case "pending":
		status = domain.OrderStatusPending
	case "confirmed":
		status = domain.OrderStatusConfirmed
	case "cancelled":
		status = domain.OrderStatusCancelled
	case "completed":
		status = domain.OrderStatusCompleted
	}

	return &domain.UpdateOrderRequest{
		Id:     orderId,
		Status: status,
	}
}

// OrdersResponse represents paginated list of orders
// @Description Paginated response containing list of orders
type OrdersResponse struct {
	// Orders
	// @Description List of orders
	Orders []*Order `json:"orders"`

	// Pagination
	// @Description Pagination information
	Pagination *Pagination `json:"pagination"`
} // @name OrdersResponse

func NewOrderItem(domainItem *domain.OrderItem) *OrderItem {
	return &OrderItem{
		Id:        domainItem.Id,
		ProductId: domainItem.ProductId,
		Quantity:  domainItem.Quantity,
		ProductSnapshot: ProductSnapshot{
			Description: domainItem.ProductSnapshot.Description,
			Tags:        domainItem.ProductSnapshot.Tags,
		},
		CreatedAt: domainItem.CreatedAt,
	}
}

func NewOrder(domainOrder *domain.Order) *Order {
	items := make([]*OrderItem, 0, len(domainOrder.Items))
	for _, domainItem := range domainOrder.Items {
		items = append(items, NewOrderItem(domainItem))
	}

	return &Order{
		Id:            domainOrder.Id,
		UserId:        domainOrder.UserId,
		Status:        domainOrder.Status,
		Items:         items,
		TotalQuantity: domainOrder.TotalQuantity(),
		CreatedAt:     domainOrder.CreatedAt,
		UpdatedAt:     domainOrder.UpdatedAt,
	}
}

func NewOrdersResponse(domainOrders []*domain.Order, pagination Pagination) *OrdersResponse {
	orders := make([]*Order, 0, len(domainOrders))
	for _, domainOrder := range domainOrders {
		orders = append(orders, NewOrder(domainOrder))
	}

	return &OrdersResponse{
		Orders:     orders,
		Pagination: &pagination,
	}
}
