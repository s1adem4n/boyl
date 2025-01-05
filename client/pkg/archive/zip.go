package archive

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

// ensure ZipExtractor implements Extractor
var _ Extractor = (*ZipExtractor)(nil)

type ZipExtractor struct {
	r    io.ReaderAt
	size int64
}

func NewZipExtractor(r io.ReaderAt, size int64) *ZipExtractor {
	return &ZipExtractor{
		r:    r,
		size: size,
	}
}

func (e *ZipExtractor) GetProgressSize() (uint64, error) {
	zipReader, err := zip.NewReader(e.r, e.size)
	if err != nil {
		return 0, err
	}

	var totalSize uint64
	for _, f := range zipReader.File {
		totalSize += f.UncompressedSize64
	}

	return totalSize, nil
}

func (e *ZipExtractor) Extract(ctx context.Context, basePath string, progress func(uint64)) error {
	zipReader, err := zip.NewReader(e.r, e.size)
	if err != nil {
		return err
	}

	var currentSize uint64
	var mu sync.Mutex
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(runtime.NumCPU())

	for _, f := range zipReader.File {
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				err := extractZipFile(ctx, f, basePath, func(written uint64) {
					mu.Lock()
					currentSize += written
					if progress != nil {
						progress(currentSize)
					}
					mu.Unlock()
				})
				if err != nil {
					return err
				}
				return nil
			}
		})
	}

	return eg.Wait()
}

func extractZipFile(ctx context.Context, f *zip.File, basePath string, progress func(written uint64)) error {
	destPath := filepath.Join(basePath, f.Name)
	if !isWithinBase(basePath, destPath) {
		return ErrIllegalPath
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(destPath, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = CopyBufferWithProgress(ctx, dst, src, nil, progress)
	if err != nil {
		return err
	}

	return nil
}
