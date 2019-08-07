package main

import (
	"fmt"
	"os"

	"github.com/f9n/nexus-cli/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Nexus CLI"
	app.Usage = "Manage Docker Private Registry on Nexus"
	app.Version = "1.2.0"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Mohamed Labouardy",
			Email: "mohamed@labouardy.com",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:  "configure",
			Usage: "Configure Nexus Credentials",
			Action: func(c *cli.Context) error {
				return cmd.SetNexusCredentials(c)
			},
		},
		{
			Name:  "image",
			Usage: "Manage Docker Images",
			Subcommands: []cli.Command{
				{
					Name:  "ls",
					Usage: "List all images in repository",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name: "detail",
						},
						cli.BoolFlag{
							Name: "sort-by-size",
						},
					},
					Action: func(c *cli.Context) error {
						return cmd.ListImages(c)
					},
				},
				{
					Name:  "tags",
					Usage: "Display all image tags",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "name, n",
							Usage: "List tags by image name",
						},
					},
					Action: func(c *cli.Context) error {
						return cmd.ListTagsByImage(c)
					},
				},
				{
					Name:  "info",
					Usage: "Show image details",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "name, n",
						},
						cli.StringFlag{
							Name: "tag, t",
						},
					},
					Action: func(c *cli.Context) error {
						return cmd.ShowImageInfo(c)
					},
				},
				{
					Name:  "delete",
					Usage: "Delete an image",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "name, n",
						},
						cli.StringFlag{
							Name: "tag, t",
						},
						cli.StringFlag{
							Name: "keep, k",
						},
						cli.BoolFlag{
							Name: "force, f",
						},
					},
					Action: func(c *cli.Context) error {
						return cmd.DeleteImage(c)
					},
				},
				{
					Name:  "size",
					Usage: "Show total size of image including all tags",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name: "name, n",
						},
						cli.BoolFlag{
							Name: "human-readable",
						},
					},
					Action: func(c *cli.Context) error {
						return cmd.ShowTotalImageSize(c)
					},
				},
				{
					Name:  "tree",
					Usage: "List all tags of images and in repository",
					Flags: []cli.Flag{},
					Action: func(c *cli.Context) error {
						return cmd.TreeOfAllImages(c)
					},
				},
			},
		},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		fmt.Fprintf(c.App.Writer, "Wrong command %q !", command)
	}
	app.Run(os.Args)
}
