package asset

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mholt/archives"

	filetype "gopkg.in/h2non/filetype.v1"
	filetype_types "gopkg.in/h2non/filetype.v1/types"
)

const (
	// Size of file header for sniffing type
	headerSize = 262
)

// A archiveExpander detects the archive type and expands it to the local
// filesystem.
//
// Supported archive types:
// - tar
// - tar-gzip
type archiveExpander struct{}

type namer interface {
	Name() string
}

// Expand an archive to a target directory.
func (a *archiveExpander) Expand(archive io.ReadSeeker, targetDirectory string) error {
	ft, err := sniffType(archive)
	if err != nil {
		return err
	}

	var format archives.Extractor

	switch ft.MIME.Value {
	case "application/x-tar":
		format = archives.Tar{}
	case "application/gzip":
		format = archives.CompressedArchive{
			Compression: archives.Gz{},
			Archival:    archives.Tar{},
			Extraction:  archives.Tar{},
		}
	default:
		return fmt.Errorf(
			"given file of format '%s' does not appear valid",
			ft.MIME.Value,
		)
	}

	_, ok := archive.(namer)
	if !ok {
		return errors.New("couldn't get path to archive")
	}

	ctx := context.Background()
	err = format.Extract(ctx, archive, func(ctx context.Context, f archives.FileInfo) error {
		if err := os.MkdirAll(filepath.Dir(targetDirectory), 0755); err != nil {
			return fmt.Errorf("failed to create directory structure: %w", err)
		}
		destPath := filepath.Join(targetDirectory, f.Name())
		if f.IsDir() {
			return os.MkdirAll(destPath, f.Mode())
		}
		destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
		defer destFile.Close()

		fileReader, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file from archive: %w", err)
		}
		defer fileReader.Close()

		_, err = io.Copy(destFile, fileReader)
		if err != nil {
			return fmt.Errorf("failed to copy file: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func sniffType(f io.ReadSeeker) (filetype_types.Type, error) {
	header := make([]byte, headerSize)
	if _, err := f.Read(header); err != nil {
		return filetype_types.Type{}, fmt.Errorf("unable to read asset header: %s", err)
	}
	ft, err := filetype.Match(header)
	if err != nil {
		return ft, err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return filetype_types.Type{}, err
	}

	return ft, nil
}
