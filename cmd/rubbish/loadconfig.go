package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/leeola/errors"
	"github.com/leeola/fixity/autoload"
	"github.com/leeola/rubbish"
	"github.com/leeola/rubbish/stores/whala"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

type Config struct {
	DontExpandHome bool `toml:"dontExpandHome"`

	FixityConfigPath string `toml:"fixityConfigPath"`
}

// TODO(leeola): Change this to a fixity config autoload path / usage
func storeFromCtx(ctx *cli.Context) (rubbish.Store, error) {
	configPath := ctx.GlobalString("config")
	if configPath == "" {
		return nil, errors.New("config PATH is required")
	}

	f, err := os.Open(configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open config")
	}
	defer f.Close()

	var conf Config
	if _, err := toml.DecodeReader(f, &conf); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	if conf.FixityConfigPath == "" {
		return nil, errors.New("missing required config value: FixityConfigPath")
	}

	if !conf.DontExpandHome {
		p, err := homedir.Expand(conf.FixityConfigPath)
		if err != nil {
			return nil, err
		}
		conf.FixityConfigPath = p
	}

	fixi, err := autoload.LoadFixity(conf.FixityConfigPath)
	if err != nil {
		return nil, err
	}

	wConf := whala.Config{
		Fixity: fixi,
	}
	return whala.New(wConf)
}
