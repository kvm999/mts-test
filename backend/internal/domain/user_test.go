package domain

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name        string
		setupUser   func() *User
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid user",
			setupUser: func() *User {
				user := &User{
					FirstName: "John",
					LastName:  "Doe",
					Age:       25,
					IsMarried: false,
				}
				err := user.SetPassword("password123")
				require.NoError(t, err)
				return user
			},
			wantErr: false,
		},
		{
			name: "user with minimum age",
			setupUser: func() *User {
				user := &User{
					FirstName: "Jane",
					LastName:  "Smith",
					Age:       18, // minimum valid age
					IsMarried: true,
				}
				err := user.SetPassword("password123")
				require.NoError(t, err)
				return user
			},
			wantErr: false,
		},
		{
			name: "user too young",
			setupUser: func() *User {
				user := &User{
					FirstName: "Young",
					LastName:  "Person",
					Age:       17, // too young
					IsMarried: false,
				}
				err := user.SetPassword("password123")
				require.NoError(t, err)
				return user
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "empty first name",
			setupUser: func() *User {
				user := &User{
					FirstName: "", // empty
					LastName:  "Doe",
					Age:       25,
					IsMarried: false,
				}
				err := user.SetPassword("password123")
				require.NoError(t, err)
				return user
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "whitespace-only first name",
			setupUser: func() *User {
				user := &User{
					FirstName: "   ", // whitespace only
					LastName:  "Doe",
					Age:       25,
					IsMarried: false,
				}
				err := user.SetPassword("password123")
				require.NoError(t, err)
				return user
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "empty last name",
			setupUser: func() *User {
				user := &User{
					FirstName: "John",
					LastName:  "", // empty
					Age:       25,
					IsMarried: false,
				}
				err := user.SetPassword("password123")
				require.NoError(t, err)
				return user
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "no password hash",
			setupUser: func() *User {
				return &User{
					FirstName:    "John",
					LastName:     "Doe",
					Age:          25,
					IsMarried:    false,
					PasswordHash: nil, // no password hash
					Salt:         []byte("salt"),
				}
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "no salt",
			setupUser: func() *User {
				return &User{
					FirstName:    "John",
					LastName:     "Doe",
					Age:          25,
					IsMarried:    false,
					PasswordHash: []byte("hash"),
					Salt:         nil, // no salt
				}
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setupUser()
			originalId := user.Id
			originalCreatedAt := user.CreatedAt

			err := user.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)

				// Check that ID is generated if nil
				if originalId == uuid.Nil {
					assert.NotEqual(t, uuid.Nil, user.Id)
				}

				// Check that CreatedAt is set if zero
				if originalCreatedAt.IsZero() {
					assert.False(t, user.CreatedAt.IsZero())
				}
			}
		})
	}
}

func TestUser_SetPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "minimum length password",
			password: "12345678", // exactly 8 characters
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "this_is_a_very_long_password_with_many_characters",
			wantErr:  false,
		},
		{
			name:        "too short password",
			password:    "1234567", // 7 characters
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name:        "empty password",
			password:    "",
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				FirstName: "John",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
			}

			err := user.SetPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
				assert.Empty(t, user.PasswordHash)
				assert.Empty(t, user.Salt)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEmpty(t, user.Salt)
				assert.Len(t, user.Salt, 16) // UUID is 16 bytes
			}
		})
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	password := "testPassword123"
	user := &User{
		FirstName: "John",
		LastName:  "Doe",
		Age:       25,
		IsMarried: false,
	}

	err := user.SetPassword(password)
	require.NoError(t, err)

	tests := []struct {
		name           string
		password       string
		expectedResult bool
	}{
		{
			name:           "correct password",
			password:       password,
			expectedResult: true,
		},
		{
			name:           "incorrect password",
			password:       "wrongPassword",
			expectedResult: false,
		},
		{
			name:           "empty password",
			password:       "",
			expectedResult: false,
		},
		{
			name:           "case sensitive password",
			password:       "testpassword123",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := user.VerifyPassword(tt.password)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestUser_FullName(t *testing.T) {
	tests := []struct {
		name         string
		firstName    string
		lastName     string
		expectedName string
	}{
		{
			name:         "normal names",
			firstName:    "John",
			lastName:     "Doe",
			expectedName: "John Doe",
		},
		{
			name:         "names with extra spaces",
			firstName:    "  Jane  ",
			lastName:     "  Smith  ",
			expectedName: "Jane     Smith", // FullName doesn't trim spaces, just concatenates
		},
		{
			name:         "single names",
			firstName:    "Madonna",
			lastName:     "",
			expectedName: "Madonna",
		},
		{
			name:         "empty first name",
			firstName:    "",
			lastName:     "Doe",
			expectedName: "Doe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				FirstName: tt.firstName,
				LastName:  tt.lastName,
			}

			fullName := user.FullName()
			assert.Equal(t, tt.expectedName, fullName)
		})
	}
}

func TestCreateUserRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		request     *CreateUserRequest
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid request",
			request: &CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
				Password:  "password123",
			},
			wantErr: false,
		},
		{
			name: "minimum valid age and password",
			request: &CreateUserRequest{
				FirstName: "Jane",
				LastName:  "Smith",
				Age:       18,
				IsMarried: true,
				Password:  "12345678",
			},
			wantErr: false,
		},
		{
			name: "empty first name",
			request: &CreateUserRequest{
				FirstName: "",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
				Password:  "password123",
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "whitespace-only first name",
			request: &CreateUserRequest{
				FirstName: "   ",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
				Password:  "password123",
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "empty last name",
			request: &CreateUserRequest{
				FirstName: "John",
				LastName:  "",
				Age:       25,
				IsMarried: false,
				Password:  "password123",
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "age too young",
			request: &CreateUserRequest{
				FirstName: "Young",
				LastName:  "Person",
				Age:       17,
				IsMarried: false,
				Password:  "password123",
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
		{
			name: "password too short",
			request: &CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
				Password:  "1234567", // 7 characters
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateUserRequest_ToDomain(t *testing.T) {
	tests := []struct {
		name        string
		request     *CreateUserRequest
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid request conversion",
			request: &CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
				Password:  "password123",
			},
			wantErr: false,
		},
		{
			name: "request with spaces in names",
			request: &CreateUserRequest{
				FirstName: "  Jane  ",
				LastName:  "  Smith  ",
				Age:       30,
				IsMarried: true,
				Password:  "securePassword456",
			},
			wantErr: false,
		},
		{
			name: "invalid request - fails validation",
			request: &CreateUserRequest{
				FirstName: "",
				LastName:  "Doe",
				Age:       25,
				IsMarried: false,
				Password:  "password123",
			},
			wantErr:     true,
			expectedErr: ErrUserValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := tt.request.ToDomain()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)

				// Verify user fields are set correctly
				assert.Equal(t, strings.TrimSpace(tt.request.FirstName), user.FirstName)
				assert.Equal(t, strings.TrimSpace(tt.request.LastName), user.LastName)
				assert.Equal(t, tt.request.Age, user.Age)
				assert.Equal(t, tt.request.IsMarried, user.IsMarried)

				// Verify password was set correctly
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEmpty(t, user.Salt)
				assert.True(t, user.VerifyPassword(tt.request.Password))
			}
		})
	}
}

func TestGetUsersRequest_Validate(t *testing.T) {
	tests := []struct {
		name            string
		request         *GetUsersRequest
		expectedRequest *GetUsersRequest
	}{
		{
			name: "valid request unchanged",
			request: &GetUsersRequest{
				Limit:  20,
				Offset: 10,
			},
			expectedRequest: &GetUsersRequest{
				Limit:  20,
				Offset: 10,
			},
		},
		{
			name: "zero limit defaults to 10",
			request: &GetUsersRequest{
				Limit:  0,
				Offset: 5,
			},
			expectedRequest: &GetUsersRequest{
				Limit:  10,
				Offset: 5,
			},
		},
		{
			name: "limit > 100 capped to 100",
			request: &GetUsersRequest{
				Limit:  150,
				Offset: 0,
			},
			expectedRequest: &GetUsersRequest{
				Limit:  100,
				Offset: 0,
			},
		},
		{
			name: "negative offset becomes 0",
			request: &GetUsersRequest{
				Limit:  10,
				Offset: -5,
			},
			expectedRequest: &GetUsersRequest{
				Limit:  10,
				Offset: 0,
			},
		},
		{
			name: "negative limit defaults to 10",
			request: &GetUsersRequest{
				Limit:  -10,
				Offset: 0,
			},
			expectedRequest: &GetUsersRequest{
				Limit:  10,
				Offset: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalIds := tt.request.Ids // Preserve IDs

			tt.request.Validate()

			assert.Equal(t, tt.expectedRequest.Limit, tt.request.Limit)
			assert.Equal(t, tt.expectedRequest.Offset, tt.request.Offset)
			assert.Equal(t, originalIds, tt.request.Ids) // IDs should be unchanged
		})
	}
}

func TestGetUsersRequest_CacheKey(t *testing.T) {
	userId1 := uuid.New()
	userId2 := uuid.New()

	tests := []struct {
		name        string
		request1    *GetUsersRequest
		request2    *GetUsersRequest
		shouldEqual bool
	}{
		{
			name: "identical requests have same cache key",
			request1: &GetUsersRequest{
				Ids:    []uuid.UUID{userId1},
				Limit:  10,
				Offset: 0,
			},
			request2: &GetUsersRequest{
				Ids:    []uuid.UUID{userId1},
				Limit:  10,
				Offset: 0,
			},
			shouldEqual: true,
		},
		{
			name: "different IDs have different cache keys",
			request1: &GetUsersRequest{
				Ids:    []uuid.UUID{userId1},
				Limit:  10,
				Offset: 0,
			},
			request2: &GetUsersRequest{
				Ids:    []uuid.UUID{userId2},
				Limit:  10,
				Offset: 0,
			},
			shouldEqual: false,
		},
		{
			name: "different limits have different cache keys",
			request1: &GetUsersRequest{
				Limit:  10,
				Offset: 0,
			},
			request2: &GetUsersRequest{
				Limit:  20,
				Offset: 0,
			},
			shouldEqual: false,
		},
		{
			name: "different offsets have different cache keys",
			request1: &GetUsersRequest{
				Limit:  10,
				Offset: 0,
			},
			request2: &GetUsersRequest{
				Limit:  10,
				Offset: 10,
			},
			shouldEqual: false,
		},
		{
			name: "empty requests have same cache key",
			request1: &GetUsersRequest{
				Limit:  10,
				Offset: 0,
			},
			request2: &GetUsersRequest{
				Limit:  10,
				Offset: 0,
			},
			shouldEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate both requests first to normalize them
			tt.request1.Validate()
			tt.request2.Validate()

			key1 := tt.request1.CacheKey()
			key2 := tt.request2.CacheKey()

			if tt.shouldEqual {
				assert.Equal(t, key1, key2)
			} else {
				assert.NotEqual(t, key1, key2)
			}
		})
	}
}

func TestUser_PasswordSecurity(t *testing.T) {
	t.Run("salt is unique for each password setting", func(t *testing.T) {
		user1 := &User{FirstName: "User1", LastName: "Test", Age: 25}
		user2 := &User{FirstName: "User2", LastName: "Test", Age: 25}

		password := "samePassword123"

		err1 := user1.SetPassword(password)
		err2 := user2.SetPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)

		// Same password should produce different hashes due to different salts
		assert.NotEqual(t, user1.Salt, user2.Salt)
		assert.NotEqual(t, user1.PasswordHash, user2.PasswordHash)

		// But both should verify correctly
		assert.True(t, user1.VerifyPassword(password))
		assert.True(t, user2.VerifyPassword(password))
	})

	t.Run("multiple password sets on same user change salt and hash", func(t *testing.T) {
		user := &User{FirstName: "John", LastName: "Doe", Age: 25}

		password1 := "firstPassword123"
		password2 := "secondPassword456"

		err := user.SetPassword(password1)
		require.NoError(t, err)

		originalSalt := make([]byte, len(user.Salt))
		copy(originalSalt, user.Salt)
		originalHash := make([]byte, len(user.PasswordHash))
		copy(originalHash, user.PasswordHash)

		// Set new password
		err = user.SetPassword(password2)
		require.NoError(t, err)

		// Salt and hash should change
		assert.NotEqual(t, originalSalt, user.Salt)
		assert.NotEqual(t, originalHash, user.PasswordHash)

		// Old password should not work, new password should work
		assert.False(t, user.VerifyPassword(password1))
		assert.True(t, user.VerifyPassword(password2))
	})
}
