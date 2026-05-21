package asmfmt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOptionalCorpus(t *testing.T) {
	root := strings.TrimSpace(os.Getenv("ASMFMT_CORPUS_DIR"))
	if root == "" {
		t.Skip("set ASMFMT_CORPUS_DIR to run optional corpus formatting checks")
	}

	var files []string
	patterns := []string{"*.s", "*.S", "*.asm"}
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(root, "**", pattern))
		if err == nil && len(matches) > 0 {
			files = append(files, matches...)
		}
	}
	if len(files) == 0 {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".s" || ext == ".asm" {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(files) == 0 {
		t.Fatalf("no assembly files found under %q", root)
	}

	for _, path := range files {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Format(strings.NewReader(string(src)))
			if err != nil {
				t.Fatalf("format %s: %v", path, err)
			}
			got2, err := Format(strings.NewReader(string(got)))
			if err != nil {
				t.Fatalf("reformat %s: %v", path, err)
			}
			if string(got) != string(got2) {
				t.Fatalf("non-idempotent formatting for %s", path)
			}
		})
	}
}
