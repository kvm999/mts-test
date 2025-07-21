package bootstrap

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"mts/internal/application"
	"mts/internal/config"
	"mts/internal/domain"
	"mts/internal/repository/storage"
	"mts/internal/transport/rest"
	"shared"
	sharedConfig "shared/config"
)

const (
	EnvPrefix      = "MTS"
	ConfigFilename = "config.yaml"
)

func NewApp() *Application {
	return &Application{}
}

type Application struct {
	Ctx    context.Context
	Cancel context.CancelFunc

	Config *sharedConfig.Config[config.Service]
	Logger zerolog.Logger

	// database connection
	PostgresConnection *pgxpool.Pool

	// repository
	UserStorage    domain.UserStorage
	ProductStorage domain.ProductStorage
	OrderStorage   domain.OrderStorage

	// application service
	UserAppService    domain.UserAppService
	ProductAppService domain.ProductAppService
	OrderAppService   domain.OrderAppService

	// transport
	RestServer *fiber.App
}

func (s *Application) Initialize() error {
	s.Ctx, s.Cancel = context.WithCancel(context.Background())

	var err error

	if s.Config == nil {
		s.Config, err = sharedConfig.Load[config.Service](EnvPrefix, ConfigFilename)
		if err != nil {
			return err
		}
	}

	// logger
	s.Logger = shared.Logger
	s.Ctx = s.Logger.WithContext(s.Ctx)

	s.PostgresConnection, err = shared.ConnectPostgres(s.Ctx, s.Config.Postgres)
	if err != nil {
		return err
	}

	// repository
	s.UserStorage = storage.NewUserStorage(s.PostgresConnection)
	s.ProductStorage = storage.NewProductStorage(s.PostgresConnection)
	s.OrderStorage = storage.NewOrderStorage(s.PostgresConnection)

	// application service
	s.UserAppService = application.NewUserAppService(s.UserStorage)
	s.ProductAppService = application.NewProductAppService(s.ProductStorage)
	s.OrderAppService = application.NewOrderAppService(s.OrderStorage, s.ProductStorage, s.UserStorage)

	s.Logger.Info().Msg("application initialized")

	return nil
}

func (s *Application) Start() error {
	ctx, cancel := signal.NotifyContext(s.Ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// rest server init
	s.RestServer = rest.New(s.UserAppService, s.ProductAppService, s.OrderAppService)

	// apply migrations
	if err := shared.ApplyMigrations(s.Config.Postgres); err != nil {
		return err
	}

	// start the application
	var eg errgroup.Group

	eg.Go(func() error {
		s.Logger.Info().
			Str("host", s.Config.Service.Host).
			Int("port", s.Config.Service.Port).
			Msg("rest server started")

		return s.RestServer.Listen(s.Config.Service.RestListenAddress(), fiber.ListenConfig{
			GracefulContext:       s.Ctx,
			DisableStartupMessage: true,
		})
	})

	eg.Go(func() error {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.RestServer.ShutdownWithContext(shutdownCtx)
	})

	s.Logger.Info().Msg("application started")

	level := zerolog.InfoLevel
	err := eg.Wait()
	if err != nil {
		level = zerolog.ErrorLevel
	}

	s.Logger.WithLevel(level).Err(err).Msg("application stopped")

	return err
}
