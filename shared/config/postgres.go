package config

import (
	"fmt"
	"time"
)

type Postgres struct {
	Host              string        `koanf:"host"`
	Port              int           `koanf:"port"`
	Username          string        `koanf:"username"`
	Password          string        `koanf:"password"`
	Database          string        `koanf:"database"`
	SslMode           string        `koanf:"ssl_mode"`
	MaxConns          int32         `koanf:"max_conns"`
	MinConns          int32         `koanf:"min_conns"`
	MaxConnLifetime   time.Duration `koanf:"max_conn_lifetime"`
	MaxConnIdleTime   time.Duration `koanf:"max_conn_idle_time"`
	HealthCheckPeriod time.Duration `koanf:"health_check_period"`
}

func (s *Postgres) Dsn() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		s.Host,
		s.Port,
		s.Username,
		s.Password,
		s.Database,
		s.SslMode,
	)
}

func (s *Postgres) Dialect() string {
	return "postgres"
}
