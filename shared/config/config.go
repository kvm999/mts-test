package config

import (
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

const defaultEnvPrefix = "APP"
const defaultFilename = "config.yaml"

type Config[S any] struct {
	Logger       *Logger   `koanf:"logger"`
	Postgres     *Postgres `koanf:"postgres"`
	Service      *S        `koanf:"service"`
	FrontBaseUrl string    `koanf:"front_base_url"`
}

func Load[S any](envPrefix string, filename string) (*Config[S], error) {
	var cfg Config[S]

	k := koanf.New(".")

	if envPrefix == "" {
		envPrefix = defaultEnvPrefix
	}

	if !strings.HasSuffix(envPrefix, "_") {
		envPrefix += "_"
	}

	if filename == "" {
		filename = defaultFilename
	}

	if _, err := os.Stat(filename); !os.IsNotExist(err) {
		if err = k.Load(file.Provider(filename), yaml.Parser()); err != nil {
			return nil, err
		}
	}

	err := k.Load(env.Provider(envPrefix, ".", func(s string) string {
		key := strings.ToLower(strings.TrimPrefix(s, envPrefix))
		parts := strings.Split(key, "_")
		if len(parts) >= 3 {
			result := parts[0] + "." + strings.Join(parts[1:], "_")
			return result
		}
		return strings.Replace(key, "_", ".", -1) // TODO: make robust implementation
	}), nil)
	if err != nil {
		return nil, err
	}

	if err = k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
