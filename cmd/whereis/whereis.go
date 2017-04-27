package main

import (
	"errors"
	"os"

	"github.com/leeola/kala/impl/local"
	"github.com/leeola/kala/indexes/storm"
	"github.com/leeola/kala/stores/disk"
	"github.com/leeola/whereis"
	"github.com/leeola/whereis/stores/whala"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "whereis"
	app.Usage = "find your stuff"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, f",
			Value: "~/.config/whereis/whereis.toml",
			Usage: "load config from `PATH`",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add an item to inventory",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id, i",
					Usage: "the unique id of the item",
				},
				cli.StringFlag{
					Name:  "container-id, c",
					Usage: "the unique id for the container of the item",
				},
				cli.StringFlag{
					Name:  "description, d",
					Usage: "the item description",
				},
				cli.BoolFlag{
					Name:  "allow-no-container",
					Usage: "allow this item to exist without a container",
				},
			},
			Action: AddCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
	}
}

func AddCmd(ctx *cli.Context) error {
	s, err := storeFromCtx(ctx)
	if err != nil {
		return err
	}

	name := ctx.Args().First()
	id := ctx.String("id")
	containerId := ctx.String("container-id")
	description := ctx.String("description")

	if name == "" {
		return errors.New("name is required")
	}

	hasContainer := containerId == ""
	if !hasContainer && !ctx.Bool("allow-no-container") {
		return errors.New("container must be specified without --allow-no-container")
	}

	// default the id to the name
	if id == "" {
		id = name
	}

	i := whereis.Item{
		Id:          id,
		Name:        name,
		ContainerId: containerId,
		Description: description,
	}
	if err := s.Add(i); err != nil {
		return err
	}

	return nil
}

func storeFromCtx(ctx *cli.Context) (whereis.Store, error) {
	// TODO(leeola): Hardcoding implementation for the moment. Remove this.
	// iConf :=

	sConf := disk.Config{
		Path: "_stores/whereis/store",
	}
	s, err := disk.New(sConf)
	if err != nil {
		return nil, err
	}

	iConf := storm.Config{
		Path: "_stores/whereis/index",
	}
	i, err := storm.New(iConf)
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
