package main

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/leeola/errors"
	"github.com/leeola/kala/impl/local"
	"github.com/leeola/kala/indexes/bleve"
	"github.com/leeola/kala/stores/disk"
	"github.com/leeola/rubbish"
	"github.com/leeola/rubbish/stores/whala"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

type Config struct {
	DontExpandHome bool `toml:"dontExpandHome"`

	// TODO(leeola): Change this to a kala config autoload path / usage
	KalaStorePath string `toml:"kalaStorePath"`
}

// TODO(leeola): Change this to a kala config autoload path / usage
func storeFromCtx(ctx *cli.Context) (rubbish.Store, error) {
	configPath := ctx.GlobalString("config")
	if configPath == "" {
		return nil, errors.New("config PATH is required")
	}

	configPath, err := homedir.Expand(configPath)
	if err != nil {
		return nil, err
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

	if conf.KalaStorePath == "" {
		return nil, errors.New("missing required config value: KalaStorePath")
	}

	if !conf.DontExpandHome {
		p, err := homedir.Expand(conf.KalaStorePath)
		if err != nil {
			return nil, err
		}
		conf.KalaStorePath = p
	}

	sConf := disk.Config{
		Path: filepath.Join(conf.KalaStorePath, "store"),
	}
	s, err := disk.New(sConf)
	if err != nil {
		return nil, err
	}

	iConf := bleve.Config{
		Path: filepath.Join(conf.KalaStorePath, "index"),
	}
	i, err := bleve.New(iConf)
	if err != nil {
		return nil, err
	}

	kConf := local.Config{
		Store: s,
		Index: i,
	}
	k, err := local.New(kConf)
	if err != nil {
		return nil, err
	}

	wConf := whala.Config{
		Kala: k,
	}
	return whala.New(wConf)
}
