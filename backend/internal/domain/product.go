package domain

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	Id          uuid.UUID
	Description string
	Tags        []string
	Quantity    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *Product) Validate() error {
	if p.Id == uuid.Nil {
		p.Id = uuid.New()
	}

	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}

	p.UpdatedAt = time.Now()

	if strings.TrimSpace(p.Description) == "" {
		return fmt.Errorf("%w: description is required", ErrProductValidation)
	}

	if p.Quantity < 0 {
		return fmt.Errorf("%w: quantity cannot be negative", ErrProductValidation)
	}

	// Clean up tags
	cleanTags := make([]string, 0, len(p.Tags))
	for _, tag := range p.Tags {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			cleanTags = append(cleanTags, tag)
		}
	}
	p.Tags = cleanTags

	return nil
}

func (p *Product) IsAvailable() bool {
	return p.Quantity > 0
}

func (p *Product) ReserveQuantity(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", ErrInvalidQuantity)
	}

	if p.Quantity < quantity {
		return fmt.Errorf("%w: requested %d but only %d available", ErrInsufficientStock, quantity, p.Quantity)
	}

	p.Quantity -= quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) RestoreQuantity(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", ErrInvalidQuantity)
	}

	p.Quantity += quantity
	p.UpdatedAt = time.Now()
	return nil
}

type CreateProductRequest struct {
	Description string
	Tags        []string
	Quantity    int
}

func (r *CreateProductRequest) Validate() error {
	if strings.TrimSpace(r.Description) == "" {
		return fmt.Errorf("%w: description is required", ErrProductValidation)
	}

	if r.Quantity < 0 {
		return fmt.Errorf("%w: quantity cannot be negative", ErrProductValidation)
	}

	return nil
}

func (r *CreateProductRequest) ToDomain() (*Product, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}

	product := &Product{
		Description: strings.TrimSpace(r.Description),
		Tags:        r.Tags,
		Quantity:    r.Quantity,
	}

	return product, nil
}

type UpdateProductRequest struct {
	Id          uuid.UUID
	Description *string
	Tags        []string
	Quantity    *int
}

func (r *UpdateProductRequest) Validate() error {
	if r.Id == uuid.Nil {
		return fmt.Errorf("%w: product ID is required", ErrProductValidation)
	}

	if r.Description != nil && strings.TrimSpace(*r.Description) == "" {
		return fmt.Errorf("%w: description cannot be empty", ErrProductValidation)
	}

	if r.Quantity != nil && *r.Quantity < 0 {
		return fmt.Errorf("%w: quantity cannot be negative", ErrProductValidation)
	}

	return nil
}

type GetProductsRequest struct {
	Ids       []uuid.UUID
	Tags      []string
	Available *bool
	Limit     int
	Offset    int
}

func (r *GetProductsRequest) Validate() {
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

func (r *GetProductsRequest) CacheKey() CacheKey {
	buf := make([]byte, 0)

	// ids
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.Ids)))
	for _, id := range r.Ids {
		buf = append(buf, id[:]...)
	}

	// tags
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.Tags)))
	for _, tag := range r.Tags {
		buf = append(buf, []byte(tag)...)
	}

	// available filter
	if r.Available != nil {
		if *r.Available {
			buf = append(buf, 1)
		} else {
			buf = append(buf, 0)
		}
	} else {
		buf = append(buf, 2) // nil case
	}

	// pagination
	buf = binary.BigEndian.AppendUint32(buf, uint32(r.Limit))
	buf = binary.BigEndian.AppendUint32(buf, uint32(r.Offset))

	return sha256.Sum256(buf)
}

type ProductStorage interface {
	CreateProduct(ctx context.Context, product *Product) error
	UpdateProduct(ctx context.Context, req *UpdateProductRequest) (*Product, error)
	Products(ctx context.Context, req *GetProductsRequest) ([]*Product, error)
	CountProducts(ctx context.Context, req *GetProductsRequest) (int, error)
}

type ProductAppService interface {
	CreateProduct(ctx context.Context, req *CreateProductRequest) (*Product, error)
	UpdateProduct(ctx context.Context, req *UpdateProductRequest) (*Product, error)
	Products(ctx context.Context, req *GetProductsRequest) ([]*Product, error)
	CountProducts(ctx context.Context, req *GetProductsRequest) (int, error)
}
