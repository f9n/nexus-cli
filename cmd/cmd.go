package cmd

import (
	"fmt"
	"html/template"
	"os"
	"sort"

	"github.com/f9n/nexus-cli/registry"
	"github.com/f9n/nexus-cli/util"
	"github.com/urfave/cli"
)

const (
	CREDENTIALS_TEMPLATES = `# Nexus Credentials
nexus_host = "{{ .Host }}"
nexus_username = "{{ .Username }}"
nexus_password = "{{ .Password }}"
nexus_repository = "{{ .Repository }}"`
)

func SetNexusCredentials(c *cli.Context) error {
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

type ImageInfo struct {
	Name string
	Size int64
}

type ImageInfos []ImageInfo

func ListImages(c *cli.Context) error {
	var allImageInfos []ImageInfo

	var detail = c.Bool("detail")
	var sortBySize = c.Bool("sort-by-size")

	images, err := registry.GetImageNames()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	for _, image := range images {
		if sortBySize {
			imageTotalSize, err := registry.GetTotalImageSize(image)
			if err == nil {
				imageInfo := ImageInfo{
					Name: image,
					Size: imageTotalSize,
				}
				allImageInfos = append(allImageInfos, imageInfo)
			}
		} else {
			if detail {
				imageTotalSize, err := registry.GetTotalImageSizeWithHumanReadable(image)
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
			imageTotalSizeWithHumanReadable := util.GetBytesAsHumanReadable(imageInfo.Size)
			fmt.Printf("%s %s\n", imageTotalSizeWithHumanReadable, imageInfo.Name)
		}
	}
	fmt.Printf("Total images: %d\n", len(allImageInfos))
	return nil
}

func TreeOfAllImages(c *cli.Context) error {
	images, err := registry.GetImageNames()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}

	for _, image := range images {
		fmt.Println(image)
		tags, err := registry.GetTagsByImage(image)
		if err != nil {
			fmt.Print("\tError: ")
			fmt.Println(err)
		}
		for _, tag := range tags {
			fmt.Printf("\t%s\n", tag)
		}
	}
	return nil
}

func ListTagsByImage(c *cli.Context) error {
	var imgName = c.String("name")
	if imgName == "" {
		cli.ShowSubcommandHelp(c)
	}
	tags, err := registry.GetTagsByImage(imgName)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	util.Compare(util.CompareStringNumber).Sort(tags)
	for _, tag := range tags {
		fmt.Println(tag)
	}
	return nil
}

func ShowImageInfo(c *cli.Context) error {
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

func DeleteImage(c *cli.Context) error {
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
				util.Compare(util.CompareStringNumber).Sort(tags)
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

func ShowTotalImageSize(c *cli.Context) error {
	var imgName = c.String("name")
	var isHumanReadable = c.Bool("human-readable")

	if imgName == "" {
		cli.ShowSubcommandHelp(c)
	} else {
		totalSize, err := registry.GetTotalImageSize(imgName)
		if err != nil {
			cli.NewExitError(err, 1)
		}
		if isHumanReadable {
			humanReadableSize := util.GetBytesAsHumanReadable(totalSize)
			fmt.Printf("%s %s\n", humanReadableSize, imgName)
		} else {
			fmt.Printf("%d %s\n", totalSize, imgName)
		}
	}
	return nil
}
