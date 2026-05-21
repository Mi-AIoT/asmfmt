package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/klauspost/asmfmt"
)

func TestConfigResolverExplicitPath(t *testing.T) {
	root := t.TempDir()
	explicit := writeConfigFile(t, root, "explicit.toml", "indent_style = \"space\"\nindent_width = 2\n")
	writeConfigFile(t, root, ".asmfmt.toml", "indent_width = 4\n")

	resolver := testResolver(root, explicit)
	opts, err := resolver.optionsForFile(filepath.Join(root, "file.s"))
	if err != nil {
		t.Fatal(err)
	}
	if opts.IndentStyle != "space" || opts.IndentWidth != 2 {
		t.Fatalf("explicit config not applied: %#v", opts)
	}
}

func TestConfigResolverProjectPrecedence(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	subdir := filepath.Join(project, "sub", "dir")
	user := filepath.Join(root, "home", ".asmfmt.toml")
	global := filepath.Join(root, "etc", "asmfmt.toml")
	writeConfigFile(t, project, ".asmfmt.toml", "indent_width = 2\n")
	writeConfigFile(t, filepath.Dir(user), filepath.Base(user), "indent_width = 4\n")
	writeConfigFile(t, filepath.Dir(global), filepath.Base(global), "indent_width = 6\n")

	resolver := testResolverWithPaths(user, global, "")
	opts, err := resolver.optionsForFile(filepath.Join(subdir, "file.s"))
	if err != nil {
		t.Fatal(err)
	}
	if opts.IndentWidth != 2 {
		t.Fatalf("project config not preferred: %#v", opts)
	}
}

func TestConfigResolverFallsBackToUserThenGlobal(t *testing.T) {
	root := t.TempDir()
	user := filepath.Join(root, "home", ".asmfmt.toml")
	global := filepath.Join(root, "etc", "asmfmt.toml")
	writeConfigFile(t, filepath.Dir(user), filepath.Base(user), "indent_width = 4\n")
	writeConfigFile(t, filepath.Dir(global), filepath.Base(global), "indent_width = 6\n")

	resolver := testResolverWithPaths(user, global, "")
	opts, err := resolver.optionsForFile(filepath.Join(root, "project", "file.s"))
	if err != nil {
		t.Fatal(err)
	}
	if opts.IndentWidth != 4 {
		t.Fatalf("user config not used: %#v", opts)
	}

	if err := removeFile(user); err != nil {
		t.Fatal(err)
	}
	opts, err = resolver.optionsForFile(filepath.Join(root, "project", "file2.s"))
	if err != nil {
		t.Fatal(err)
	}
	if opts.IndentWidth != 6 {
		t.Fatalf("global config not used: %#v", opts)
	}
}

func TestConfigResolverStdinSkipsProjectLookup(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	user := filepath.Join(root, "home", ".asmfmt.toml")
	writeConfigFile(t, project, ".asmfmt.toml", "indent_width = 2\n")
	writeConfigFile(t, filepath.Dir(user), filepath.Base(user), "indent_width = 4\n")

	resolver := testResolverWithPaths(user, filepath.Join(root, "etc", "asmfmt.toml"), "")
	opts, err := resolver.optionsForStdin()
	if err != nil {
		t.Fatal(err)
	}
	if opts.IndentWidth != 4 {
		t.Fatalf("stdin should use user/global only: %#v", opts)
	}
}

func TestConfigResolverDirectoryUsesRootConfig(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	subdir := filepath.Join(project, "nested")
	writeConfigFile(t, project, ".asmfmt.toml", "indent_width = 2\n")
	writeConfigFile(t, subdir, ".asmfmt.toml", "indent_width = 4\n")

	resolver := testResolver(root, "")
	opts, err := resolver.optionsForDir(project)
	if err != nil {
		t.Fatal(err)
	}
	if opts.IndentWidth != 2 {
		t.Fatalf("directory walk should reuse root config: %#v", opts)
	}
}

func TestConfigResolverMissingExplicitConfigErrors(t *testing.T) {
	root := t.TempDir()
	resolver := testResolver(root, filepath.Join(root, "missing.toml"))
	if _, err := resolver.optionsForStdin(); err == nil {
		t.Fatal("expected missing explicit config error")
	}
}

func TestConfigResolverInvalidConfigErrors(t *testing.T) {
	root := t.TempDir()
	bad := writeConfigFile(t, root, "bad.toml", "indent_width = \"oops\"\n")
	resolver := testResolver(root, bad)
	if _, err := resolver.optionsForStdin(); err == nil {
		t.Fatal("expected invalid TOML type error")
	}

	unknown := writeConfigFile(t, root, "unknown.toml", "mystery = true\n")
	resolver = testResolver(root, unknown)
	if _, err := resolver.optionsForStdin(); err == nil || !strings.Contains(err.Error(), "unknown config field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestConfigResolverCachesByResolvedPath(t *testing.T) {
	root := t.TempDir()
	configPath := writeConfigFile(t, root, ".asmfmt.toml", "indent_width = 2\n")
	loads := 0
	resolver := testResolver(root, "")
	resolver.load = func(path string) (asmfmt.Options, error) {
		loads++
		return asmfmt.LoadOptionsFile(path)
	}
	if _, err := resolver.optionsForFile(filepath.Join(root, "one.s")); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.optionsForFile(filepath.Join(root, ".", "two.s")); err != nil {
		t.Fatal(err)
	}
	if loads != 1 {
		t.Fatalf("expected cached config for %s, got %d loads", configPath, loads)
	}
}

func testResolver(root, explicit string) *configResolver {
	return testResolverWithPaths(filepath.Join(root, "home", ".asmfmt.toml"), filepath.Join(root, "etc", "asmfmt.toml"), explicit)
}

func testResolverWithPaths(home, global, explicit string) *configResolver {
	return &configResolver{
		explicitPath: explicit,
		homePath:     home,
		globalPath:   global,
		cache:        make(map[string]asmfmt.Options),
		load:         asmfmt.LoadOptionsFile,
	}
}

func writeConfigFile(t *testing.T, dir, name, contents string) string {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o666); err != nil {
		t.Fatal(err)
	}
	return path
}

func removeFile(path string) error {
	return os.Remove(path)
}
