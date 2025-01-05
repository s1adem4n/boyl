package archive

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
	"github.com/ulikunitz/xz"
	"github.com/ulikunitz/xz/lzma"
)

var ErrIllegalPath = errors.New("illegal path")
var ErrUnsupportedArchive = errors.New("unsupported archive")

// Extractor is an interface for extracting archives.
type Extractor interface {
	// Returns the total size for calculating the progress with the progress function in Extract.
	// If possible, this is the uncompressed size of the archive.
	// For tar and other sequential archives, this is the size of the archive file.
	GetProgressSize() (uint64, error)
	// Extracts the archive to the given basePath. Accepts a context that cancels the extraction. Not resumable.
	Extract(ctx context.Context, basePath string, progress func(uint64)) error
}

// NewExtractor returns an Extractor for the given archive file.
func NewExtractor(filename string, r io.ReaderAt, size int64) (Extractor, error) {
	if strings.HasSuffix(filename, ".zip") {
		return NewZipExtractor(r, size), nil
	}
	if strings.HasSuffix(filename, ".7z") {
		return NewSevenZipExtractor(r, size), nil
	}
	if strings.HasSuffix(filename, ".rar") {
		return NewRarExtractor(io.NewSectionReader(r, 0, size), size), nil
	}

	if strings.HasSuffix(filename, ".tar.gz") {
		readCounter := NewReadCounter(io.NewSectionReader(r, 0, size), nil)
		gzipReader, err := pgzip.NewReader(readCounter)
		if err != nil {
			return nil, err
		}
		return NewTarExtractor(gzipReader, readCounter, size), nil
	}
	if strings.HasSuffix(filename, ".tar.zst") {
		readCounter := NewReadCounter(io.NewSectionReader(r, 0, size), nil)
		zstdReader, err := zstd.NewReader(readCounter)
		if err != nil {
			return nil, err
		}
		return NewTarExtractor(zstdReader, readCounter, size), nil
	}
	if strings.HasSuffix(filename, ".tar.xz") {
		readCounter := NewReadCounter(io.NewSectionReader(r, 0, size), nil)
		xzReader, err := xz.NewReader(readCounter)
		if err != nil {
			return nil, err
		}
		return NewTarExtractor(xzReader, readCounter, size), nil
	}
	if strings.HasSuffix(filename, ".tar.lzma") {
		readCounter := NewReadCounter(io.NewSectionReader(r, 0, size), nil)
		lzmaReader, err := lzma.NewReader(readCounter)
		if err != nil {
			return nil, err
		}
		return NewTarExtractor(lzmaReader, readCounter, size), nil
	}
	if strings.HasSuffix(filename, ".tar") {
		readCounter := NewReadCounter(io.NewSectionReader(r, 0, size), nil)
		return NewTarExtractor(readCounter, readCounter, size), nil
	}

	return nil, ErrUnsupportedArchive
}
