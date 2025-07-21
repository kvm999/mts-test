package shared

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit"
	"github.com/go-ldap/ldap/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"shared/config"
)

const (
	postgresImage    = "postgres:16-alpine"
	postgresUsername = "postgres"
	postgresPassword = "postgres"
	postgresDatabase = "postgres"

	ldapImage = "osixia/openldap:1.5.0"

	mailhogImage = "mailhog/mailhog:v1.0.1"
)

type Suite[S any] struct {
	suite.Suite

	Ctx    context.Context
	Cancel context.CancelFunc
	Config *config.Config[S]
	Logger zerolog.Logger

	PostgresEnabled   bool
	PostgresContainer *postgres.PostgresContainer
	PostgresConn      *pgxpool.Pool

	LdapEnabled   bool
	LdapContainer testcontainers.Container
	LdapDomain    string
	LdapUsername  string
	LdapPassword  string
	LdapConn      *ldap.Conn
	LdapPort      int

	MailhogEnabled   bool
	MailhogContainer testcontainers.Container
	MailhogPort      int
}

func (s *Suite[S]) SetupSuite() {
	s.Ctx, s.Cancel = context.WithCancel(s.T().Context())

	s.Ctx = Logger.WithContext(s.Ctx)
	s.Logger = Logger

	// Initialize LDAP settings
	s.LdapDomain = "example.com"
	s.LdapUsername = "admin"
	s.LdapPassword = gofakeit.Password(true, true, true, true, false, 12)

	var err error
	s.Config, err = config.Load[S]("", "")
	s.Require().NoError(err)
	s.Logger.Info().Msg("config loaded")

	if s.PostgresEnabled {
		s.startPostgres()
		s.migratePostgres()
	}

	if s.MailhogEnabled {
		s.startMailhog()
	}

	if s.LdapEnabled {
		s.startLdap()
	}
}

func (s *Suite[S]) startPostgres() {
	// NOTE: https://github.com/testcontainers/testcontainers-go/issues/279#issuecomment-866840540
	s.NoError(os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"))

	// start postgres container
	var err error
	s.PostgresContainer, err = postgres.Run(
		s.Ctx,
		postgresImage,
		postgres.WithDatabase(postgresDatabase),
		postgres.WithUsername(postgresUsername),
		postgres.WithPassword(postgresPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithStartupTimeout(10*time.Second),
			wait.ForListeningPort("5432/tcp")),
	)
	s.Require().NoError(err)

	// update config database port
	containerPort, err := s.PostgresContainer.MappedPort(s.Ctx, "5432")
	s.Require().NoError(err)
	s.Config.Postgres = &config.Postgres{
		Host:              "localhost",
		Port:              containerPort.Int(),
		Username:          postgresUsername,
		Password:          postgresPassword,
		Database:          postgresDatabase,
		SslMode:           "disable",
		MaxConns:          5,
		MinConns:          1,
		MaxConnLifetime:   time.Minute,
		MaxConnIdleTime:   time.Minute,
		HealthCheckPeriod: time.Second * 30,
	}

	s.Logger.Info().
		Str("host", s.Config.Postgres.Host).
		Int("port", s.Config.Postgres.Port).
		Msg("postgres up")

	// connect to postgres
	s.PostgresConn, err = ConnectPostgres(s.Ctx, s.Config.Postgres)
	s.Require().NoError(err)
}

func (s *Suite[S]) migratePostgres() {
	if !s.PostgresEnabled {
		return
	}

	postgresMigrationDir, err := MigrationDirectory("postgres")
	if err != nil {
		s.Logger.Warn().Err(err).Msg("postgres skip migrations")
		s.Require().ErrorIs(err, ErrMigrationDirectoryNotFound)
		return
	}

	// apply migrations
	postgresNativeConn, err := sql.Open("postgres", s.Config.Postgres.Dsn())
	s.Require().NoError(err)

	s.NoError(goose.SetDialect("postgres"))
	s.NoError(goose.Up(postgresNativeConn, postgresMigrationDir))
}

func (s *Suite[S]) startMailhog() {
	req := testcontainers.ContainerRequest{
		Image:        mailhogImage,
		ExposedPorts: []string{"1025/tcp", "8025/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("1025/tcp"),
		),
	}
	var err error
	s.MailhogContainer, err = testcontainers.GenericContainer(s.Ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)

	mailhogPort, err := s.MailhogContainer.MappedPort(s.Ctx, "1025")
	s.Require().NoError(err)
	s.MailhogPort = mailhogPort.Int()
}

func (s *Suite[S]) startLdap() {
	s.NoError(os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"))

	var err error
	containerRequest := testcontainers.ContainerRequest{
		Image:        ldapImage,
		ExposedPorts: []string{"389/tcp", "636/tcp"},
		Env: map[string]string{
			"LDAP_ORGANISATION":          "Example Organization",
			"LDAP_DOMAIN":                s.LdapDomain,
			"LDAP_ADMIN_PASSWORD":        s.LdapPassword,
			"LDAP_CONFIG_PASSWORD":       s.LdapPassword,
			"LDAP_READONLY_USER":         "false",
			"LDAP_RFC2307BIS_SCHEMA":     "false",
			"LDAP_BACKEND":               "mdb",
			"LDAP_TLS":                   "true",
			"LDAP_TLS_CRT_FILENAME":      "ldap.crt",
			"LDAP_TLS_KEY_FILENAME":      "ldap.key",
			"LDAP_TLS_DH_PARAM_FILENAME": "dhparam.pem",
			"LDAP_TLS_CA_CRT_FILENAME":   "ca.crt",
			"LDAP_TLS_ENFORCE":           "false",
			"LDAP_TLS_CIPHER_SUITE":      "SECURE256:-VERS-SSL3.0",
			"LDAP_TLS_VERIFY_CLIENT":     "demand",
			"LDAP_REPLICATION":           "false",
		},
		WaitingFor: wait.ForLog("slapd starting").WithStartupTimeout(30 * time.Second),
	}

	s.LdapContainer, err = testcontainers.GenericContainer(
		s.Ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerRequest,
			Started:          true,
		},
	)
	s.Require().NoError(err)

	// Get the mapped port and store it
	containerPort, err := s.LdapContainer.MappedPort(s.Ctx, "389")
	s.Require().NoError(err)
	s.LdapPort = containerPort.Int()

	s.Logger.Info().
		Int("port", s.LdapPort).
		Msg("ldap up")

	// Set up LDAP connection
	ldapUrl := fmt.Sprintf("ldap://localhost:%d", s.LdapPort)
	s.LdapConn, err = ldap.DialURL(ldapUrl)
	s.Require().NoError(err)

	// Bind as admin
	adminDn := fmt.Sprintf("cn=%s,dc=%s", s.LdapUsername, strings.ReplaceAll(s.LdapDomain, ".", ",dc="))
	err = s.LdapConn.Bind(adminDn, s.LdapPassword)
	s.Require().NoError(err)
}

func (s *Suite[S]) CreateLdapUser(username, password, email string) string {
	baseDn := fmt.Sprintf("dc=%s", strings.ReplaceAll(s.LdapDomain, ".", ",dc="))
	userDn := fmt.Sprintf("uid=%s,%s", username, baseDn)

	if email == "" {
		email = username + "@" + s.LdapDomain
	}

	addReq := ldap.NewAddRequest(userDn, nil)
	addReq.Attribute("objectClass", []string{"inetOrgPerson"})
	addReq.Attribute("uid", []string{username})
	addReq.Attribute("cn", []string{username})
	addReq.Attribute("sn", []string{"TestUser"})
	addReq.Attribute("mail", []string{email})
	addReq.Attribute("userPassword", []string{password})

	err := s.LdapConn.Add(addReq)
	s.Require().NoError(err)

	return userDn
}

func (s *Suite[S]) TearDownSuite() {
	if s.PostgresContainer != nil {
		s.NoError(s.PostgresContainer.Terminate(s.Ctx))
	}

	if s.MailhogContainer != nil {
		s.NoError(s.MailhogContainer.Terminate(s.Ctx))
	}

	if s.LdapContainer != nil {
		s.NoError(s.LdapContainer.Terminate(s.Ctx))
	}

	if s.LdapConn != nil {
		s.LdapConn.Close()
	}

	s.Cancel()
}
