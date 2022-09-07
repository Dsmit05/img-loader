package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type FileLoader struct {
	cfg Config
}

func NewFileLoader(cfg Config) (*FileLoader, error) {
	err := CreatePath(cfg.BasePath)
	if err != nil {
		return nil, err
	}
	return &FileLoader{cfg: cfg}, nil
}

func (f *FileLoader) Load(_ context.Context, url, path, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		errS := fmt.Sprintf("http.NewRequest(GET, %v, nil)", url)
		return errors.Wrap(err, errS)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "http.DefaultClient.Do()")
	}
	defer res.Body.Close()

	if name == "" {
		name = f.filenameFromURL(url)
	}

	if path != "" {
		err := CreatePath(filepath.Join(f.cfg.BasePath, path))
		if err != nil {
			return errors.Wrap(err, "CreatePath()")
		}
	}

	createFileIn := filepath.Join(f.cfg.BasePath, path, name)
	newFile, err := os.Create(createFileIn)
	if err != nil {
		errS := fmt.Sprintf("os.Create(%v)", createFileIn)
		return errors.Wrap(err, errS)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, err = io.CopyN(newFile, res.Body, f.cfg.CopyBufferSize)
			if err != nil {
				if err == io.EOF {
					return nil
				} else {
					return errors.Wrap(err, "io.CopyN()")
				}
			}

		}
	}

}

func (f *FileLoader) GenerateTask(ctx context.Context, url, path, name string) func() error {
	return func() error {
		return f.Load(ctx, url, path, name)
	}
}

func (f *FileLoader) filenameFromURL(url string) string {
	filename := path.Base(url)
	index := strings.IndexRune(filename, '?')
	if index != -1 {
		filename = filename[:index]
	}

	return filename
}
