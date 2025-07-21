package storage

import (
	"context"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jellydator/ttlcache/v3"

	"mts/internal/domain"
)

func NewProductStorage(pool *pgxpool.Pool) domain.ProductStorage {
	return &productStorage{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		cache: ttlcache.New[domain.CacheKey, []*domain.Product](
			ttlcache.WithTTL[domain.CacheKey, []*domain.Product](time.Hour),
		),
	}
}

type productStorage struct {
	pool  *pgxpool.Pool
	psql  sq.StatementBuilderType
	cache *ttlcache.Cache[domain.CacheKey, []*domain.Product]
}

func (s *productStorage) CreateProduct(ctx context.Context, product *domain.Product) error {
	s.cache.DeleteAll()

	if err := product.Validate(); err != nil {
		return err
	}

	dto, err := toProductDto(product)
	if err != nil {
		return err
	}

	query := s.psql.Insert("products").
		Columns("id", "description", "tags", "quantity", "created_at", "updated_at").
		Values(dto.Id, dto.Description, dto.Tags, dto.Quantity, dto.CreatedAt, dto.UpdatedAt)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *productStorage) UpdateProduct(ctx context.Context, req *domain.UpdateProductRequest) (*domain.Product, error) {
	s.cache.DeleteAll()

	if err := req.Validate(); err != nil {
		return nil, err
	}

	updateQuery := s.psql.Update("products").
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.Id})

	if req.Description != nil {
		updateQuery = updateQuery.Set("description", strings.TrimSpace(*req.Description))
	}

	if req.Quantity != nil {
		updateQuery = updateQuery.Set("quantity", *req.Quantity)
	}

	if len(req.Tags) > 0 {
		// Convert tags to JSON
		product := &domain.Product{Tags: req.Tags}
		dto, err := toProductDto(product)
		if err != nil {
			return nil, err
		}
		updateQuery = updateQuery.Set("tags", dto.Tags)
	}

	sql, args, err := updateQuery.ToSql()
	if err != nil {
		return nil, err
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Get updated product
	products, err := s.Products(ctx, &domain.GetProductsRequest{
		Ids:   []uuid.UUID{req.Id},
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, domain.ErrProductNotFound
	}

	return products[0], nil
}

func (s *productStorage) Products(ctx context.Context, req *domain.GetProductsRequest) ([]*domain.Product, error) {
	s.cache.DeleteExpired()
	req.Validate()

	if cacheProducts := s.cache.Get(req.CacheKey()); cacheProducts != nil {
		return cacheProducts.Value(), nil
	}

	query := s.psql.Select("id", "description", "tags", "quantity", "created_at", "updated_at").
		From("products")

	if len(req.Ids) > 0 {
		query = query.Where(sq.Eq{"id": req.Ids})
	}

	if len(req.Tags) > 0 {
		// Search for products that contain any of the specified tags
		for _, tag := range req.Tags {
			query = query.Where(sq.Like{"tags": "%" + tag + "%"})
		}
	}

	if req.Available != nil {
		if *req.Available {
			query = query.Where(sq.Gt{"quantity": 0})
		} else {
			query = query.Where(sq.Eq{"quantity": 0})
		}
	}

	query = query.OrderBy("created_at DESC", "id").
		Limit(uint64(req.Limit)).
		Offset(uint64(req.Offset))

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var dto productDto

		err := rows.Scan(&dto.Id, &dto.Description, &dto.Tags, &dto.Quantity, &dto.CreatedAt, &dto.UpdatedAt)
		if err != nil {
			return nil, err
		}

		product, err := dto.toDomain()
		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	s.cache.Set(req.CacheKey(), products, ttlcache.DefaultTTL)

	return products, nil
}

func (s *productStorage) CountProducts(ctx context.Context, req *domain.GetProductsRequest) (int, error) {
	req.Validate()

	query := s.psql.Select("COUNT(*)").
		From("products")

	if len(req.Ids) > 0 {
		query = query.Where(sq.Eq{"id": req.Ids})
	}

	if len(req.Tags) > 0 {
		for _, tag := range req.Tags {
			query = query.Where(sq.Like{"tags": "%" + tag + "%"})
		}
	}

	if req.Available != nil {
		if *req.Available {
			query = query.Where(sq.Gt{"quantity": 0})
		} else {
			query = query.Where(sq.Eq{"quantity": 0})
		}
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var count int
	err = s.pool.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
