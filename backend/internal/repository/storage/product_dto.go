package storage

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"mts/internal/domain"
)

type productDto struct {
	Id          uuid.UUID `db:"id"`
	Description string    `db:"description"`
	Tags        string    `db:"tags"` // JSON encoded
	Quantity    int       `db:"quantity"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (dto *productDto) toDomain() (*domain.Product, error) {
	product := &domain.Product{
		Id:          dto.Id,
		Description: dto.Description,
		Quantity:    dto.Quantity,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	}

	if dto.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(dto.Tags), &tags); err != nil {
			return nil, err
		}
		product.Tags = tags
	}

	return product, nil
}

func toProductDto(product *domain.Product) (*productDto, error) {
	dto := &productDto{
		Id:          product.Id,
		Description: product.Description,
		Quantity:    product.Quantity,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	if len(product.Tags) > 0 {
		tagsJson, err := json.Marshal(product.Tags)
		if err != nil {
			return nil, err
		}
		dto.Tags = string(tagsJson)
	}

	return dto, nil
}
