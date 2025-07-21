package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"mts/internal/domain"
)

type orderDto struct {
	Id        uuid.UUID `db:"id"`
	UserId    uuid.UUID `db:"user_id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type orderItemDto struct {
	Id              uuid.UUID `db:"id"`
	OrderId         uuid.UUID `db:"order_id"`
	ProductId       uuid.UUID `db:"product_id"`
	Quantity        int       `db:"quantity"`
	ProductSnapshot string    `db:"product_snapshot"` // JSON encoded ProductSnapshot
	CreatedAt       time.Time `db:"created_at"`
}

func (dto *orderDto) toDomain() (*domain.Order, error) {
	return &domain.Order{
		Id:        dto.Id,
		UserId:    dto.UserId,
		Status:    dto.Status,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		Items:     []*domain.OrderItem{}, // Items will be loaded separately
	}, nil
}

func toOrderDto(order *domain.Order) (*orderDto, error) {
	return &orderDto{
		Id:        order.Id,
		UserId:    order.UserId,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}, nil
}

func (dto *orderItemDto) toDomain() (*domain.OrderItem, error) {
	item := &domain.OrderItem{
		Id:        dto.Id,
		OrderId:   dto.OrderId,
		ProductId: dto.ProductId,
		Quantity:  dto.Quantity,
		CreatedAt: dto.CreatedAt,
	}

	if dto.ProductSnapshot != "" {
		var snapshot domain.ProductSnapshot
		if err := json.Unmarshal([]byte(dto.ProductSnapshot), &snapshot); err != nil {
			return nil, err
		}
		item.ProductSnapshot = snapshot
	}

	return item, nil
}

func toOrderItemDto(item *domain.OrderItem) (*orderItemDto, error) {
	dto := &orderItemDto{
		Id:        item.Id,
		OrderId:   item.OrderId,
		ProductId: item.ProductId,
		Quantity:  item.Quantity,
		CreatedAt: item.CreatedAt,
	}

	snapshotJson, err := json.Marshal(item.ProductSnapshot)
	if err != nil {
		return nil, err
	}
	dto.ProductSnapshot = string(snapshotJson)

	return dto, nil
}
