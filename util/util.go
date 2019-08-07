package util

import "github.com/inhies/go-bytesize"

func GetBytesAsHumanReadable(totalSize int64) string {
	humanReadableSize := bytesize.New(float64(totalSize))
	return humanReadableSize.String()
}
