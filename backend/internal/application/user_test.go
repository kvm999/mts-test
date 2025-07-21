package application

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"mts/internal/domain"
)

// Mock для UserStorage
type mockUserStorage struct {
	mock.Mock
}

func (m *mockUserStorage) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserStorage) Users(ctx context.Context, req *domain.GetUsersRequest) ([]*domain.User, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *mockUserStorage) CountUsers(ctx context.Context, req *domain.GetUsersRequest) (int, error) {
	args := m.Called(ctx, req)
	return args.Int(0), args.Error(1)
}

func TestUserAppService_RegisterUser(t *testing.T) {
	tests := []struct {
		name        string
		request     *domain.CreateUserRequest
		setupMock   func(*mockUserStorage)
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successful registration",
			request: &domain.CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
				Password:  "password123",
			},
			setupMock: func(m *mockUserStorage) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid age - too young",
			request: &domain.CreateUserRequest{
				FirstName: "Jane",
				LastName:  "Doe",
				Age:       17, // Invalid: too young
				IsMarried: false,
				Password:  "password123",
			},
			setupMock: func(m *mockUserStorage) {
				// Mock не вызывается, так как валидация не пройдет
			},
			wantErr:     true,
			expectedErr: domain.ErrUserValidation,
		},
		{
			name: "invalid password - too short",
			request: &domain.CreateUserRequest{
				FirstName: "Bob",
				LastName:  "Smith",
				Age:       30,
				IsMarried: true,
				Password:  "123", // Invalid: too short
			},
			setupMock: func(m *mockUserStorage) {
				// Mock не вызывается, так как валидация не пройдет
			},
			wantErr:     true,
			expectedErr: domain.ErrUserValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockStorage := new(mockUserStorage)
			tt.setupMock(mockStorage)

			userAppService := NewUserAppService(mockStorage)
			ctx := context.Background()

			// Act
			user, err := userAppService.RegisterUser(ctx, tt.request)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.request.FirstName, user.FirstName)
				assert.Equal(t, tt.request.LastName, user.LastName)
				assert.Equal(t, tt.request.Age, user.Age)
				assert.Equal(t, tt.request.IsMarried, user.IsMarried)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEmpty(t, user.Salt)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}
