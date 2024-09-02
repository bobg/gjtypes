package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRun(t *testing.T) {
	updateGolden, _ := strconv.ParseBool(os.Getenv("UPDATE_GOLDEN"))

	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), "~") {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".golden") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			f, err := os.Open(filepath.Join("testdata", entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			got := new(bytes.Buffer)

			if err := run(got, f); err != nil {
				t.Fatal(err)
			}

			goldenPath := filepath.Join("testdata", entry.Name()+".golden")

			if updateGolden {
				if err := os.WriteFile(goldenPath, got.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(string(want), got.String()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}

	if updateGolden {
		t.Fatal("golden files updated")
	}
}
