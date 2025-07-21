package storage

import (
	"encoding/hex"
	"time"

	"github.com/google/uuid"

	"mts/internal/domain"
)

type userDto struct {
	Id           uuid.UUID `db:"id"`
	FirstName    string    `db:"first_name"`
	LastName     string    `db:"last_name"`
	Age          int       `db:"age"`
	IsMarried    bool      `db:"is_married"`
	PasswordHash string    `db:"password_hash"`
	Salt         string    `db:"salt"`
	CreatedAt    time.Time `db:"created_at"`
}

func (dto *userDto) toDomain() (*domain.User, error) {
	user := &domain.User{
		Id:        dto.Id,
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Age:       dto.Age,
		IsMarried: dto.IsMarried,
		CreatedAt: dto.CreatedAt,
	}

	if dto.PasswordHash != "" {
		passwordHash, err := hex.DecodeString(dto.PasswordHash)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = passwordHash
	}

	if dto.Salt != "" {
		salt, err := hex.DecodeString(dto.Salt)
		if err != nil {
			return nil, err
		}
		user.Salt = salt
	}

	return user, nil
}

func toUserDto(user *domain.User) (*userDto, error) {
	dto := &userDto{
		Id:        user.Id,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age:       user.Age,
		IsMarried: user.IsMarried,
		CreatedAt: user.CreatedAt,
	}

	if len(user.PasswordHash) > 0 {
		dto.PasswordHash = hex.EncodeToString(user.PasswordHash)
	}

	if len(user.Salt) > 0 {
		dto.Salt = hex.EncodeToString(user.Salt)
	}

	return dto, nil
}
