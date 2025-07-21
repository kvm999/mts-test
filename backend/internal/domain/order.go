package domain

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderStatus = string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusCompleted OrderStatus = "completed"
)

// OrderItem represents a product in an order with historical information
type OrderItem struct {
	Id        uuid.UUID
	OrderId   uuid.UUID
	ProductId uuid.UUID
	Quantity  int

	// Historical product data - captured at order time
	ProductSnapshot ProductSnapshot

	CreatedAt time.Time
}

// ProductSnapshot captures product state at order time for historical purposes
type ProductSnapshot struct {
	Description string
	Tags        []string
	// В реальном проекте здесь могла бы быть цена
}

func (item *OrderItem) Validate() error {
	if item.Id == uuid.Nil {
		item.Id = uuid.New()
	}

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	if item.OrderId == uuid.Nil {
		return fmt.Errorf("%w: order ID is required", ErrOrderValidation)
	}

	if item.ProductId == uuid.Nil {
		return fmt.Errorf("%w: product ID is required", ErrOrderValidation)
	}

	if item.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", ErrOrderValidation)
	}

	if item.ProductSnapshot.Description == "" {
		return fmt.Errorf("%w: product snapshot description is required", ErrOrderValidation)
	}

	return nil
}

type Order struct {
	Id        uuid.UUID
	UserId    uuid.UUID
	Status    OrderStatus
	Items     []*OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (o *Order) Validate() error {
	if o.Id == uuid.Nil {
		o.Id = uuid.New()
	}

	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}

	o.UpdatedAt = time.Now()

	if o.UserId == uuid.Nil {
		return fmt.Errorf("%w: user ID is required", ErrOrderValidation)
	}

	if o.Status == "" {
		o.Status = OrderStatusPending
	}

	if len(o.Items) == 0 {
		return fmt.Errorf("%w: order must contain at least one item", ErrOrderValidation)
	}

	// Validate all items
	for _, item := range o.Items {
		item.OrderId = o.Id
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (o *Order) TotalQuantity() int {
	total := 0
	for _, item := range o.Items {
		total += item.Quantity
	}
	return total
}

func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending || o.Status == OrderStatusConfirmed
}

func (o *Order) Cancel() error {
	if !o.CanBeCancelled() {
		return fmt.Errorf("%w: order cannot be cancelled in status %s", ErrOrderValidation, o.Status)
	}

	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) Confirm() error {
	if o.Status != OrderStatusPending {
		return fmt.Errorf("%w: order cannot be confirmed in status %s", ErrOrderValidation, o.Status)
	}

	o.Status = OrderStatusConfirmed
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) Complete() error {
	if o.Status != OrderStatusConfirmed {
		return fmt.Errorf("%w: order cannot be completed in status %s", ErrOrderValidation, o.Status)
	}

	o.Status = OrderStatusCompleted
	o.UpdatedAt = time.Now()
	return nil
}

type CreateOrderItemRequest struct {
	ProductId uuid.UUID
	Quantity  int
}

func (r *CreateOrderItemRequest) Validate() error {
	if r.ProductId == uuid.Nil {
		return fmt.Errorf("%w: product ID is required", ErrOrderValidation)
	}

	if r.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", ErrOrderValidation)
	}

	return nil
}

type CreateOrderRequest struct {
	UserId uuid.UUID
	Items  []CreateOrderItemRequest
}

func (r *CreateOrderRequest) Validate() error {
	if r.UserId == uuid.Nil {
		return fmt.Errorf("%w: user ID is required", ErrOrderValidation)
	}

	if len(r.Items) == 0 {
		return fmt.Errorf("%w: order must contain at least one item", ErrOrderValidation)
	}

	for _, item := range r.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

type UpdateOrderRequest struct {
	Id     uuid.UUID
	Status OrderStatus
}

func (r *UpdateOrderRequest) Validate() error {
	if r.Id == uuid.Nil {
		return fmt.Errorf("%w: order ID is required", ErrOrderValidation)
	}

	validStatuses := []OrderStatus{OrderStatusPending, OrderStatusConfirmed, OrderStatusCancelled, OrderStatusCompleted}
	isValid := false
	for _, status := range validStatuses {
		if r.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("%w: invalid order status %s", ErrOrderValidation, r.Status)
	}

	return nil
}

type GetOrdersRequest struct {
	Ids      []uuid.UUID
	UserIds  []uuid.UUID
	Statuses []OrderStatus
	Limit    int
	Offset   int
}

func (r *GetOrdersRequest) Validate() {
	if r.Limit <= 0 {
		r.Limit = 10
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
}

func (r *GetOrdersRequest) CacheKey() CacheKey {
	buf := make([]byte, 0)

	// ids
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.Ids)))
	for _, id := range r.Ids {
		buf = append(buf, id[:]...)
	}

	// user ids
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.UserIds)))
	for _, id := range r.UserIds {
		buf = append(buf, id[:]...)
	}

	// statuses
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.Statuses)))
	for _, status := range r.Statuses {
		buf = append(buf, []byte(status)...)
	}

	// pagination
	buf = binary.BigEndian.AppendUint32(buf, uint32(r.Limit))
	buf = binary.BigEndian.AppendUint32(buf, uint32(r.Offset))

	return sha256.Sum256(buf)
}

type OrderStorage interface {
	CreateOrder(ctx context.Context, order *Order) error
	UpdateOrder(ctx context.Context, req *UpdateOrderRequest) (*Order, error)
	Orders(ctx context.Context, req *GetOrdersRequest) ([]*Order, error)
	CountOrders(ctx context.Context, req *GetOrdersRequest) (int, error)
}

type OrderAppService interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)
	UpdateOrder(ctx context.Context, req *UpdateOrderRequest) (*Order, error)
	Orders(ctx context.Context, req *GetOrdersRequest) ([]*Order, error)
	CancelOrder(ctx context.Context, orderId uuid.UUID) (*Order, error)
}
