package upcloudimport

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	contentTypeDefault string = "application/octet-stream"
	contentTypeGzip    string = "application/gzip"
)

type image struct {
	Path        string
	ContentType string
	info        fs.FileInfo
}

func NewImage(path string) (*image, error) {
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

	im := image{Path: path, info: s, ContentType: contentTypeDefault}
	if im.SizeGB() > storageMaxSizeGB {
		return nil, fmt.Errorf("storage size %dGB exceeds allowed maximum %dGB", im.SizeGB(), storageMaxSizeGB)
	}

	if ext == ".gz" {
		im.ContentType = contentTypeGzip
	}

	return &im, err
}

// Size returns image size in bytes.
func (i *image) Size() int64 {
	return i.info.Size()
}

// SizeGB return image size in gigabytes rounded up to nearest integer.
func (i *image) SizeGB() int {
	return int(i.Size()/1024/1024/1024) + 1
}

// File returns image file name.
func (i *image) File() string {
	return filepath.Base(i.Path)
}

// CheckSHA256 compares image's sha256 checksum with one provided as parameter
// and returns error if checksum differs or if an error was encountered during reading image checksum.
func (i *image) CheckSHA256(sha256Sum string) error {
	src, err := os.Open(i.Path)
	if err != nil {
		return fmt.Errorf("unable to check '%s' checksum: %w", i.Path, err)
	}
	defer func() {
		_ = src.Close()
	}()

	cs := sha256.New()
	if i.ContentType == contentTypeGzip {
		gsrc, err := gzip.NewReader(src)
		defer func() {
			_ = gsrc.Close()
		}()
		if err != nil {
			return err
		}
		if _, err := io.Copy(cs, gsrc); err != nil {
			return err
		}
	} else {
		if _, err := io.Copy(cs, src); err != nil {
			return err
		}
	}
	csString := hex.EncodeToString(cs.Sum(nil))
	if sha256Sum != csString {
		return fmt.Errorf("uploaded image checksum mismatch want '%s' got '%s'", csString, sha256Sum)
	}
	return nil
}
