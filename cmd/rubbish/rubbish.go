package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/leeola/rubbish"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rubbish"
	app.Usage = "find your rubbish stuff"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "config, f",
			Value:  "~/.config/rubbish/config.toml",
			Usage:  "load config from `PATH`",
			EnvVar: "RUBBISH_CONFIG",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add an item to inventory",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "tag, t",
					Usage: "specify a tag for the given item",
				},
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
				cli.StringSliceFlag{
					Name:  "tag, t",
					Usage: "search for a given tag",
				},
				cli.BoolFlag{
					Name:  "container-id, c",
					Usage: "search the container id",
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
		fmt.Println(err.Error())
		os.Exit(1)
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

	hasContainer := containerId != ""
	if !hasContainer && !ctx.Bool("allow-no-container") {
		return errors.New("container-id is required without --allow-no-container flag")
	}

	i := rubbish.Item{
		Id:          id,
		Name:        name,
		ContainerId: containerId,
		Description: description,
		Tags:        ctx.StringSlice("tag"),
	}
	id, err = s.Add(i)
	if err != nil {
		return err
	}

	fmt.Println("added id:", id)

	return nil
}

func SearchCmd(ctx *cli.Context) error {
	s, err := storeFromCtx(ctx)
	if err != nil {
		return err
	}

	tags := ctx.StringSlice("tag")
	searchFor := strings.Join(ctx.Args(), " ")

	var items []rubbish.Item
	switch {
	// case ctx.Bool("container-id"):
	// 	items, err = s.SearchId(searchFor)
	case ctx.Bool("description"):
		items, err = s.SearchDescription(searchFor, tags)
	default:
		items, err = s.Search(searchFor, tags)
	}
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	fmt.Fprintln(w, "\t"+strings.Join([]string{
		"ID", "NAME", "CONTAINERID", "DESCRIPTION", "TAGS",
	}, "\t"))
	for _, i := range items {
		fmt.Fprintln(w, "\t"+strings.Join([]string{
			i.Id, i.Name, i.ContainerId, i.Description, strings.Join(i.Tags, ", "),
		}, "\t"))
	}

	return w.Flush()
}
