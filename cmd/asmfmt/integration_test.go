package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
