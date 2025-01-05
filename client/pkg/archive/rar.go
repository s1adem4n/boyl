package archive

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/nwaples/rardecode"
)

// ensure RarExtractor implements Extractor
var _ Extractor = (*RarExtractor)(nil)

type RarExtractor struct {
	r    io.Reader
	size int64
}

func NewRarExtractor(r io.Reader, size int64) *RarExtractor {
	return &RarExtractor{
		r:    r,
		size: size,
	}
}

func (e *RarExtractor) GetProgressSize() (uint64, error) {
	return uint64(e.size), nil
}

func (e *RarExtractor) Extract(ctx context.Context, basePath string, progress func(uint64)) error {
	readCounter := NewReadCounter(e.r, progress)
	rarReader, err := rardecode.NewReader(readCounter, "")
	if err != nil {
		return err
	}

loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			h, err := rarReader.Next()
			if err == io.EOF {
				break loop
			}
			if err != nil {
				return err
			}

			err = extractRarFile(ctx, h, rarReader, basePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func extractRarFile(ctx context.Context, hdr *rardecode.FileHeader, tr io.Reader, basePath string) error {
	destPath := filepath.Join(basePath, hdr.Name)
	if !isWithinBase(basePath, destPath) {
		return ErrIllegalPath
	}

	if hdr.IsDir {
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}
	}

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = CopyBufferWithProgress(ctx, dst, tr, nil, nil)
	if err != nil {
		return err
	}

	return nil
}
