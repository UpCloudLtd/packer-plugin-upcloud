package upcloudimport

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type image struct {
	Path        string
	ContentType string
	info        fs.FileInfo
}

func (i *image) Size() int64 {
	return i.info.Size()
}

func (i *image) SizeGB() int {
	return int(i.Size()/1024/1024/1024) + 1
}

func (i *image) File() string {
	return filepath.Base(i.Path)
}

func newImage(path string) (*image, error) {
	s, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(path)
	switch ext {
	case ".gz", ".raw":
		break
	default:
		return nil, fmt.Errorf("only '.raw' and '.gz' files are supported got %s", path)
	}

	im := image{Path: path, info: s, ContentType: "application/octet-stream"}
	if im.SizeGB() > storageMaxSizeGB {
		return nil, fmt.Errorf("storage size %dGB exceeds allowed maximum %dGB", im.SizeGB(), storageMaxSizeGB)
	}

	if ext == ".gz" {
		im.ContentType = "application/gzip"
	}

	return &im, err
}
