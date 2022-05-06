package upcloudimport

import (
	"fmt"
	"io/fs"
	"net/http"
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
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}
	switch filepath.Ext(path) {
	case ".gz", ".raw":
		break
	default:
		return nil, fmt.Errorf("only '.raw' and '.gz' files are supported got %s", path)
	}
	im := image{Path: path, info: s}
	if im.SizeGB() > storageMaxSizeGB {
		return nil, fmt.Errorf("storage size %dGB exceeds allowed maximum %dGB", im.SizeGB(), storageMaxSizeGB)
	}

	im.ContentType, err = fileContentType(f)
	return &im, err
}

func fileContentType(f *os.File) (string, error) {
	b := make([]byte, 512)
	if _, err := f.Read(b); err != nil {
		return "", err
	}
	return http.DetectContentType(b), nil
}
