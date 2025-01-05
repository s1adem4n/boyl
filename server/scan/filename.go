package scan

import (
	"errors"
	"regexp"
	"strconv"
)

type FilenameMetadata struct {
	Name    string
	Version string
	Year    int
}

// filename is in format: Some Name Here (v<theversion>) (<4digityear>).ext
var filenameRegex = regexp.MustCompile(`(.*) \(v(.*?)\) \((\d{4})\).\w+`)

var ErrInvalidFilename = errors.New("invalid filename")

func ParseFilename(filename string) (*FilenameMetadata, error) {
	matches := filenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return nil, ErrInvalidFilename
	}

	year, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, errors.Join(ErrInvalidFilename, err)
	}

	return &FilenameMetadata{
		Name:    matches[1],
		Version: matches[2],
		Year:    year,
	}, nil
}
