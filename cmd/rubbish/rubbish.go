package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/leeola/kala/impl/local"
	"github.com/leeola/kala/indexes/bleve"
	"github.com/leeola/kala/stores/disk"
	"github.com/leeola/rubbish"
	"github.com/leeola/rubbish/stores/whala"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rubbish"
	app.Usage = "find your rubbish stuff"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, f",
			Value: "~/.config/rubbish.toml",
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
		{
			Name:    "search",
			Aliases: []string{"s"},
			Usage:   "search for an item",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "container-id, c",
					Usage: "search the constainer id",
				},
				cli.BoolFlag{
					Name:  "description, d",
					Usage: "search the item description",
				},
			},
			Action: SearchCmd,
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

	i := rubbish.Item{
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

func SearchCmd(ctx *cli.Context) error {
	s, err := storeFromCtx(ctx)
	if err != nil {
		return err
	}

	searchFor := strings.Join(ctx.Args(), " ")
	if searchFor == "" {
		return errors.New("text to search for is required")
	}

	var items []rubbish.Item
	switch {
	// case ctx.Bool("container-id"):
	// 	items, err = s.SearchId(searchFor)
	// case ctx.Bool("description"):
	// 	items, err = s.SearchId(searchFor)
	default:
		items, err = s.SearchName(searchFor)
	}
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\t"+strings.Join([]string{
		"ID", "NAME", "CONTAINERID", "DESCRIPTION",
	}, "\t"))
	for _, i := range items {
		fmt.Fprintln(w, "\t"+strings.Join([]string{
			i.Id, i.Name, i.ContainerId, i.Description,
		}, "\t"))
	}

	return w.Flush()
}

func storeFromCtx(ctx *cli.Context) (rubbish.Store, error) {
	// TODO(leeola): Hardcoding implementation for the moment. Remove this.
	// iConf :=

	sConf := disk.Config{
		Path: "_stores/rubbish/store",
	}
	s, err := disk.New(sConf)
	if err != nil {
		return nil, err
	}

	iConf := bleve.Config{
		Path: "_stores/rubbish/index",
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
