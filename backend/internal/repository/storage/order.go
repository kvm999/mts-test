package storage

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jellydator/ttlcache/v3"

	"mts/internal/domain"
)

func NewOrderStorage(pool *pgxpool.Pool) domain.OrderStorage {
	return &orderStorage{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		cache: ttlcache.New[domain.CacheKey, []*domain.Order](
			ttlcache.WithTTL[domain.CacheKey, []*domain.Order](time.Hour),
		),
	}
}

type orderStorage struct {
	pool  *pgxpool.Pool
	psql  sq.StatementBuilderType
	cache *ttlcache.Cache[domain.CacheKey, []*domain.Order]
}

func (s *orderStorage) CreateOrder(ctx context.Context, order *domain.Order) error {
	s.cache.DeleteAll()

	if err := order.Validate(); err != nil {
		return err
	}

	// Start transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert order
	orderDto, err := toOrderDto(order)
	if err != nil {
		return err
	}

	orderQuery := s.psql.Insert("orders").
		Columns("id", "user_id", "status", "created_at", "updated_at").
		Values(orderDto.Id, orderDto.UserId, orderDto.Status, orderDto.CreatedAt, orderDto.UpdatedAt)

	sql, args, err := orderQuery.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		itemDto, err := toOrderItemDto(item)
		if err != nil {
			return err
		}

		itemQuery := s.psql.Insert("order_items").
			Columns("id", "order_id", "product_id", "quantity", "product_snapshot", "created_at").
			Values(itemDto.Id, itemDto.OrderId, itemDto.ProductId, itemDto.Quantity, itemDto.ProductSnapshot, itemDto.CreatedAt)

		sql, args, err := itemQuery.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *orderStorage) UpdateOrder(ctx context.Context, req *domain.UpdateOrderRequest) (*domain.Order, error) {
	s.cache.DeleteAll()

	if err := req.Validate(); err != nil {
		return nil, err
	}

	updateQuery := s.psql.Update("orders").
		Set("status", req.Status).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": req.Id})

	sql, args, err := updateQuery.ToSql()
	if err != nil {
		return nil, err
	}

	result, err := s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if result.RowsAffected() == 0 {
		return nil, domain.ErrOrderNotFound
	}

	// Get updated order
	orders, err := s.Orders(ctx, &domain.GetOrdersRequest{
		Ids:   []uuid.UUID{req.Id},
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return nil, domain.ErrOrderNotFound
	}

	return orders[0], nil
}

func (s *orderStorage) Orders(ctx context.Context, req *domain.GetOrdersRequest) ([]*domain.Order, error) {
	s.cache.DeleteExpired()
	req.Validate()

	if cacheOrders := s.cache.Get(req.CacheKey()); cacheOrders != nil {
		return cacheOrders.Value(), nil
	}

	// Query orders
	query := s.psql.Select("id", "user_id", "status", "created_at", "updated_at").
		From("orders")

	if len(req.Ids) > 0 {
		query = query.Where(sq.Eq{"id": req.Ids})
	}

	if len(req.UserIds) > 0 {
		query = query.Where(sq.Eq{"user_id": req.UserIds})
	}

	if len(req.Statuses) > 0 {
		query = query.Where(sq.Eq{"status": req.Statuses})
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

	var orders []*domain.Order
	var orderIds []uuid.UUID

	for rows.Next() {
		var dto orderDto
		err := rows.Scan(&dto.Id, &dto.UserId, &dto.Status, &dto.CreatedAt, &dto.UpdatedAt)
		if err != nil {
			return nil, err
		}

		order, err := dto.toDomain()
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
		orderIds = append(orderIds, order.Id)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Load order items if we have orders
	if len(orders) > 0 {
		err = s.loadOrderItems(ctx, orders, orderIds)
		if err != nil {
			return nil, err
		}
	}

	s.cache.Set(req.CacheKey(), orders, ttlcache.DefaultTTL)

	return orders, nil
}

func (s *orderStorage) CountOrders(ctx context.Context, req *domain.GetOrdersRequest) (int, error) {
	req.Validate()

	query := s.psql.Select("COUNT(*)").
		From("orders")

	if len(req.Ids) > 0 {
		query = query.Where(sq.Eq{"id": req.Ids})
	}

	if len(req.UserIds) > 0 {
		query = query.Where(sq.Eq{"user_id": req.UserIds})
	}

	if len(req.Statuses) > 0 {
		query = query.Where(sq.Eq{"status": req.Statuses})
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

func (s *orderStorage) loadOrderItems(ctx context.Context, orders []*domain.Order, orderIds []uuid.UUID) error {
	// Query all items for these orders
	itemQuery := s.psql.Select("id", "order_id", "product_id", "quantity", "product_snapshot", "created_at").
		From("order_items").
		Where(sq.Eq{"order_id": orderIds}).
		OrderBy("created_at")

	sql, args, err := itemQuery.ToSql()
	if err != nil {
		return err
	}

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Group items by order ID
	itemsByOrderId := make(map[uuid.UUID][]*domain.OrderItem)

	for rows.Next() {
		var dto orderItemDto
		err := rows.Scan(&dto.Id, &dto.OrderId, &dto.ProductId, &dto.Quantity, &dto.ProductSnapshot, &dto.CreatedAt)
		if err != nil {
			return err
		}

		item, err := dto.toDomain()
		if err != nil {
			return err
		}

		itemsByOrderId[dto.OrderId] = append(itemsByOrderId[dto.OrderId], item)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	// Assign items to orders
	for _, order := range orders {
		order.Items = itemsByOrderId[order.Id]
	}

	return nil
}
