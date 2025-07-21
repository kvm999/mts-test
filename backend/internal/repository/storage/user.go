package storage

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jellydator/ttlcache/v3"

	"mts/internal/domain"
)

func NewUserStorage(pool *pgxpool.Pool) domain.UserStorage {
	return &userStorage{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
		cache: ttlcache.New[domain.CacheKey, []*domain.User](
			ttlcache.WithTTL[domain.CacheKey, []*domain.User](time.Hour),
		),
	}
}

type userStorage struct {
	pool  *pgxpool.Pool
	psql  sq.StatementBuilderType
	cache *ttlcache.Cache[domain.CacheKey, []*domain.User]
}

func (s *userStorage) CreateUser(ctx context.Context, user *domain.User) error {
	s.cache.DeleteAll()

	if err := user.Validate(); err != nil {
		return err
	}

	dto, err := toUserDto(user)
	if err != nil {
		return err
	}

	query := s.psql.Insert("users").
		Columns("id", "first_name", "last_name", "age", "is_married", "password_hash", "salt", "created_at").
		Values(dto.Id, dto.FirstName, dto.LastName, dto.Age, dto.IsMarried, dto.PasswordHash, dto.Salt, dto.CreatedAt)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	return err
}

func (s *userStorage) Users(ctx context.Context, req *domain.GetUsersRequest) ([]*domain.User, error) {
	s.cache.DeleteExpired()
	req.Validate()

	if cacheUsers := s.cache.Get(req.CacheKey()); cacheUsers != nil {
		return cacheUsers.Value(), nil
	}

	query := s.psql.Select("id", "first_name", "last_name", "age", "is_married", "password_hash", "salt", "created_at").
		From("users")

	if len(req.Ids) > 0 {
		query = query.Where(sq.Eq{"id": req.Ids})
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

	var users []*domain.User
	for rows.Next() {
		var dto userDto

		err := rows.Scan(&dto.Id, &dto.FirstName, &dto.LastName, &dto.Age, &dto.IsMarried, &dto.PasswordHash, &dto.Salt, &dto.CreatedAt)
		if err != nil {
			return nil, err
		}

		user, err := dto.toDomain()
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	s.cache.Set(req.CacheKey(), users, ttlcache.DefaultTTL)

	return users, nil
}

func (s *userStorage) CountUsers(ctx context.Context, req *domain.GetUsersRequest) (int, error) {
	req.Validate()

	query := s.psql.Select("COUNT(*)").
		From("users")

	if len(req.Ids) > 0 {
		query = query.Where(sq.Eq{"id": req.Ids})
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
