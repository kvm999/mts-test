package rest

import (
	"time"

	"github.com/google/uuid"

	"mts/internal/domain"
)

// User represents a user in the API
// @Description User information
type User struct {
	// User ID
	// @Description Unique identifier for the user
	// @Example 123e4567-e89b-12d3-a456-426614174000
	Id uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000" swaggertype:"string"`

	// First name
	// @Description User's first name
	// @Example John
	FirstName string `json:"first_name" example:"John"`

	// Last name
	// @Description User's last name
	// @Example Doe
	LastName string `json:"last_name" example:"Doe"`

	// Full name
	// @Description User's full name (first + last)
	// @Example John Doe
	FullName string `json:"full_name" example:"John Doe"`

	// Age
	// @Description User's age (must be 18 or older)
	// @Example 25
	Age int `json:"age" example:"25"`

	// Is married
	// @Description Whether the user is married
	// @Example false
	IsMarried bool `json:"is_married" example:"false"`

	// Created at
	// @Description When the user was created
	// @Example 2024-01-15T10:30:00Z
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
} // @name User

// CreateUserRequest represents request to create a new user
// @Description Request payload for user registration
type CreateUserRequest struct {
	// First name
	// @Description User's first name (required)
	// @Example John
	FirstName string `json:"first_name" binding:"required" validate:"required" example:"John"`

	// Last name
	// @Description User's last name (required)
	// @Example Doe
	LastName string `json:"last_name" binding:"required" validate:"required" example:"Doe"`

	// Age
	// @Description User's age (must be 18 or older)
	// @Example 25
	Age int `json:"age" binding:"required" validate:"required,gte=18" example:"25"`

	// Is married
	// @Description Whether the user is married
	// @Example false
	IsMarried bool `json:"is_married" example:"false"`

	// Password
	// @Description User's password (minimum 8 characters)
	// @Example password123
	Password string `json:"password" binding:"required" validate:"required,min=8" example:"password123"`
} // @name CreateUserRequest

func (req *CreateUserRequest) ToDomain() *domain.CreateUserRequest {
	return &domain.CreateUserRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		IsMarried: req.IsMarried,
		Password:  req.Password,
	}
}

// UsersResponse represents paginated list of users
// @Description Paginated response containing list of users
type UsersResponse struct {
	// Users
	// @Description List of users
	Users []*User `json:"users"`

	// Pagination
	// @Description Pagination information
	Pagination *Pagination `json:"pagination"`
} // @name UsersResponse

func NewUser(domainUser *domain.User) *User {
	return &User{
		Id:        domainUser.Id,
		FirstName: domainUser.FirstName,
		LastName:  domainUser.LastName,
		FullName:  domainUser.FullName(),
		Age:       domainUser.Age,
		IsMarried: domainUser.IsMarried,
		CreatedAt: domainUser.CreatedAt,
	}
}

func NewUsersResponse(domainUsers []*domain.User, pagination Pagination) *UsersResponse {
	users := make([]*User, 0, len(domainUsers))
	for _, domainUser := range domainUsers {
		users = append(users, NewUser(domainUser))
	}

	return &UsersResponse{
		Users:      users,
		Pagination: &pagination,
	}
}
