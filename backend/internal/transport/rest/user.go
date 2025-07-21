package rest

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"mts/internal/domain"
)

type userHandler struct {
	userAppService domain.UserAppService
}

func newUserHandler(userAppService domain.UserAppService) *userHandler {
	return &userHandler{
		userAppService: userAppService,
	}
}

// registerUser registers a new user in the system
// @Summary Register new user
// @Description Register a new user with validation (age >= 18, password >= 8 chars)
// @Tags Users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User registration data"
// @Success 201 {object} User "User registered successfully"
// @Failure 400 {object} ErrorResponse "Bad request - validation failed"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/users [post]
func (h *userHandler) registerUser(c fiber.Ctx) error {
	var req CreateUserRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := h.userAppService.RegisterUser(c.Context(), req.ToDomain())
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, domain.ErrUserValidation) {
			status = fiber.StatusBadRequest
		}
		return fiber.NewError(status, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(NewUser(user))
}

// getUsers retrieves a paginated list of users
// @Summary Get users list
// @Description Retrieve a paginated list of all users in the system
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination" default(1) minimum(1)
// @Param size query int false "Number of items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} UsersResponse "Users retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid pagination parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/users [get]
func (h *userHandler) getUsers(c fiber.Ctx) error {
	pagination := NewPaginationFromRequest(c)

	users, err := h.userAppService.Users(c.Context(), &domain.GetUsersRequest{
		Limit:  pagination.Limit(),
		Offset: pagination.Offset(),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	count, err := h.userAppService.CountUsers(c.Context(), &domain.GetUsersRequest{})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	pagination.Total = count
	pagination.CalculateTotalPages()

	return c.JSON(NewUsersResponse(users, *pagination))
}

// getUser retrieves a specific user by ID
// @Summary Get user by ID
// @Description Retrieve detailed information about a specific user using their unique identifier
// @Tags Users
// @Accept json
// @Produce json
// @Param user_id path string true "User unique identifier" format(uuid)
// @Success 200 {object} User "User information retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad request - invalid user ID format"
// @Failure 404 {object} ErrorResponse "Not found - user with specified ID does not exist"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/users/{user_id} [get]
func (h *userHandler) getUser(c fiber.Ctx) error {
	userId, err := uuid.Parse(c.Params("user_id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user ID format")
	}

	users, err := h.userAppService.Users(c.Context(), &domain.GetUsersRequest{
		Ids:   []uuid.UUID{userId},
		Limit: 1,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	if len(users) == 0 {
		return fiber.NewError(fiber.StatusNotFound, domain.ErrUserNotFound.Error())
	}

	return c.JSON(NewUser(users[0]))
}
