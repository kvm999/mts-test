package config

import "fmt"

type ListenAddress struct {
	Host string
	Port int
}

func (s *ListenAddress) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
