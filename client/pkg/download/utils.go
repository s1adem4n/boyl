package download

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type MovingAverage struct {
	Values []struct {
		Value float64
		Time  time.Time
	}
	Window time.Duration
}

func NewMovingAverage(window time.Duration) *MovingAverage {
	return &MovingAverage{
		Window: window,
	}
}

func (m *MovingAverage) Add(value float64) {
	m.Values = append(m.Values, struct {
		Value float64
		Time  time.Time
	}{Value: value, Time: time.Now()})

	for i := 0; i < len(m.Values); i++ {
		if time.Since(m.Values[i].Time) > m.Window {
			m.Values = m.Values[i:]
			break
		}
	}
}

func (m *MovingAverage) Get() float64 {
	if len(m.Values) == 0 {
		return 0
	}

	var total float64
	for _, v := range m.Values {
		total += v.Value
	}
	return total / float64(len(m.Values))
}

func getPathDepth(path string) int {
	return len(strings.Split(path, string(filepath.Separator)))
}

func isExecutable(info os.FileInfo) bool {
	return info.Mode()&0111 != 0 || strings.HasSuffix(info.Name(), ".exe")
}

func FindExecutablePath(directory string) (string, error) {
	var path string
	depth := -1

	err := filepath.WalkDir(directory, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if isExecutable(info) {
			thisDepth := getPathDepth(p)
			if thisDepth > depth {
				path = p
				depth = thisDepth
			}
		}

		return nil
	})

	return path, err
}
