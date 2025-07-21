package shared

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	_ "github.com/lib/pq"

	"shared/config"
)

func ConnectPostgres(ctx context.Context, cfg *config.Postgres) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.Dsn())
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	zerolog.Ctx(ctx).
		Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("username", cfg.Username).
		Str("database", cfg.Database).
		Msg("connected to postgres")

	return pool, nil
}
