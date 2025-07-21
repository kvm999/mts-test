package rest

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// Pagination represents pagination information
// @Description Pagination metadata for API responses
type Pagination struct {
	// Page number (1-based)
	// @Description Current page number
	// @Example 1
	Page int `json:"page" example:"1"`

	// Size (items per page)
	// @Description Number of items per page
	// @Example 10
	Size int `json:"size" example:"10"`

	// Total count
	// @Description Total number of items available
	// @Example 100
	Total int `json:"total" example:"100"`

	// Total pages
	// @Description Total number of pages
	// @Example 10
	TotalPages int `json:"total_pages" example:"10"`
} // @name Pagination

func (p *Pagination) Limit() int {
	if p.Size <= 0 {
		return 10
	}
	if p.Size > 100 {
		return 100
	}
	return p.Size
}

func (p *Pagination) Offset() int {
	if p.Page <= 0 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}

func NewPaginationFromRequest(c fiber.Ctx) *Pagination {
	page := 1
	size := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			size = s
		}
	}

	return &Pagination{
		Page: page,
		Size: size,
	}
}

func (p *Pagination) CalculateTotalPages() {
	if p.Size <= 0 {
		p.TotalPages = 0
		return
	}
	p.TotalPages = (p.Total + p.Size - 1) / p.Size
}
