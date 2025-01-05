package archive

import (
	"context"
	"io"
	"path/filepath"
	"strings"
)

func isWithinBase(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return !filepath.IsAbs(rel) && !strings.HasPrefix(rel, "..")
}

type ReadCounter struct {
	Total      uint64
	Progress   func(uint64)
	underlying io.Reader
}

func NewReadCounter(r io.Reader, progress func(uint64)) *ReadCounter {
	return &ReadCounter{
		underlying: r,
		Progress:   progress,
	}
}

func (r *ReadCounter) Read(p []byte) (n int, err error) {
	n, err = r.underlying.Read(p)
	r.Total += uint64(n)
	if r.Progress != nil {
		r.Progress(r.Total)
	}
	return
}

// Similar to io.copyBuffer, but with an optional progress callback and a context for cancellation.
func CopyBufferWithProgress(ctx context.Context, dst io.Writer, src io.Reader, buf []byte, progress func(written uint64)) (written int64, err error) {
	if wt, ok := src.(io.WriterTo); ok && progress == nil {
		return wt.WriteTo(dst)
	}
	if rf, ok := dst.(io.ReaderFrom); ok && progress == nil {
		return rf.ReadFrom(src)
	}
	if buf == nil {
		buf = make([]byte, 32*1024)
	}

loop:
	for {
		select {
		case <-ctx.Done():
			return written, ctx.Err()
		default:
			nr, er := src.Read(buf)
			if nr > 0 {
				nw, ew := dst.Write(buf[:nr])
				if nw < 0 || nr < nw {
					nw = 0
					if ew == nil {
						ew = io.ErrShortWrite
					}
				}
				written += int64(nw)
				if progress != nil {
					progress(uint64(nw))
				}
				if ew != nil {
					err = ew
					break loop
				}
				if nr != nw {
					err = io.ErrShortWrite
					break loop
				}
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break loop
			}
		}
	}
	return written, err
}
