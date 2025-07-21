package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"mts/internal/domain"
)

func NewOrderAppService(
	orderStorage domain.OrderStorage,
	productStorage domain.ProductStorage,
	userStorage domain.UserStorage,
) domain.OrderAppService {
	return &orderAppService{
		orderStorage:   orderStorage,
		productStorage: productStorage,
		userStorage:    userStorage,
	}
}

type orderAppService struct {
	orderStorage   domain.OrderStorage
	productStorage domain.ProductStorage
	userStorage    domain.UserStorage
}

func (s *orderAppService) CreateOrder(ctx context.Context, req *domain.CreateOrderRequest) (*domain.Order, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "CreateOrder").
		Str("user_id", req.UserId.String()).
		Int("items_count", len(req.Items)).
		Logger()

	logger.Info().Msg("creating new order")

	if err := req.Validate(); err != nil {
		logger.Error().Err(err).Msg("order request validation failed")
		return nil, err
	}

	// Check if user exists
	users, err := s.userStorage.Users(ctx, &domain.GetUsersRequest{
		Ids:   []uuid.UUID{req.UserId},
		Limit: 1,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch user")
		return nil, err
	}
	if len(users) == 0 {
		logger.Error().Msg("user not found")
		return nil, domain.ErrUserNotFound
	}

	// Get all products from request
	productIds := make([]uuid.UUID, 0, len(req.Items))
	requestedQuantities := make(map[uuid.UUID]int)

	for _, item := range req.Items {
		productIds = append(productIds, item.ProductId)
		requestedQuantities[item.ProductId] += item.Quantity
	}

	logger.Info().
		Int("unique_products", len(productIds)).
		Msg("fetching products for order")

	products, err := s.productStorage.Products(ctx, &domain.GetProductsRequest{
		Ids: productIds,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch products")
		return nil, err
	}

	// Check if all products exist and have sufficient quantity
	productMap := make(map[uuid.UUID]*domain.Product)
	for _, product := range products {
		productMap[product.Id] = product
	}

	for productId, requestedQty := range requestedQuantities {
		product, exists := productMap[productId]
		if !exists {
			logger.Error().
				Str("product_id", productId.String()).
				Msg("product not found")
			return nil, fmt.Errorf("%w: product %s not found", domain.ErrProductNotFound, productId)
		}

		if product.Quantity < requestedQty {
			logger.Error().
				Str("product_id", productId.String()).
				Int("available", product.Quantity).
				Int("requested", requestedQty).
				Msg("insufficient stock")
			return nil, fmt.Errorf("%w: product %s has only %d items but %d requested",
				domain.ErrInsufficientStock, productId, product.Quantity, requestedQty)
		}
	}

	logger.Info().Msg("reserving product quantities")

	// Reserve products (decrease quantities)
	for productId, requestedQty := range requestedQuantities {
		product := productMap[productId]
		err = product.ReserveQuantity(requestedQty)
		if err != nil {
			logger.Error().
				Err(err).
				Str("product_id", productId.String()).
				Msg("failed to reserve product quantity")
			return nil, err
		}

		// Update product in storage
		_, err = s.productStorage.UpdateProduct(ctx, &domain.UpdateProductRequest{
			Id:       product.Id,
			Quantity: &product.Quantity,
		})
		if err != nil {
			logger.Error().
				Err(err).
				Str("product_id", productId.String()).
				Msg("failed to update product quantity in storage")
			return nil, err
		}
	}

	// Create order with historical product snapshots
	order := &domain.Order{
		UserId: req.UserId,
		Status: domain.OrderStatusPending,
	}

	// Create order items with product snapshots
	for _, itemReq := range req.Items {
		product := productMap[itemReq.ProductId]

		item := &domain.OrderItem{
			ProductId: itemReq.ProductId,
			Quantity:  itemReq.Quantity,
			ProductSnapshot: domain.ProductSnapshot{
				Description: product.Description,
				Tags:        product.Tags,
			},
		}

		order.Items = append(order.Items, item)
	}

	err = s.orderStorage.CreateOrder(ctx, order)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create order in storage")
		return nil, err
	}

	logger.Info().
		Str("order_id", order.Id.String()).
		Msg("order created successfully")

	return order, nil
}

func (s *orderAppService) UpdateOrder(ctx context.Context, req *domain.UpdateOrderRequest) (*domain.Order, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "UpdateOrder").
		Str("order_id", req.Id.String()).
		Logger()

	logger.Info().Msg("updating order")

	order, err := s.orderStorage.UpdateOrder(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update order in storage")
		return nil, err
	}

	logger.Info().Msg("order updated successfully")

	return order, nil
}

func (s *orderAppService) Orders(ctx context.Context, req *domain.GetOrdersRequest) ([]*domain.Order, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "Orders").
		Int("ids_count", len(req.Ids)).
		Logger()

	logger.Info().Msg("fetching orders")

	orders, err := s.orderStorage.Orders(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch orders from storage")
		return nil, err
	}

	logger.Info().
		Int("orders_count", len(orders)).
		Msg("orders fetched successfully")

	return orders, nil
}

func (s *orderAppService) CancelOrder(ctx context.Context, orderId uuid.UUID) (*domain.Order, error) {
	// Get the order
	orders, err := s.orderStorage.Orders(ctx, &domain.GetOrdersRequest{
		Ids:   []uuid.UUID{orderId},
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, domain.ErrOrderNotFound
	}

	order := orders[0]

	// Check if order can be cancelled
	if !order.CanBeCancelled() {
		return nil, fmt.Errorf("%w: order in status %s cannot be cancelled", domain.ErrOrderValidation, order.Status)
	}

	// Restore product quantities
	productQuantityToRestore := make(map[uuid.UUID]int)
	for _, item := range order.Items {
		productQuantityToRestore[item.ProductId] += item.Quantity
	}

	for productId, quantityToRestore := range productQuantityToRestore {
		products, err := s.productStorage.Products(ctx, &domain.GetProductsRequest{
			Ids:   []uuid.UUID{productId},
			Limit: 1,
		})
		if err != nil {
			return nil, err
		}
		if len(products) == 0 {
			continue // Product might have been deleted
		}

		product := products[0]
		err = product.RestoreQuantity(quantityToRestore)
		if err != nil {
			return nil, err
		}

		_, err = s.productStorage.UpdateProduct(ctx, &domain.UpdateProductRequest{
			Id:       product.Id,
			Quantity: &product.Quantity,
		})
		if err != nil {
			return nil, err
		}
	}

	// Cancel the order
	err = order.Cancel()
	if err != nil {
		return nil, err
	}

	return s.orderStorage.UpdateOrder(ctx, &domain.UpdateOrderRequest{
		Id:     order.Id,
		Status: order.Status,
	})
}
