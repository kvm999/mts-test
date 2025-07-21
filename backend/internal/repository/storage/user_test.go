package storage

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"mts/internal/domain"
	"shared"
)

type UserStorageSuite struct {
	shared.Suite[any]
	storage domain.UserStorage
}

func (s *UserStorageSuite) SetupSuite() {
	s.PostgresEnabled = true
	s.Suite.SetupSuite()
	s.storage = NewUserStorage(s.PostgresConn)
}

func (s *UserStorageSuite) TearDownTest() {
	// Clear all users after each test for isolation
	_, err := s.PostgresConn.Exec(s.Ctx, "TRUNCATE TABLE users RESTART IDENTITY CASCADE")
	s.Require().NoError(err)
}

func (s *UserStorageSuite) TestCreateUser_Success() {
	user := &domain.User{
		FirstName: "John",
		LastName:  "Doe",
		Age:       25,
		IsMarried: false,
	}
	err := user.SetPassword("password123")
	s.Require().NoError(err)

	err = s.storage.CreateUser(s.Ctx, user)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, user.Id)
	s.False(user.CreatedAt.IsZero())
	s.NotEmpty(user.PasswordHash)
	s.NotEmpty(user.Salt)
}

func (s *UserStorageSuite) TestCreateUser_MinimalValidFields() {
	user := &domain.User{
		FirstName: "Jane",
		LastName:  "Smith",
		Age:       18, // minimum age
		IsMarried: true,
	}
	err := user.SetPassword("12345678") // minimum password
	s.Require().NoError(err)

	err = s.storage.CreateUser(s.Ctx, user)
	s.Require().NoError(err)
	s.NotEqual(uuid.Nil, user.Id)
	s.False(user.CreatedAt.IsZero())
}

func (s *UserStorageSuite) TestCreateUser_InvalidAge() {
	user := &domain.User{
		FirstName: "Bob",
		LastName:  "Young",
		Age:       17, // Invalid: too young
		IsMarried: false,
	}
	err := user.SetPassword("password123")
	s.Require().NoError(err)

	err = s.storage.CreateUser(s.Ctx, user)
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserValidation)
}

func (s *UserStorageSuite) TestCreateUser_EmptyFirstName() {
	user := &domain.User{
		FirstName: "", // Invalid: empty
		LastName:  "Test",
		Age:       25,
		IsMarried: false,
	}
	err := user.SetPassword("password123")
	s.Require().NoError(err)

	err = s.storage.CreateUser(s.Ctx, user)
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserValidation)
}

func (s *UserStorageSuite) TestCreateUser_EmptyLastName() {
	user := &domain.User{
		FirstName: "Test",
		LastName:  "", // Invalid: empty
		Age:       25,
		IsMarried: false,
	}
	err := user.SetPassword("password123")
	s.Require().NoError(err)

	err = s.storage.CreateUser(s.Ctx, user)
	s.Error(err)
	s.ErrorIs(err, domain.ErrUserValidation)
}

func (s *UserStorageSuite) TestUsers_GetAll() {
	// Create test users
	var testUsers []*domain.User
	for i := 0; i < 5; i++ {
		user := &domain.User{
			FirstName: "User",
			LastName:  "Test",
			Age:       20 + i,
			IsMarried: i%2 == 0,
		}
		err := user.SetPassword("password123")
		s.Require().NoError(err)

		err = s.storage.CreateUser(s.Ctx, user)
		s.Require().NoError(err)

		testUsers = append(testUsers, user)
	}

	req := &domain.GetUsersRequest{
		Limit:  10,
		Offset: 0,
	}
	users, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Len(users, 5)

	// Should be ordered by created_at DESC
	for i := 0; i < len(users)-1; i++ {
		s.True(users[i].CreatedAt.After(users[i+1].CreatedAt) ||
			users[i].CreatedAt.Equal(users[i+1].CreatedAt))
	}

	// Validate user fields
	for _, user := range users {
		s.NotEqual(uuid.Nil, user.Id)
		s.NotEmpty(user.FirstName)
		s.NotEmpty(user.LastName)
		s.GreaterOrEqual(user.Age, 18)
		s.NotEmpty(user.PasswordHash)
		s.NotEmpty(user.Salt)
		s.False(user.CreatedAt.IsZero())
	}
}

func (s *UserStorageSuite) TestUsers_WithLimit() {
	// Create test users
	for i := 0; i < 5; i++ {
		user := &domain.User{
			FirstName: "User",
			LastName:  "Test",
			Age:       20 + i,
			IsMarried: i%2 == 0,
		}
		err := user.SetPassword("password123")
		s.Require().NoError(err)

		err = s.storage.CreateUser(s.Ctx, user)
		s.Require().NoError(err)
	}

	req := &domain.GetUsersRequest{
		Limit:  2,
		Offset: 0,
	}
	users, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Len(users, 2)
}

func (s *UserStorageSuite) TestUsers_WithOffset() {
	// Create test users
	for i := 0; i < 5; i++ {
		user := &domain.User{
			FirstName: "User",
			LastName:  "Test",
			Age:       20 + i,
			IsMarried: i%2 == 0,
		}
		err := user.SetPassword("password123")
		s.Require().NoError(err)

		err = s.storage.CreateUser(s.Ctx, user)
		s.Require().NoError(err)
	}

	req := &domain.GetUsersRequest{
		Limit:  10,
		Offset: 2,
	}
	users, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Len(users, 3)
}

func (s *UserStorageSuite) TestUsers_ByIds() {
	// Create test users
	var testUsers []*domain.User
	for i := 0; i < 3; i++ {
		user := &domain.User{
			FirstName: "User",
			LastName:  "Test",
			Age:       20 + i,
			IsMarried: i%2 == 0,
		}
		err := user.SetPassword("password123")
		s.Require().NoError(err)

		err = s.storage.CreateUser(s.Ctx, user)
		s.Require().NoError(err)

		testUsers = append(testUsers, user)
	}

	req := &domain.GetUsersRequest{
		Ids:    []uuid.UUID{testUsers[0].Id, testUsers[2].Id},
		Limit:  10,
		Offset: 0,
	}
	users, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Len(users, 2)

	ids := []uuid.UUID{users[0].Id, users[1].Id}
	s.Contains(ids, testUsers[0].Id)
	s.Contains(ids, testUsers[2].Id)
}

func (s *UserStorageSuite) TestUsers_NonExistentIds() {
	req := &domain.GetUsersRequest{
		Ids:    []uuid.UUID{uuid.New()},
		Limit:  10,
		Offset: 0,
	}
	users, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Empty(users)
}

func (s *UserStorageSuite) TestCountUsers_All() {
	// Create test users
	for i := 0; i < 3; i++ {
		user := &domain.User{
			FirstName: "Count",
			LastName:  "Test",
			Age:       25,
			IsMarried: false,
		}
		err := user.SetPassword("password123")
		s.Require().NoError(err)

		err = s.storage.CreateUser(s.Ctx, user)
		s.Require().NoError(err)
	}

	req := &domain.GetUsersRequest{}
	count, err := s.storage.CountUsers(s.Ctx, req)
	s.Require().NoError(err)
	s.Equal(3, count)
}

func (s *UserStorageSuite) TestCountUsers_ByIds() {
	// Create test users
	var testUsers []*domain.User
	for i := 0; i < 3; i++ {
		user := &domain.User{
			FirstName: "Count",
			LastName:  "Test",
			Age:       25,
			IsMarried: false,
		}
		err := user.SetPassword("password123")
		s.Require().NoError(err)

		err = s.storage.CreateUser(s.Ctx, user)
		s.Require().NoError(err)

		testUsers = append(testUsers, user)
	}

	req := &domain.GetUsersRequest{
		Ids: []uuid.UUID{testUsers[0].Id, testUsers[1].Id},
	}
	count, err := s.storage.CountUsers(s.Ctx, req)
	s.Require().NoError(err)
	s.Equal(2, count)
}

func (s *UserStorageSuite) TestCountUsers_NonExistentIds() {
	req := &domain.GetUsersRequest{
		Ids: []uuid.UUID{uuid.New()},
	}
	count, err := s.storage.CountUsers(s.Ctx, req)
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *UserStorageSuite) TestUsers_Cache() {
	// Create a test user
	user := &domain.User{
		FirstName: "Cache",
		LastName:  "Test",
		Age:       30,
		IsMarried: false,
	}
	err := user.SetPassword("password123")
	s.Require().NoError(err)

	err = s.storage.CreateUser(s.Ctx, user)
	s.Require().NoError(err)

	req := &domain.GetUsersRequest{
		Ids:    []uuid.UUID{user.Id},
		Limit:  1,
		Offset: 0,
	}

	// First call - should hit database and cache the result
	users1, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Len(users1, 1)

	// Second call with same request - should hit cache
	users2, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Len(users2, 1)
	s.Equal(users1[0].Id, users2[0].Id)

	// Creating new user should invalidate cache
	newUser := &domain.User{
		FirstName: "New",
		LastName:  "User",
		Age:       25,
		IsMarried: true,
	}
	err = newUser.SetPassword("password123")
	s.Require().NoError(err)

	err = s.storage.CreateUser(s.Ctx, newUser)
	s.Require().NoError(err)

	// Cache should be cleared, so this should hit database again
	users3, err := s.storage.Users(s.Ctx, req)
	s.Require().NoError(err)
	s.Len(users3, 1)
}

func (s *UserStorageSuite) TestUsers_RequestValidation() {
	s.Run("zero limit defaults to 10", func() {
		req := &domain.GetUsersRequest{
			Limit:  0,
			Offset: 0,
		}
		_, err := s.storage.Users(s.Ctx, req)
		s.Require().NoError(err)
		s.Equal(10, req.Limit)
	})

	s.Run("limit > 100 capped to 100", func() {
		req := &domain.GetUsersRequest{
			Limit:  200,
			Offset: 0,
		}
		_, err := s.storage.Users(s.Ctx, req)
		s.Require().NoError(err)
		s.Equal(100, req.Limit)
	})

	s.Run("negative offset becomes 0", func() {
		req := &domain.GetUsersRequest{
			Limit:  10,
			Offset: -5,
		}
		_, err := s.storage.Users(s.Ctx, req)
		s.Require().NoError(err)
		s.Equal(0, req.Offset)
	})
}

func TestUserStorageSuite(t *testing.T) {
	suite.Run(t, new(UserStorageSuite))
}
