package json

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/leeola/errors"
)

type Config struct {
	Path string
}

type Store struct {
}

func New(c Config) (*Store, error) {
	return &Store{}, nil
}

func FromConfig(configPath string) (*Store, error) {
	f, err := os.Open(configPath)
	if err != nil {
		return nil, false, errors.Wrap(err, "failed to open config")
	}
	defer f.Close()

	var conf struct {
		Config Config `toml:"jsonStore"`
	}
	if _, err := toml.DecodeReader(f, &conf); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	if !structs.IsZero(conf.Config) {
		return nil, false, nil
	}

	s, err := New(conf.Config)
	if err != nil {
		return nil, false, err
	}

	return s, true, nil
}
