package main

import (
	"fmt"
	"html/template"
	"os"
	"sort"

	bytesize "github.com/inhies/go-bytesize"
	"github.com/mlabouardy/nexus-cli/registry"
	"github.com/urfave/cli"
)

const (
	CREDENTIALS_TEMPLATES = `# Nexus Credentials
nexus_host = "{{ .Host }}"
nexus_username = "{{ .Username }}"
nexus_password = "{{ .Password }}"
nexus_repository = "{{ .Repository }}"`
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
				return setNexusCredentials(c)
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
						return listImages(c)
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
						return listTagsByImage(c)
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
						return showImageInfo(c)
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
						return deleteImage(c)
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
						return showTotalImageSize(c)
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

func setNexusCredentials(c *cli.Context) error {
	var hostname, repository, username, password string
	fmt.Print("Enter Nexus Host: ")
	fmt.Scan(&hostname)
	fmt.Print("Enter Nexus Repository Name: ")
	fmt.Scan(&repository)
	fmt.Print("Enter Nexus Username: ")
	fmt.Scan(&username)
	fmt.Print("Enter Nexus Password: ")
	fmt.Scan(&password)

	data := struct {
		Host       string
		Username   string
		Password   string
		Repository string
	}{
		hostname,
		username,
		password,
		repository,
	}

	tmpl, err := template.New(".credentials").Parse(CREDENTIALS_TEMPLATES)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	f, err := os.Create(".credentials")
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func getImageNames() ([]string, error) {
	r, err := registry.NewRegistry()
	if err != nil {
		return []string{}, err
	}
	images, err := r.ListImages()
	if err != nil {
		return []string{}, err
	}

	return images, nil
}

type ImageInfo struct {
	Name string
	Size int64
}

type ImageInfos []ImageInfo

func listImages(c *cli.Context) error {
	var allImageInfos []ImageInfo

	var detail = c.Bool("detail")
	var sortBySize = c.Bool("sort-by-size")

	images, err := getImageNames()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	for _, image := range images {
		if sortBySize {
			imageTotalSize, err := getTotalImageSize(image)
			if err == nil {
				imageInfo := ImageInfo{
					Name: image,
					Size: imageTotalSize,
				}
				allImageInfos = append(allImageInfos, imageInfo)
			}
		} else {
			if detail {
				imageTotalSize, err := getTotalImageSizeWithHumanReadable(image)
				if err != nil {
					fmt.Printf("<%s> ", err.Error())
				}
				fmt.Printf("%s ", imageTotalSize)
			}
			fmt.Println(image)
		}
	}
	if sortBySize {
		sort.Slice(allImageInfos, func(i, j int) bool {
			return allImageInfos[i].Size > allImageInfos[j].Size
		})
		for _, imageInfo := range allImageInfos {
			imageTotalSizeWithHumanReadable := getBytesAsHumanReadable(imageInfo.Size)
			fmt.Printf("%s %s\n", imageTotalSizeWithHumanReadable, imageInfo.Name)
		}
	}
	fmt.Printf("Total images: %d\n", len(allImageInfos))
	return nil
}

func listTagsByImage(c *cli.Context) error {
	var imgName = c.String("name")
	r, err := registry.NewRegistry()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if imgName == "" {
		cli.ShowSubcommandHelp(c)
	}
	tags, err := r.ListTagsByImage(imgName)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	Compare(CompareStringNumber).Sort(tags)
	for _, tag := range tags {
		fmt.Println(tag)
	}
	fmt.Printf("There are %d images for %s\n", len(tags), imgName)
	return nil
}

func showImageInfo(c *cli.Context) error {
	var imgName = c.String("name")
	var tag = c.String("tag")
	r, err := registry.NewRegistry()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	if imgName == "" || tag == "" {
		cli.ShowSubcommandHelp(c)
	}
	manifest, err := r.ImageManifest(imgName, tag)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	fmt.Printf("Image: %s:%s\n", imgName, tag)
	fmt.Printf("Size: %d\n", manifest.Config.Size)
	fmt.Println("Layers:")
	for _, layer := range manifest.Layers {
		fmt.Printf("\t%s\t%d\n", layer.Digest, layer.Size)
	}
	return nil
}

func deleteImage(c *cli.Context) error {
	var imgName = c.String("name")
	var tag = c.String("tag")
	var keep = c.Int("keep")
	var force = c.Bool("force")
	if imgName == "" {
		fmt.Fprintf(c.App.Writer, "You should specify the image name\n")
		cli.ShowSubcommandHelp(c)
	} else {
		r, err := registry.NewRegistry()
		if err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		if tag == "" {
			if keep == 0 {
				tags, err := r.ListTagsByImage(imgName)
				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				if !force {
					var response string
					fmt.Print("We will delete all tags of this image. Are you about that? ('yes' or 'no')")
					fmt.Scan(&response)
					if response != "yes" {
						fmt.Println("Okey. We won't delete these.")
						return nil
					}
				}

				for _, tag := range tags {
					r.DeleteImageByTag(imgName, tag)
				}
			} else {
				tags, err := r.ListTagsByImage(imgName)
				if err != nil {
					return cli.NewExitError(err.Error(), 1)
				}
				Compare(CompareStringNumber).Sort(tags)
				if len(tags) >= keep {
					for _, tag := range tags[:len(tags)-keep] {
						fmt.Printf("%s:%s image will be deleted ...\n", imgName, tag)
						r.DeleteImageByTag(imgName, tag)
					}
				} else {
					fmt.Printf("Only %d images are available\n", len(tags))
				}
			}
		} else {
			err = r.DeleteImageByTag(imgName, tag)
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}
		}
	}
	return nil
}

func getTotalImageSize(imageName string) (int64, error) {
	var totalSize int64
	r, err := registry.NewRegistry()
	if err != nil {
		return 0, err
	}

	tags, err := r.ListTagsByImage(imageName)
	if err != nil {
		return 0, err
	}

	for _, tag := range tags {
		manifest, err := r.ImageManifest(imageName, tag)
		if err != nil {
			return 0, err
		}

		sizeInfo := make(map[string]int64)

		for _, layer := range manifest.Layers {
			sizeInfo[layer.Digest] = layer.Size
		}

		for _, size := range sizeInfo {
			totalSize += size
		}
	}

	return totalSize, nil
}

func getTotalImageSizeWithHumanReadable(imageName string) (string, error) {
	totalSize, err := getTotalImageSize(imageName)
	if err != nil {
		return "", err
	}
	humanReadableSize := getBytesAsHumanReadable(totalSize)
	return humanReadableSize, nil
}

func getBytesAsHumanReadable(totalSize int64) string {
	humanReadableSize := bytesize.New(float64(totalSize))
	return humanReadableSize.String()
}

func showTotalImageSize(c *cli.Context) error {
	var imgName = c.String("name")
	var isHumanReadable = c.Bool("human-readable")

	if imgName == "" {
		cli.ShowSubcommandHelp(c)
	} else {
		totalSize, err := getTotalImageSize(imgName)
		if err != nil {
			cli.NewExitError(err, 1)
		}
		if isHumanReadable {
			humanReadableSize := getBytesAsHumanReadable(totalSize)
			fmt.Printf("%s %s\n", humanReadableSize, imgName)
		} else {
			fmt.Printf("%d %s\n", totalSize, imgName)
		}
	}
	return nil
}
