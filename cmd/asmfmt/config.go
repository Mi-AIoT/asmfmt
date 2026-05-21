package main

import (
	stderrors "errors"
	"os"
	"path/filepath"

	"github.com/klauspost/asmfmt"
)

const (
	projectConfigName = ".asmfmt.toml"
	globalConfigPath  = "/etc/asmfmt.toml"
)

type configResolver struct {
	explicitPath string
	homePath     string
	globalPath   string
	cache        map[string]asmfmt.Options
	load         func(string) (asmfmt.Options, error)
}

func newConfigResolver(explicitPath string) (*configResolver, error) {
	homePath := ""
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		homePath = filepath.Join(homeDir, projectConfigName)
	}
	return &configResolver{
		explicitPath: explicitPath,
		homePath:     homePath,
		globalPath:   globalConfigPath,
		cache:        make(map[string]asmfmt.Options),
		load:         asmfmt.LoadOptionsFile,
	}, nil
}

func (r *configResolver) optionsForFile(filename string) (asmfmt.Options, error) {
	if r.explicitPath != "" {
		return r.loadCached(r.explicitPath)
	}
	projectPath, err := r.findProjectConfig(filepath.Dir(filename))
	if err != nil {
		return asmfmt.Options{}, err
	}
	if projectPath != "" {
		return r.loadCached(projectPath)
	}
	return r.optionsFromUserOrGlobal()
}

func (r *configResolver) optionsForDir(dir string) (asmfmt.Options, error) {
	if r.explicitPath != "" {
		return r.loadCached(r.explicitPath)
	}
	projectPath, err := r.findProjectConfig(dir)
	if err != nil {
		return asmfmt.Options{}, err
	}
	if projectPath != "" {
		return r.loadCached(projectPath)
	}
	return r.optionsFromUserOrGlobal()
}

func (r *configResolver) optionsForStdin() (asmfmt.Options, error) {
	if r.explicitPath != "" {
		return r.loadCached(r.explicitPath)
	}
	return r.optionsFromUserOrGlobal()
}

func (r *configResolver) optionsFromUserOrGlobal() (asmfmt.Options, error) {
	if r.homePath != "" {
		exists, err := fileExists(r.homePath)
		if err != nil {
			return asmfmt.Options{}, err
		}
		if exists {
			return r.loadCached(r.homePath)
		}
	}
	exists, err := fileExists(r.globalPath)
	if err != nil {
		return asmfmt.Options{}, err
	}
	if exists {
		return r.loadCached(r.globalPath)
	}
	return asmfmt.DefaultOptions(), nil
}

func (r *configResolver) findProjectConfig(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, projectConfigName)
		exists, err := fileExists(candidate)
		if err != nil {
			return "", err
		}
		if exists {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func (r *configResolver) loadCached(path string) (asmfmt.Options, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return asmfmt.Options{}, err
	}
	if opts, ok := r.cache[resolved]; ok {
		return opts, nil
	}
	opts, err := r.load(resolved)
	if err != nil {
		return asmfmt.Options{}, err
	}
	r.cache[resolved] = opts
	return opts, nil
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if stderrors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
