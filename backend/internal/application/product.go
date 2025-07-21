package application

import (
	"context"

	"mts/internal/domain"

	"github.com/rs/zerolog"
)

func NewProductAppService(productStorage domain.ProductStorage) domain.ProductAppService {
	return &productAppService{
		productStorage: productStorage,
	}
}

type productAppService struct {
	productStorage domain.ProductStorage
}

func (s *productAppService) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "CreateProduct").
		Str("description", req.Description).
		Int("quantity", req.Quantity).
		Logger()

	logger.Info().Msg("creating new product")

	product, err := req.ToDomain()
	if err != nil {
		logger.Error().Err(err).Msg("failed to convert request to domain")
		return nil, err
	}

	err = s.productStorage.CreateProduct(ctx, product)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create product in storage")
		return nil, err
	}

	logger.Info().
		Str("product_id", product.Id.String()).
		Msg("product created successfully")

	return product, nil
}

func (s *productAppService) UpdateProduct(ctx context.Context, req *domain.UpdateProductRequest) (*domain.Product, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "UpdateProduct").
		Str("product_id", req.Id.String()).
		Logger()

	logger.Info().Msg("updating product")

	product, err := s.productStorage.UpdateProduct(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update product in storage")
		return nil, err
	}

	logger.Info().Msg("product updated successfully")

	return product, nil
}

func (s *productAppService) Products(ctx context.Context, req *domain.GetProductsRequest) ([]*domain.Product, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "Products").
		Int("ids_count", len(req.Ids)).
		Logger()

	logger.Info().Msg("fetching products")

	products, err := s.productStorage.Products(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch products from storage")
		return nil, err
	}

	logger.Info().
		Int("products_count", len(products)).
		Msg("products fetched successfully")

	return products, nil
}

func (s *productAppService) CountProducts(ctx context.Context, req *domain.GetProductsRequest) (int, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "CountProducts").
		Int("ids_count", len(req.Ids)).
		Logger()

	logger.Debug().Msg("counting products")

	count, err := s.productStorage.CountProducts(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to count products in storage")
		return 0, err
	}

	logger.Debug().
		Int("count", count).
		Msg("products counted successfully")

	return count, nil
}
