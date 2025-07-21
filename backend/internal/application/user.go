package application

import (
	"context"

	"mts/internal/domain"

	"github.com/rs/zerolog"
)

func NewUserAppService(userStorage domain.UserStorage) domain.UserAppService {
	return &userAppService{
		userStorage: userStorage,
	}
}

type userAppService struct {
	userStorage domain.UserStorage
}

func (s *userAppService) RegisterUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "RegisterUser").
		Str("first_name", req.FirstName).
		Str("last_name", req.LastName).
		Int("age", req.Age).
		Logger()

	logger.Info().Msg("registering new user")

	user, err := req.ToDomain()
	if err != nil {
		logger.Error().Err(err).Msg("failed to convert request to domain")
		return nil, err
	}

	err = s.userStorage.CreateUser(ctx, user)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create user in storage")
		return nil, err
	}

	logger.Info().
		Str("user_id", user.Id.String()).
		Msg("user registered successfully")

	return user, nil
}

func (s *userAppService) Users(ctx context.Context, req *domain.GetUsersRequest) ([]*domain.User, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "Users").
		Int("limit", req.Limit).
		Int("offset", req.Offset).
		Int("ids_count", len(req.Ids)).
		Logger()

	logger.Info().Msg("fetching users")

	users, err := s.userStorage.Users(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch users from storage")
		return nil, err
	}

	logger.Info().
		Int("users_count", len(users)).
		Msg("users fetched successfully")

	return users, nil
}

func (s *userAppService) CountUsers(ctx context.Context, req *domain.GetUsersRequest) (int, error) {
	logger := zerolog.Ctx(ctx).With().
		Str("operation", "CountUsers").
		Int("ids_count", len(req.Ids)).
		Logger()

	logger.Debug().Msg("counting users")

	count, err := s.userStorage.CountUsers(ctx, req)
	if err != nil {
		logger.Error().Err(err).Msg("failed to count users in storage")
		return 0, err
	}

	logger.Debug().
		Int("count", count).
		Msg("users counted successfully")

	return count, nil
}
