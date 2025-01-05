package archive

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/bodgit/sevenzip"
)

// ensure SevenZipExtractor implements Extractor
var _ Extractor = (*SevenZipExtractor)(nil)

type SevenZipExtractor struct {
	r    io.ReaderAt
	size int64
}

func NewSevenZipExtractor(r io.ReaderAt, size int64) *SevenZipExtractor {
	return &SevenZipExtractor{
		r:    r,
		size: size,
	}
}

func (e *SevenZipExtractor) GetProgressSize() (uint64, error) {
	reader, err := sevenzip.NewReader(e.r, e.size)
	if err != nil {
		return 0, err
	}

	var totalSize uint64
	for _, f := range reader.File {
		totalSize += f.UncompressedSize
	}

	return totalSize, nil
}

func (e *SevenZipExtractor) Extract(ctx context.Context, basePath string, progress func(uint64)) error {
	zipReader, err := sevenzip.NewReader(e.r, e.size)
	if err != nil {
		return err
	}

	var currentSize uint64

	for _, f := range zipReader.File {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := extractSevenZipFile(ctx, f, basePath, func(written uint64) {
				currentSize += written
				if progress != nil {
					progress(currentSize)
				}
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func extractSevenZipFile(ctx context.Context, f *sevenzip.File, basePath string, progress func(written uint64)) error {
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
