package registry

import (
	"github.com/f9n/nexus-cli/util"
)

func GetImageNames() ([]string, error) {
	r, err := NewRegistry()
	if err != nil {
		return []string{}, err
	}
	images, err := r.ListImages()
	if err != nil {
		return []string{}, err
	}

	return images, nil
}

func GetTagsByImage(imgName string) ([]string, error) {
	r, err := NewRegistry()
	if err != nil {
		return []string{}, err
	}
	tags, err := r.ListTagsByImage(imgName)
	if err != nil {
		return []string{}, err
	}
	return tags, nil
}

func GetTotalImageSize(imageName string) (int64, error) {
	var totalSize int64
	r, err := NewRegistry()
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

func GetTotalImageSizeWithHumanReadable(imageName string) (string, error) {
	totalSize, err := GetTotalImageSize(imageName)
	if err != nil {
		return "", err
	}
	humanReadableSize := util.GetBytesAsHumanReadable(totalSize)
	return humanReadableSize, nil
}
