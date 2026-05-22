package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLIConfigExplicitFile(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "input.s")
	cfg := filepath.Join(root, "explicit.toml")
	writeFile(t, src, "TEXT foo(SB),$0\nMOVQ AX,BX\n")
	writeFile(t, cfg, "indent_style = \"space\"\nindent_width = 2\n")

	stdout, stderr, err := runCLI(t, root, nil, nil, "-config", cfg, src)
	if err != nil {
		t.Fatalf("runCLI: %v\nstderr:\n%s", err, stderr)
	}
	want := "TEXT foo(SB), $0\n  MOVQ AX, BX\n"
	if stdout != want {
		t.Fatalf("explicit config output = %q; want %q", stdout, want)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func TestCLIConfigProjectDiscovery(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	nested := filepath.Join(project, "nested")
	src := filepath.Join(nested, "input.s")
	cfg := filepath.Join(project, ".asmfmt.toml")
	writeFile(t, cfg, "indent_style = \"space\"\nindent_width = 2\n")
	writeFile(t, src, "TEXT foo(SB),$0\nMOVQ AX,BX\n")

	env := withHomeEnv(filepath.Join(root, "home"))
	stdout, stderr, err := runCLI(t, nested, env, nil, src)
	if err != nil {
		t.Fatalf("runCLI: %v\nstderr:\n%s", err, stderr)
	}
	want := "TEXT foo(SB), $0\n  MOVQ AX, BX\n"
	if stdout != want {
		t.Fatalf("project config output = %q; want %q", stdout, want)
	}
}

func TestCLIConfigStdinUsesUserConfigOnly(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	home := filepath.Join(root, "home")
	writeFile(t, filepath.Join(project, ".asmfmt.toml"), "indent_style = \"space\"\nindent_width = 4\n")
	writeFile(t, filepath.Join(home, ".asmfmt.toml"), "line_comment_space = false\n")

	env := withHomeEnv(home)
	stdout, stderr, err := runCLI(t, project, env, []byte("// comment\naddi a0, a0, 1 // note\n"))
	if err != nil {
		t.Fatalf("runCLI: %v\nstderr:\n%s", err, stderr)
	}
	want := "//comment\naddi a0, a0, 1 //note\n"
	if stdout != want {
		t.Fatalf("stdin config output = %q; want %q", stdout, want)
	}
}

func runCLI(t *testing.T, dir string, env []string, stdin []byte, args ...string) (string, string, error) {
	t.Helper()
	bin := buildCLI(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func buildCLI(t *testing.T) string {
	t.Helper()
	name := "asmfmt-test"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	bin := filepath.Join(t.TempDir(), name)
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build cmd/asmfmt: %v\n%s", err, output)
	}
	return bin
}

func withHomeEnv(home string) []string {
	return []string{
		"HOME=" + home,
		"USERPROFILE=" + home,
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o666); err != nil {
		t.Fatal(err)
	}
}
func TestCLIInitOption(t *testing.T) {
	root := t.TempDir()

	// Run -init for the first time
	_, stderr, err := runCLI(t, root, nil, nil, "-init")
	if err != nil {
		t.Fatalf("runCLI -init: %v\nstderr:\n%s", err, stderr)
	}

	cfgPath := filepath.Join(root, ".asmfmt.toml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Fatal(".asmfmt.toml was not created")
	}

	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfgBytes) == 0 {
		t.Fatal(".asmfmt.toml is empty")
	}

	// Run -init again - should fail and error out
	_, stderr2, err2 := runCLI(t, root, nil, nil, "-init")
	if err2 == nil {
		t.Fatal("expected -init to fail when .asmfmt.toml already exists")
	}
	if !bytes.Contains([]byte(stderr2), []byte("already exists")) {
		t.Fatalf("expected already exists error, got: %s", stderr2)
	}
}
func TestCLIVersionOption(t *testing.T) {
	root := t.TempDir()

	stdout, stderr, err := runCLI(t, root, nil, nil, "-version")
	if err != nil {
		t.Fatalf("runCLI -version: %v\nstderr:\n%s", err, stderr)
	}

	if !bytes.Contains([]byte(stdout), []byte("version:")) {
		t.Fatalf("expected stdout to contain version:, got: %s", stdout)
	}
	if !bytes.Contains([]byte(stdout), []byte("git hash:")) {
		t.Fatalf("expected stdout to contain git hash:, got: %s", stdout)
	}
	if !bytes.Contains([]byte(stdout), []byte("build time:")) {
		t.Fatalf("expected stdout to contain build time:, got: %s", stdout)
	}
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
}

func createMockArchive(t *testing.T, dummyContent []byte) []byte {
	var buf bytes.Buffer
	binaryName := "asmfmt"
	if runtime.GOOS == "windows" {
		binaryName = "asmfmt.exe"
	}

	if runtime.GOOS == "windows" {
		zw := zip.NewWriter(&buf)
		f, err := zw.Create(binaryName)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(dummyContent); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		gw := gzip.NewWriter(&buf)
		tw := tar.NewWriter(gw)
		hdr := &tar.Header{
			Name: binaryName,
			Mode: 0755,
			Size: int64(len(dummyContent)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(dummyContent); err != nil {
			t.Fatal(err)
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gw.Close(); err != nil {
			t.Fatal(err)
		}
	}
	return buf.Bytes()
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		t.Fatal(err)
	}
}

func TestCLIUpgradeOption(t *testing.T) {
	archiveBytes := createMockArchive(t, []byte("MOCK_BINARY_UPGRADE_CONTENT"))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			assetName := fmt.Sprintf("asmfmt_%s_%s", runtime.GOOS, runtime.GOARCH)
			if runtime.GOOS == "windows" {
				assetName += ".zip"
			} else {
				assetName += ".tar.gz"
			}
			downloadURL := fmt.Sprintf("http://%s/download/%s", r.Host, assetName)

			jsonStr := fmt.Sprintf(`{
				"tag_name": "v1.5.0",
				"assets": [
					{
						"name": "%s",
						"browser_download_url": "%s"
					}
				]
			}`, assetName, downloadURL)
			w.Write([]byte(jsonStr))
			return
		}

		if strings.Contains(r.URL.Path, "/download/") {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(archiveBytes)
			return
		}

		http.NotFound(w, r)
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	tempBin := filepath.Join(tempDir, filepath.Base(bin))
	copyFile(t, bin, tempBin)

	cmd := exec.Command(tempBin, "-update", "latest")
	cmd.Env = append(os.Environ(),
		"ASMFMT_UPDATE_URL="+ts.URL,
		"ASMFMT_UPGRADE_REPO=test-owner/test-repo",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("update failed: %v\nstdout:\n%s\nstderr:\n%s", err, stdout.String(), stderr.String())
	}

	newContent, err := os.ReadFile(tempBin)
	if err != nil {
		t.Fatalf("reading updated binary: %v", err)
	}

	if string(newContent) != "MOCK_BINARY_UPGRADE_CONTENT" {
		t.Fatalf("unexpected content after upgrade: got %q, want %q", string(newContent), "MOCK_BINARY_UPGRADE_CONTENT")
	}

	if runtime.GOOS == "windows" {
		if _, err := os.Stat(tempBin + ".old"); err != nil {
			t.Fatalf("expected windows old binary to exist, got error: %v", err)
		}
	}
}

func TestCLIUpgradeOptionInvalid(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	bin := buildCLI(t)
	tempDir := t.TempDir()
	tempBin := filepath.Join(tempDir, filepath.Base(bin))
	copyFile(t, bin, tempBin)

	cmd := exec.Command(tempBin, "-update", "invalid")
	cmd.Env = append(os.Environ(),
		"ASMFMT_UPDATE_URL="+ts.URL,
		"ASMFMT_UPGRADE_REPO=test-owner/test-repo",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected update to fail but it succeeded")
	}

	errStr := stderr.String()
	if !strings.Contains(errStr, "404") && !strings.Contains(errStr, "Not Found") {
		t.Fatalf("expected error message to contain HTTP 404 or Not Found error, got: %s", errStr)
	}
}

func TestCLIExitCodeOnDiff(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "input.s")

	// Unformatted code (needs formatting: space inside comma, indentation)
	unformatted := "TEXT foo(SB),$0\nMOVQ AX,BX\n"

	// 1. Check with -d (should exit with code 1, and print diff)
	writeFile(t, src, unformatted)
	stdout, stderr, err := runCLI(t, root, nil, nil, "-d", src)
	if err == nil {
		t.Fatalf("expected non-zero exit code with -d on unformatted file, got 0")
	}
	if !strings.Contains(stdout, "diff") {
		t.Fatalf("expected stdout to contain diff, got: %q", stdout)
	}

	// 2. Check with -l (should exit with code 1, and print filename)
	writeFile(t, src, unformatted)
	stdout, stderr, err = runCLI(t, root, nil, nil, "-l", src)
	if err == nil {
		t.Fatalf("expected non-zero exit code with -l on unformatted file, got 0")
	}
	if !strings.Contains(stdout, "input.s") {
		t.Fatalf("expected stdout to contain input.s, got: %q", stdout)
	}

	// 3. Run without -d or -l or -w (should exit with 0, and print formatted code to stdout)
	writeFile(t, src, unformatted)
	stdout, stderr, err = runCLI(t, root, nil, nil, src)
	if err != nil {
		t.Fatalf("unexpected error without diff flags: %v\nstderr:\n%s", err, stderr)
	}
	expectedFormatted := "TEXT foo(SB), $0\n\tMOVQ AX, BX\n"
	if stdout != expectedFormatted {
		t.Fatalf("unexpected formatted output: got %q, want %q", stdout, expectedFormatted)
	}

	// 4. Run with -w (should exit with 0, and modify file in-place)
	writeFile(t, src, unformatted)
	stdout, stderr, err = runCLI(t, root, nil, nil, "-w", src)
	if err != nil {
		t.Fatalf("unexpected error with -w: %v\nstderr:\n%s", err, stderr)
	}
	fileBytes, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if string(fileBytes) != expectedFormatted {
		t.Fatalf("expected file to be formatted in-place, got %q", string(fileBytes))
	}
}

func TestCLILint(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "input.s")

	// 1. Valid code (should pass with exit code 0, no output)
	validCode := "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a1, 1\n\tret\n"
	writeFile(t, src, validCode)
	stdout, stderr, err := runCLI(t, root, nil, nil, "-lint", src)
	if err != nil {
		t.Fatalf("expected valid file to pass linting, got err: %v\nstderr:\n%s", err, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected no stderr output, got: %q", stderr)
	}

	// 2. Invalid code (should fail with exit code 1, print violation to stderr)
	invalidCode := "// Copyright\n// SPDX-License-Identifier: Apache\n\taddi a0, a1, 1\n\taddi x10, x11, 1\n\tret\n"
	writeFile(t, src, invalidCode)
	stdout, stderr, err = runCLI(t, root, nil, nil, "-lint", src)
	if err == nil {
		t.Fatalf("expected invalid file to fail linting, but got exit code 0")
	}
	if stdout != "" {
		t.Fatalf("expected stdout to be empty, got: %q", stdout)
	}
	if !strings.Contains(stderr, "[L101][abi_registers]") {
		t.Fatalf("expected stderr to contain rule L101 violation, got: %q", stderr)
	}
	expectedFormatSub := "input.s:4: [L101][abi_registers] register \"x10\" should be replaced with its ABI name (error)"
	if !strings.Contains(stderr, expectedFormatSub) {
		t.Fatalf("expected stderr to contain format %q, got: %q", expectedFormatSub, stderr)
	}

	// 3. Invalid code via stdin
	stdout, stderr, err = runCLI(t, root, nil, []byte(invalidCode), "-lint")
	if stdout != "" {
		t.Fatalf("expected stdin stdout to be empty, got: %q", stdout)
	}
	if err == nil {
		t.Fatalf("expected stdin invalid code to fail linting, but got exit code 0")
	}
	if !strings.Contains(stderr, "<standard input>:4: [L101][abi_registers]") {
		t.Fatalf("expected stdin stderr to contain <standard input>:4: [L101][abi_registers], got: %q", stderr)
	}
}
