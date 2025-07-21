package rest

import (
	"time"

	"github.com/google/uuid"

	"mts/internal/domain"
)

// Product represents a product in the API
// @Description Product information
type Product struct {
	// Product ID
	// @Description Unique identifier for the product
	// @Example 456e7890-e12b-34d5-a678-901234567890
	Id uuid.UUID `json:"id" example:"456e7890-e12b-34d5-a678-901234567890" swaggertype:"string"`

	// Description
	// @Description Product description
	// @Example "High-quality smartphone"
	Description string `json:"description" example:"High-quality smartphone"`

	// Tags
	// @Description Product tags for categorization
	// @Example ["electronics", "mobile"]
	Tags []string `json:"tags" example:"electronics,mobile"`

	// Quantity
	// @Description Available quantity in stock
	// @Example 100
	Quantity int `json:"quantity" example:"100"`

	// Available
	// @Description Whether the product is available (quantity > 0)
	// @Example true
	Available bool `json:"available" example:"true"`

	// Created at
	// @Description When the product was created
	// @Example 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`

	// Updated at
	// @Description When the product was last updated
	// @Example 2024-01-15T10:30:00Z
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
} // @name Product

// CreateProductRequest represents request to create a new product
// @Description Request payload for creating a product
type CreateProductRequest struct {
	// Description
	// @Description Product description (required)
	// @Example "High-quality smartphone"
	Description string `json:"description" binding:"required" validate:"required" example:"High-quality smartphone"`

	// Tags
	// @Description Product tags for categorization
	// @Example ["electronics", "mobile"]
	Tags []string `json:"tags" example:"electronics,mobile"`

	// Quantity
	// @Description Initial quantity in stock
	// @Example 100
	Quantity int `json:"quantity" binding:"required" validate:"gte=0" example:"100"`
} // @name CreateProductRequest

func (req *CreateProductRequest) ToDomain() *domain.CreateProductRequest {
	return &domain.CreateProductRequest{
		Description: req.Description,
		Tags:        req.Tags,
		Quantity:    req.Quantity,
	}
}

// UpdateProductRequest represents request to update a product
// @Description Request payload for updating a product
type UpdateProductRequest struct {
	// Description
	// @Description Product description (optional)
	// @Example "Updated smartphone description"
	Description *string `json:"description,omitempty" example:"Updated smartphone description"`

	// Tags
	// @Description Product tags for categorization (optional)
	// @Example ["electronics", "mobile", "updated"]
	Tags []string `json:"tags,omitempty" example:"electronics,mobile,updated"`

	// Quantity
	// @Description Quantity in stock (optional)
	// @Example 150
	Quantity *int `json:"quantity,omitempty" validate:"omitempty,gte=0" example:"150"`
} // @name UpdateProductRequest

func (req *UpdateProductRequest) ToDomain(productId uuid.UUID) *domain.UpdateProductRequest {
	return &domain.UpdateProductRequest{
		Id:          productId,
		Description: req.Description,
		Tags:        req.Tags,
		Quantity:    req.Quantity,
	}
}

// ProductsResponse represents paginated list of products
// @Description Paginated response containing list of products
type ProductsResponse struct {
	// Products
	// @Description List of products
	Products []*Product `json:"products"`

	// Pagination
	// @Description Pagination information
	Pagination *Pagination `json:"pagination"`
} // @name ProductsResponse

func NewProduct(domainProduct *domain.Product) *Product {
	return &Product{
		Id:          domainProduct.Id,
		Description: domainProduct.Description,
		Tags:        domainProduct.Tags,
		Quantity:    domainProduct.Quantity,
		Available:   domainProduct.IsAvailable(),
		CreatedAt:   domainProduct.CreatedAt,
		UpdatedAt:   domainProduct.UpdatedAt,
	}
}

func NewProductsResponse(domainProducts []*domain.Product, pagination Pagination) *ProductsResponse {
	products := make([]*Product, 0, len(domainProducts))
	for _, domainProduct := range domainProducts {
		products = append(products, NewProduct(domainProduct))
	}

	return &ProductsResponse{
		Products:   products,
		Pagination: &pagination,
	}
}
