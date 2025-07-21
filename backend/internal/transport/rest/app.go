// Package rest provides REST API transport layer
// @title MTS API
// @version 1.0
// @description MTS test assignment API for managing users, products and orders
// @contact.name API Support
//
// @host localhost:8080
// @BasePath /
//
// @tag.name Users
// @tag.description User registration and management
//
// @tag.name Products
// @tag.description Product catalog management
//
// @tag.name Orders
// @tag.description Order management with stock control
package rest

import (
	"shared"

	"github.com/Flussen/swagger-fiber-v3"
	"github.com/gofiber/fiber/v3"

	_ "mts/internal/transport/rest/docs"

	"mts/internal/domain"
)

func New(
	userAppService domain.UserAppService,
	productAppService domain.ProductAppService,
	orderAppService domain.OrderAppService,
) *fiber.App {
	app := fiber.New()

	// Используем shared логер и middleware
	app.Use(func(c fiber.Ctx) error {
		shared.Logger.Info().
			Str("method", c.Method()).
			Str("path", c.Path()).
			Msg("request received")
		return c.Next()
	})

	app.Get("/docs/*", swagger.HandlerDefault)

	v1 := app.Group("/api/v1")

	// Users routes
	user := newUserHandler(userAppService)
	v1.Group("/users").
		Post("", user.registerUser).
		Get("", user.getUsers).
		Get(":user_id", user.getUser)

	// Products routes
	product := newProductHandler(productAppService)
	v1.Group("/products").
		Post("", product.createProduct).
		Get("", product.getProducts).
		Get(":product_id", product.getProduct).
		Put(":product_id", product.updateProduct)

	// Orders routes
	order := newOrderHandler(orderAppService)
	v1.Group("/orders").
		Post("", order.createOrder).
		Get("", order.getOrders).
		Get(":order_id", order.getOrder).
		Put(":order_id", order.updateOrder).
		Post(":order_id/cancel", order.cancelOrder)

	return app
}
