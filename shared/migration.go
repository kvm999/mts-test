package shared

import (
	"database/sql"
	"os"
	"path"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"

	"shared/config"
)

const maxParentDirectoryDepth = 5

type GooseLogger struct {
	logger zerolog.Logger
}

func (s GooseLogger) Fatalf(format string, args ...any) {
	s.logger.Fatal().Msgf(format, args...)
}

func (s GooseLogger) Printf(format string, args ...any) {
	s.logger.Info().Msgf(format, args...)
}

func MigrationDirectory(dialect string) (string, error) {
	pathBuf := path.Join("migration", dialect)
	for range maxParentDirectoryDepth {
		if _, err := os.Stat(pathBuf); err == nil {
			return filepath.Abs(pathBuf)
		}
		pathBuf = path.Join("..", pathBuf)
	}
	return "", ErrMigrationDirectoryNotFound
}

func ApplyMigrations(cfg config.Migration) error {
	migrationDir, err := MigrationDirectory(cfg.Dialect())
	if err != nil {
		return err
	}
	conn, err := sql.Open(cfg.Dialect(), cfg.Dsn())
	if err != nil {
		return err
	}
	if err = goose.SetDialect(cfg.Dialect()); err != nil {
		return err
	}
	goose.SetLogger(GooseLogger{logger: Logger})
	return goose.Up(conn, migrationDir)
}
