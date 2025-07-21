package config

type Migration interface {
	Dsn() string
	Dialect() string
}
