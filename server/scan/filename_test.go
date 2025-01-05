package scan_test

import (
	"boyl/server/scan"
	"testing"
)

func TestParseName(t *testing.T) {
	tests := []struct {
		filename string
		expected *scan.FilenameMetadata
	}{
		{"Anno 1404 Gold Edition (v2.01) (2010).7z", &scan.FilenameMetadata{"Anno 1404 Gold Edition", "2.01", 2010}},
		{"Anno 1503 AD (v2.0.0.5) (2003).7z", &scan.FilenameMetadata{"Anno 1503 AD", "2.0.0.5", 2003}},
		{"Anno 1602 (v1.05) (1998).7z", &scan.FilenameMetadata{"Anno 1602", "1.05", 1998}},
		{"Anno 1701 AD (v2.0.0.4) (2006).7z", &scan.FilenameMetadata{"Anno 1701 AD", "2.0.0.4", 2006}},
		{"Balatro (v1.0.1n) (2024).7z", &scan.FilenameMetadata{"Balatro", "1.0.1n", 2024}},
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			meta, err := scan.ParseFilename(test.filename)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if meta == nil {
				t.Fatalf("expected non-nil meta")
			}
			if meta.Name != test.expected.Name || meta.Version != test.expected.Version || meta.Year != test.expected.Year {
				t.Errorf("expected %+v, got %+v", test.expected, meta)
			}
		})
	}
}
