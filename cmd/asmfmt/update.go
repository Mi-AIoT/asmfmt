package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func runUpdate(target string) error {
	repo := os.Getenv("ASMFMT_UPGRADE_REPO")
	if repo == "" {
		repo = "Mi-AIoT/asmfmt"
	}
	apiBase := os.Getenv("ASMFMT_UPDATE_URL")
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	apiBase = strings.TrimSuffix(apiBase, "/")

	var urlPath string
	switch target {
	case "latest":
		urlPath = fmt.Sprintf("/repos/%s/releases/latest", repo)
	case "beta":
		urlPath = fmt.Sprintf("/repos/%s/releases/tags/beta", repo)
	default:
		urlPath = fmt.Sprintf("/repos/%s/releases/tags/%s", repo, target)
	}

	apiURL := apiBase + urlPath

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("User-Agent", "asmfmt-self-upgrade")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching release info from Github: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned non-OK status: %s (url: %s)", resp.Status, apiURL)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("decoding release JSON: %w", err)
	}

	// Match platform asset
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var extension string
	if goos == "windows" {
		extension = ".zip"
	} else {
		extension = ".tar.gz"
	}

	var downloadURL string
	var assetName string
	for _, asset := range release.Assets {
		nameLower := strings.ToLower(asset.Name)
		if !strings.HasSuffix(nameLower, extension) {
			continue
		}
		// Match os and arch
		if strings.Contains(nameLower, goos) && strings.Contains(nameLower, goarch) {
			downloadURL = asset.BrowserDownloadURL
			assetName = asset.Name
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no matching release asset found for OS=%s ARCH=%s (extension %s)", goos, goarch, extension)
	}

	fmt.Printf("Downloading %s from %s...\n", assetName, downloadURL)

	assetReq, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return fmt.Errorf("creating request for asset: %w", err)
	}
	assetReq.Header.Set("User-Agent", "asmfmt-self-upgrade")

	assetResp, err := client.Do(assetReq)
	if err != nil {
		return fmt.Errorf("downloading asset: %w", err)
	}
	defer assetResp.Body.Close()

	if assetResp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading asset failed with status: %s", assetResp.Status)
	}

	archiveData, err := io.ReadAll(assetResp.Body)
	if err != nil {
		return fmt.Errorf("reading asset body: %w", err)
	}

	var binaryBytes []byte
	binaryName := "asmfmt"
	if goos == "windows" {
		binaryName = "asmfmt.exe"
	}

	if extension == ".zip" {
		zipReader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
		if err != nil {
			return fmt.Errorf("reading zip archive: %w", err)
		}
		var found bool
		for _, file := range zipReader.File {
			// Compare base name
			if filepath.Base(file.Name) == binaryName {
				rc, err := file.Open()
				if err != nil {
					return fmt.Errorf("opening file inside zip: %w", err)
				}
				binaryBytes, err = io.ReadAll(rc)
				rc.Close()
				if err != nil {
					return fmt.Errorf("reading file from zip: %w", err)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("binary %s not found in zip archive", binaryName)
		}
	} else {
		// tar.gz
		gzipReader, err := gzip.NewReader(bytes.NewReader(archiveData))
		if err != nil {
			return fmt.Errorf("reading gzip: %w", err)
		}
		defer gzipReader.Close()

		tarReader := tar.NewReader(gzipReader)
		var found bool
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("reading tar header: %w", err)
			}
			if filepath.Base(header.Name) == binaryName {
				binaryBytes, err = io.ReadAll(tarReader)
				if err != nil {
					return fmt.Errorf("reading binary from tar: %w", err)
				}
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("binary %s not found in tar.gz archive", binaryName)
		}
	}

	// Safe Replacement
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting executable path: %w", err)
	}

	evalExecPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		// fallback to original if evaluation fails
		evalExecPath = execPath
	}

	execDir := filepath.Dir(evalExecPath)
	tmpFile, err := os.CreateTemp(execDir, "asmfmt-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temporary file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() {
		// Clean up the temp file if it wasn't successfully renamed
		if _, err := os.Stat(tmpPath); err == nil {
			os.Remove(tmpPath)
		}
	}()

	if _, err := tmpFile.Write(binaryBytes); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing to temporary file: %w", err)
	}

	if err := tmpFile.Chmod(0755); err != nil {
		tmpFile.Close()
		return fmt.Errorf("setting permissions on temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temporary file: %w", err)
	}

	if goos == "windows" {
		oldPath := evalExecPath + ".old"
		// If old file exists, try to delete it
		_ = os.Remove(oldPath)
		if err := os.Rename(evalExecPath, oldPath); err != nil {
			return fmt.Errorf("renaming running binary to .old: %w", err)
		}
		if err := os.Rename(tmpPath, evalExecPath); err != nil {
			// Roll back
			_ = os.Rename(oldPath, evalExecPath)
			return fmt.Errorf("moving new binary into place: %w", err)
		}
		fmt.Printf("Upgrade successful. Old binary renamed to %s\n", filepath.Base(oldPath))
	} else {
		// Non-Windows
		if err := os.Rename(tmpPath, evalExecPath); err != nil {
			return fmt.Errorf("moving new binary into place: %w", err)
		}
		fmt.Println("Upgrade successful.")
	}

	return nil
}
