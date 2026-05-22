// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Modified by Klaus Post 2015 for asmfmt
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"runtime/pprof"
	"strings"

	"github.com/klauspost/asmfmt"
)

var (
	version   = "devel"
	buildTime = "unknown"
	gitHash   = "unknown"
)

var (
	// main operation modes
	list        = flag.Bool("l", false, "list files whose formatting differs from asmfmt's")
	write       = flag.Bool("w", false, "write result to (source) file instead of stdout")
	doDiff      = flag.Bool("d", false, "display diffs instead of rewriting files")
	allErrors   = flag.Bool("e", false, "report all errors (not just the first 10 on different lines)")
	config      = flag.String("config", "", "read formatting options from this TOML file")
	initCfg     = flag.Bool("init", false, "create a default .asmfmt.toml configuration file in the current directory")
	showVersion = flag.Bool("version", false, "print version information and exit")
	update      = flag.String("update", "", "upgrade to the specified version: 'latest', 'beta', or an arbitrary tag (e.g., 'v2.0.0')")

	// debugging
	cpuprofile = flag.String("cpuprofile", "", "write CPU profile to file (primarily for debugging and performance analysis)")
	lintFlag   = flag.Bool("lint", false, "check style rules and report violations")
)

const (
	tabWidth = 8
)

var (
	exitCode = 0
	errors   = 0
)

func report(err error) {
	fmt.Fprintln(os.Stderr, err)
	errors++
	if !*allErrors && errors >= 10 {
		os.Exit(2)
	}
	exitCode = 2
}

func initVersion() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if gitHash == "unknown" && setting.Value != "" {
					gitHash = setting.Value
				}
			case "vcs.time":
				if buildTime == "unknown" && setting.Value != "" {
					buildTime = setting.Value
				}
			}
		}
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			if version == "devel" {
				version = info.Main.Version
			}
		}
	}
}

func usage() {
	initVersion()
	fmt.Fprintf(os.Stderr, "asmfmt version: %s (commit: %s, built: %s)\n\n", version, gitHash, buildTime)
	fmt.Fprintf(os.Stderr, "usage: asmfmt [flags] [path ...]\n\n")
	fmt.Fprintf(os.Stderr, "asmfmt formats Go/Plan 9, GAS, and RISC-V assembly source files.\n")
	fmt.Fprintf(os.Stderr, "If no paths are provided, it reads from standard input and writes to standard output.\n\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nConfiguration Discovery:\n")
	fmt.Fprintf(os.Stderr, "  If -config is not specified, asmfmt searches for a config in this order:\n")
	fmt.Fprintf(os.Stderr, "  1. The nearest \".asmfmt.toml\" walking upward from the formatted file's directory.\n")
	fmt.Fprintf(os.Stderr, "  2. \"~/.asmfmt.toml\"\n")
	fmt.Fprintf(os.Stderr, "  3. \"/etc/asmfmt.toml\"\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  Print version information:\n")
	fmt.Fprintf(os.Stderr, "    asmfmt -version\n\n")
	fmt.Fprintf(os.Stderr, "  Create a default configuration file in the current directory:\n")
	fmt.Fprintf(os.Stderr, "    asmfmt -init\n\n")
	fmt.Fprintf(os.Stderr, "  Format a file in place:\n")
	fmt.Fprintf(os.Stderr, "    asmfmt -w path/to/file.s\n\n")
	fmt.Fprintf(os.Stderr, "  Format all assembly files in a directory tree:\n")
	fmt.Fprintf(os.Stderr, "    asmfmt -w ./...\n\n")
	fmt.Fprintf(os.Stderr, "  Show diff of formatting changes for a file:\n")
	fmt.Fprintf(os.Stderr, "    asmfmt -d path/to/file.s\n\n")
	fmt.Fprintf(os.Stderr, "  Format standard input with a specific configuration:\n")
	fmt.Fprintf(os.Stderr, "    cat file.s | asmfmt -config /path/to/asmfmt.toml\n")
	os.Exit(2)
}

func isAsmFile(f os.FileInfo) bool {
	// ignore non-Asm files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".s")
}

// If in == nil, the source is the contents of the file with the given filename.
func processFile(filename string, in io.Reader, out io.Writer, stdin bool, opts asmfmt.Options) error {
	if in == nil {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		in = f
	}

	src, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}

	if *lintFlag {
		problems, err := asmfmt.Lint(filename, bytes.NewBuffer(src), opts)
		if err != nil {
			return err
		}
		hasError := false
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, p)
			if p.Severity == "error" {
				hasError = true
			}
		}
		if hasError && exitCode == 0 {
			exitCode = 1
		}
		return nil
	}

	res, err := asmfmt.FormatWithOptions(bytes.NewBuffer(src), opts)
	if err != nil {
		return err
	}

	if !bytes.Equal(src, res) {
		// formatting has changed
		if *list || *doDiff {
			exitCode = 1
		}
		if *list {
			fmt.Fprintln(out, filename)
		}
		if *write {
			err = ioutil.WriteFile(filename, res, 0644)
			if err != nil {
				return err
			}
		}
		if *doDiff {
			data, err := diff(src, res)
			if err != nil {
				return fmt.Errorf("computing diff: %s", err)
			}
			fmt.Printf("diff %s asmfmt/%s\n", filename, filename)
			out.Write(data)
		}
	}

	if !*list && !*write && !*doDiff {
		_, err = out.Write(res)
	}

	return err
}

func visitFile(path string, f os.FileInfo, err error, opts asmfmt.Options) error {
	if err == nil && isAsmFile(f) {
		err = processFile(path, nil, os.Stdout, false, opts)
	}
	if err != nil {
		report(err)
	}
	return nil
}

func walkDir(path string, opts asmfmt.Options) {
	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		return visitFile(walkPath, info, err, opts)
	})
}

func main() {
	// call gofmtMain in a separate function
	// so that it can use defer and have them
	// run before the exit.
	gofmtMain()
	os.Exit(exitCode)
}

func gofmtMain() {
	flag.Usage = usage
	flag.Parse()

	initVersion()
	if *showVersion {
		fmt.Printf("asmfmt version: %s\n", version)
		fmt.Printf("git hash: %s\n", gitHash)
		fmt.Printf("build time: %s\n", buildTime)
		return
	}

	if *update != "" {
		if err := runUpdate(*update); err != nil {
			fmt.Fprintf(os.Stderr, "upgrade failed: %v\n", err)
			exitCode = 2
			return
		}
		return
	}

	if *initCfg {
		filename := ".asmfmt.toml"
		if _, err := os.Stat(filename); err == nil {
			fmt.Fprintf(os.Stderr, "error: %s already exists in the current directory\n", filename)
			exitCode = 2
			return
		}
		err := os.WriteFile(filename, asmfmt.DefaultConfigTemplate, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing configuration file: %s\n", err)
			exitCode = 2
			return
		}
		fmt.Printf("Created default configuration file: %s\n", filename)
		return
	}

	resolver, err := newConfigResolver(*config)
	if err != nil {
		report(err)
		return
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "creating cpu profile: %s\n", err)
			exitCode = 2
			return
		}
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flag.NArg() == 0 {
		if *write {
			fmt.Fprintln(os.Stderr, "error: cannot use -w with standard input")
			exitCode = 2
			return
		}
		opts, err := resolver.optionsForStdin()
		if err != nil {
			report(err)
			return
		}
		if err := processFile("<standard input>", os.Stdin, os.Stdout, true, opts); err != nil {
			report(err)
		}
		return
	}

	for i := 0; i < flag.NArg(); i++ {
		path := flag.Arg(i)
		switch dir, err := os.Stat(path); {
		case err != nil:
			report(err)
		case dir.IsDir():
			opts, err := resolver.optionsForDir(path)
			if err != nil {
				report(err)
				continue
			}
			walkDir(path, opts)
		default:
			opts, err := resolver.optionsForFile(path)
			if err != nil {
				report(err)
				continue
			}
			if err := processFile(path, nil, os.Stdout, false, opts); err != nil {
				report(err)
			}
		}
	}
}

func diff(b1, b2 []byte) (data []byte, err error) {
	f1, err := ioutil.TempFile("", "asmfmt")
	if err != nil {
		return
	}
	defer os.Remove(f1.Name())
	defer f1.Close()

	f2, err := ioutil.TempFile("", "asmfmt")
	if err != nil {
		return
	}
	defer os.Remove(f2.Name())
	defer f2.Close()

	f1.Write(b1)
	f2.Write(b2)

	data, err = exec.Command("diff", "-u", f1.Name(), f2.Name()).CombinedOutput()
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}
	return

}
