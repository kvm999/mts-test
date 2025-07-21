package rest

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"mts/internal/domain"
)

type productHandler struct {
	productAppService domain.ProductAppService
}

func newProductHandler(productAppService domain.ProductAppService) *productHandler {
	return &productHandler{
		productAppService: productAppService,
	}
}

// createProduct creates a new product in the system
// @Summary Create new product
// @Description Create a new product with description, tags, and initial quantity
// @Tags Products
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "Product creation data"
// @Success 201 {object} Product "Product created successfully"
// @Failure 400 {object} ErrorResponse "Bad request - validation failed"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products [post]
func (h *productHandler) createProduct(c fiber.Ctx) error {
	var req CreateProductRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	product, err := h.productAppService.CreateProduct(c.Context(), req.ToDomain())
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, domain.ErrProductValidation) {
			status = fiber.StatusBadRequest
		}
		return fiber.NewError(status, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(NewProduct(product))
}

// getProducts retrieves a paginated list of products
// @Summary Get products list
// @Description Retrieve a paginated list of all products in the system
// @Tags Products
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1) minimum(1)
// @Param size query int false "Number of items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} ProductsResponse "Products retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid pagination parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products [get]
func (h *productHandler) getProducts(c fiber.Ctx) error {
	pagination := NewPaginationFromRequest(c)

	products, err := h.productAppService.Products(c.Context(), &domain.GetProductsRequest{})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	count, err := h.productAppService.CountProducts(c.Context(), &domain.GetProductsRequest{})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	pagination.Total = count
	pagination.CalculateTotalPages()

	return c.JSON(NewProductsResponse(products, *pagination))
}

// getProduct retrieves a specific product by ID
// @Summary Get product by ID
// @Description Retrieve detailed information about a specific product using its unique identifier
// @Tags Products
// @Accept json
// @Produce json
// @Param product_id path string true "Product unique identifier" format(uuid)
// @Success 200 {object} Product "Product information retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid product ID format"
// @Failure 404 {object} ErrorResponse "Not found - product with specified ID does not exist"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products/{product_id} [get]
func (h *productHandler) getProduct(c fiber.Ctx) error {
	productId, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product ID format")
	}

	products, err := h.productAppService.Products(c.Context(), &domain.GetProductsRequest{
		Ids: []uuid.UUID{productId},
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if len(products) == 0 {
		return fiber.NewError(fiber.StatusNotFound, domain.ErrProductNotFound.Error())
	}

	return c.JSON(NewProduct(products[0]))
}

// updateProduct updates an existing product
// @Summary Update product
// @Description Update an existing product's information including description, tags, and quantity
// @Tags Products
// @Accept json
// @Produce json
// @Param product_id path string true "Product unique identifier" format(uuid)
// @Param request body UpdateProductRequest true "Product update data"
// @Success 200 {object} Product "Product updated successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid product ID format or validation failed"
// @Failure 404 {object} ErrorResponse "Not found - product with specified ID does not exist"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/products/{product_id} [put]
func (h *productHandler) updateProduct(c fiber.Ctx) error {
	productId, err := uuid.Parse(c.Params("product_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product ID format")
	}

	var req UpdateProductRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	updateReq := req.ToDomain(productId)

	product, err := h.productAppService.UpdateProduct(c.Context(), updateReq)
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, domain.ErrProductNotFound) {
			status = fiber.StatusNotFound
		} else if errors.Is(err, domain.ErrProductValidation) {
			status = fiber.StatusBadRequest
		}
		return fiber.NewError(status, err.Error())
	}

	return c.JSON(NewProduct(product))
}
