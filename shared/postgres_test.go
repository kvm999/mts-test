package shared

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type PostgresSuite struct {
	Suite[any]
}

func (s *PostgresSuite) SetupSuite() {
	s.PostgresEnabled = true
	s.Suite.SetupSuite()
}

func (s *PostgresSuite) TestConnectPostgres() {
	postgresConn, err := ConnectPostgres(s.Ctx, s.Config.Postgres)
	s.Require().NoError(err)
	s.Require().NotNil(postgresConn)

	rows, err := postgresConn.Query(s.Ctx, "select version();")
	s.Require().NoError(err)
	s.Require().True(rows.Next())
	var version string
	s.Require().NoError(rows.Scan(&version))
	s.Require().NotEmpty(version)
	s.Require().Contains(version, "PostgreSQL")
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
