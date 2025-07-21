package domain

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CacheKey = [sha256.Size]byte

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

type User struct {
	Id           uuid.UUID
	FirstName    string
	LastName     string
	Age          int
	IsMarried    bool
	PasswordHash []byte
	Salt         []byte
	CreatedAt    time.Time
}

func (u *User) Validate() error {
	if u.Id == uuid.Nil {
		u.Id = uuid.New()
	}

	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	if strings.TrimSpace(u.FirstName) == "" {
		return fmt.Errorf("%w: first name is required", ErrUserValidation)
	}

	if strings.TrimSpace(u.LastName) == "" {
		return fmt.Errorf("%w: last name is required", ErrUserValidation)
	}

	if u.Age < 18 {
		return fmt.Errorf("%w: user must be at least 18 years old", ErrUserValidation)
	}

	if len(u.PasswordHash) == 0 {
		return fmt.Errorf("%w: password hash is required", ErrUserValidation)
	}

	if len(u.Salt) == 0 {
		return fmt.Errorf("%w: salt is required", ErrUserValidation)
	}

	return nil
}

func (u *User) FullName() string {
	return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

func (u *User) SetPassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", ErrUserValidation)
	}

	// Generate salt
	salt := uuid.New()
	u.Salt = salt[:]

	// Hash password with salt
	saltedPassword := append([]byte(password), u.Salt...)
	hash, err := bcrypt.GenerateFromPassword(saltedPassword, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: failed to hash password: %v", ErrUserValidation, err)
	}

	u.PasswordHash = hash
	return nil
}

func (u *User) VerifyPassword(password string) bool {
	saltedPassword := append([]byte(password), u.Salt...)
	err := bcrypt.CompareHashAndPassword(u.PasswordHash, saltedPassword)
	return err == nil
}

type CreateUserRequest struct {
	FirstName string
	LastName  string
	Age       int
	IsMarried bool
	Password  string
}

func (r *CreateUserRequest) Validate() error {
	if strings.TrimSpace(r.FirstName) == "" {
		return fmt.Errorf("%w: first name is required", ErrUserValidation)
	}

	if strings.TrimSpace(r.LastName) == "" {
		return fmt.Errorf("%w: last name is required", ErrUserValidation)
	}

	if r.Age < 18 {
		return fmt.Errorf("%w: user must be at least 18 years old", ErrUserValidation)
	}

	if len(r.Password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", ErrUserValidation)
	}

	return nil
}

func (r *CreateUserRequest) ToDomain() (*User, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}

	user := &User{
		FirstName: strings.TrimSpace(r.FirstName),
		LastName:  strings.TrimSpace(r.LastName),
		Age:       r.Age,
		IsMarried: r.IsMarried,
	}

	if err := user.SetPassword(r.Password); err != nil {
		return nil, err
	}

	return user, nil
}

type GetUsersRequest struct {
	Ids    []uuid.UUID
	Limit  int
	Offset int
}

func (r *GetUsersRequest) Validate() {
	if r.Limit <= 0 {
		r.Limit = 10
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
}

func (r *GetUsersRequest) CacheKey() CacheKey {
	buf := make([]byte, 0)

	// ids
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(r.Ids)))
	for _, id := range r.Ids {
		buf = append(buf, id[:]...)
	}

	// pagination
	buf = binary.BigEndian.AppendUint32(buf, uint32(r.Limit))
	buf = binary.BigEndian.AppendUint32(buf, uint32(r.Offset))

	return sha256.Sum256(buf)
}

type UserStorage interface {
	CreateUser(ctx context.Context, user *User) error
	Users(ctx context.Context, req *GetUsersRequest) ([]*User, error)
	CountUsers(ctx context.Context, req *GetUsersRequest) (int, error)
}

type UserAppService interface {
	RegisterUser(ctx context.Context, req *CreateUserRequest) (*User, error)
	Users(ctx context.Context, req *GetUsersRequest) ([]*User, error)
	CountUsers(ctx context.Context, req *GetUsersRequest) (int, error)
}
