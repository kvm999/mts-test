package config

import (
	"encoding/hex"
	"fmt"
	"time"
)

type Service struct {
	JwtSecret     string        `koanf:"jwt_secret"`
	TokenLifetime time.Duration `koanf:"token_lifetime"`

	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

func (s *Service) RestListenAddress() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s *Service) JwtSecretBytes() ([]byte, error) {
	return hex.DecodeString(s.JwtSecret)
}
