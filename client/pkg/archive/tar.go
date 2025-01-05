package archive

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
)

// ensure TarExtractor implements Extractor
var _ Extractor = (*TarExtractor)(nil)

// TarExtractor extracts tar archives. Since tar is used on top of a compression algorithm, a readCounter is accepted that is the counter for the compression reader. It is then used to calculate the progress.
// e. g. ReadCounter -> GzipReader -> TarExtractor
type TarExtractor struct {
	r           io.Reader
	readCounter *ReadCounter
	size        int64
}

func NewTarExtractor(r io.Reader, counter *ReadCounter, size int64) *TarExtractor {
	return &TarExtractor{
		r:           r,
		readCounter: counter,
		size:        size,
	}
}

func (e *TarExtractor) GetProgressSize() (uint64, error) {
	return uint64(e.size), nil
}

func (e *TarExtractor) Extract(ctx context.Context, basePath string, progress func(uint64)) error {
	tarReader := tar.NewReader(e.r)
	e.readCounter.Progress = progress

loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			h, err := tarReader.Next()
			if err == io.EOF {
				break loop
			}
			if err != nil {
				return err
			}

			err = extractTarFile(ctx, h, tarReader, basePath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func extractTarFile(ctx context.Context, hdr *tar.Header, tr *tar.Reader, basePath string) error {
	destPath := filepath.Join(basePath, hdr.Name)
	if !isWithinBase(basePath, destPath) {
		return ErrIllegalPath
	}

	if hdr.FileInfo().IsDir() {
		return os.MkdirAll(destPath, 0755)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, hdr.FileInfo().Mode())
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
