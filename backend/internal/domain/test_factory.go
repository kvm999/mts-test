package domain

import (
	"time"

	"github.com/google/uuid"
)

type Factory struct{}

func (f *Factory) User() *User {
	user := &User{
		Id:        uuid.New(),
		FirstName: "John",
		LastName:  "Doe",
		Age:       25,
		IsMarried: false,
		CreatedAt: time.Now(),
	}

	_ = user.SetPassword("password123")
	return user
}

func (f *Factory) UserWithAge(age int) *User {
	user := f.User()
	user.Age = age
	return user
}

func (f *Factory) Product() *Product {
	return &Product{
		Id:          uuid.New(),
		Description: "Test Product",
		Tags:        []string{"tag1", "tag2"},
		Quantity:    100,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (f *Factory) ProductWithQuantity(quantity int) *Product {
	product := f.Product()
	product.Quantity = quantity
	return product
}

func (f *Factory) ProductWithDescription(description string) *Product {
	product := f.Product()
	product.Description = description
	return product
}

func (f *Factory) Order(userId uuid.UUID, productIds ...uuid.UUID) *Order {
	items := make([]*OrderItem, 0, len(productIds))

	for _, productId := range productIds {
		item := &OrderItem{
			Id:        uuid.New(),
			ProductId: productId,
			Quantity:  1,
			ProductSnapshot: ProductSnapshot{
				Description: "Test Product",
				Tags:        []string{"tag1"},
			},
			CreatedAt: time.Now(),
		}
		items = append(items, item)
	}

	return &Order{
		Id:        uuid.New(),
		UserId:    userId,
		Status:    OrderStatusPending,
		Items:     items,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (f *Factory) OrderItem(orderId, productId uuid.UUID, quantity int) *OrderItem {
	return &OrderItem{
		Id:        uuid.New(),
		OrderId:   orderId,
		ProductId: productId,
		Quantity:  quantity,
		ProductSnapshot: ProductSnapshot{
			Description: "Test Product",
			Tags:        []string{"tag1"},
		},
		CreatedAt: time.Now(),
	}
}

func (f *Factory) CreateUserRequest() *CreateUserRequest {
	return &CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Age:       25,
		IsMarried: false,
		Password:  "password123",
	}
}

func (f *Factory) CreateProductRequest() *CreateProductRequest {
	return &CreateProductRequest{
		Description: "Test Product",
		Tags:        []string{"tag1", "tag2"},
		Quantity:    100,
	}
}

func (f *Factory) CreateOrderRequest(userId uuid.UUID, productIds ...uuid.UUID) *CreateOrderRequest {
	items := make([]CreateOrderItemRequest, 0, len(productIds))

	for _, productId := range productIds {
		items = append(items, CreateOrderItemRequest{
			ProductId: productId,
			Quantity:  1,
		})
	}

	return &CreateOrderRequest{
		UserId: userId,
		Items:  items,
	}
}
