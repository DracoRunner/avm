package tooling

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	nodeToolName     = "node"
	nodeInstallRoot  = ".avm/tools"
	nodeDistBaseURL  = "https://nodejs.org/dist"
	metadataFileName = "meta.json"
)

var nodeDistURL = nodeDistBaseURL

type nodeProvider struct{}

type nodeInstallMetadata struct {
	Tool        string    `json:"tool"`
	Version     string    `json:"version"`
	Platform    string    `json:"platform"`
	Arch        string    `json:"arch"`
	Source      string    `json:"source"`
	InstalledAt time.Time `json:"installed_at"`
	NodeDistURL string    `json:"node_dist_url"`
	Archive     string    `json:"archive"`
	Checksum    string    `json:"checksum"`
}

func init() {
	RegisterProvider(&nodeProvider{})
}

func (n *nodeProvider) Name() string {
	return nodeToolName
}

func (n *nodeProvider) IsInstalled(tool string, version string) bool {
	if tool != nodeToolName {
		return false
	}

	version = normalizeVersion(version)
	execPath, err := n.ToolExecutablePath(tool, version)
	if err != nil {
		return false
	}
	_, err = os.Stat(execPath)
	return err == nil
}

func (n *nodeProvider) InstalledVersions(tool string) ([]string, error) {
	if tool != nodeToolName {
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	toolDir := filepath.Join(home, nodeInstallRoot, tool)
	entries, err := os.ReadDir(toolDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var versions []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		version := entry.Name()
		metadataPath := filepath.Join(toolDir, version, metadataFileName)
		raw, err := os.ReadFile(metadataPath)
		if err != nil {
			continue
		}
		var metadata nodeInstallMetadata
		if err := json.Unmarshal(raw, &metadata); err != nil {
			continue
		}
		if metadata.Tool != nodeToolName {
			continue
		}
		versions = append(versions, version)
	}

	return versions, nil
}

func (n *nodeProvider) Install(tool string, version string) error {
	if tool != nodeToolName {
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	normalizedVersion := normalizeVersion(version)
	if n.IsInstalled(tool, normalizedVersion) {
		return nil
	}

	platform, arch, ext, err := nodePlatform()
	if err != nil {
		return err
	}

	versionTag := toVersionTag(normalizedVersion)
	fileName := fmt.Sprintf("node-%s-%s-%s.%s", versionTag, platform, arch, ext)
	archiveURL := fmt.Sprintf("%s/%s/%s", nodeDistURL, versionTag, fileName)

	tmpDir, err := os.MkdirTemp("", "avm-node-install-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, fileName)
	if err := downloadFile(archiveURL, archivePath); err != nil {
		return err
	}

	expectedChecksum, err := nodeChecksum(versionTag, fileName)
	if err != nil {
		return err
	}

	actualChecksum, err := fileSHA256(archivePath)
	if err != nil {
		return err
	}

	if expectedChecksum != "" && !strings.EqualFold(expectedChecksum, actualChecksum) {
		return fmt.Errorf("checksum mismatch for %s", fileName)
	}

	stagingDir := filepath.Join(tmpDir, "staging")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return err
	}

	if err := extractNodeArchive(archivePath, stagingDir, fileName, ext); err != nil {
		return err
	}

	installDir := nodeInstallDir(normalizedVersion)
	distPrefix := fmt.Sprintf("node-%s-%s-%s", versionTag, platform, arch)
	extractedRoot := filepath.Join(stagingDir, distPrefix)
	extractedRoot = strings.TrimSuffix(extractedRoot, fmt.Sprintf(".%s", ext))
	if _, err := os.Stat(extractedRoot); err != nil {
		return fmt.Errorf("unable to locate extracted archive root at %s", extractedRoot)
	}

	if err := os.RemoveAll(installDir); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(installDir), 0755); err != nil {
		return err
	}

	if err := os.Rename(extractedRoot, installDir); err != nil {
		return err
	}

	metadata := nodeInstallMetadata{
		Tool:        nodeToolName,
		Version:     normalizedVersion,
		Platform:    platform,
		Arch:        arch,
		Source:      archiveURL,
		InstalledAt: time.Now().UTC(),
		NodeDistURL: nodeDistURL,
		Archive:     fileName,
		Checksum:    actualChecksum,
	}

	rawMetadata, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(installDir, metadataFileName), rawMetadata, 0644); err != nil {
		return err
	}

	return nil
}

func (n *nodeProvider) Uninstall(tool string, version string) error {
	if tool != nodeToolName {
		return fmt.Errorf("unsupported tool: %s", tool)
	}

	version = normalizeVersion(version)
	installDir := nodeInstallDir(version)
	if _, err := os.Stat(installDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return os.RemoveAll(installDir)
}

func (n *nodeProvider) ToolExecutablePath(tool string, version string) (string, error) {
	if tool != nodeToolName {
		return "", fmt.Errorf("unsupported tool: %s", tool)
	}

	version = normalizeVersion(version)
	installDir := nodeInstallDir(version)
	executable := "node"
	if runtime.GOOS == "windows" {
		return filepath.Join(installDir, "node.exe"), nil
	}
	execPath := filepath.Join(installDir, "bin", executable)
	return execPath, nil
}

func nodeInstallDir(version string) string {
	normalized := normalizeVersion(version)
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".avm", "tools", nodeToolName, normalized)
	}
	return filepath.Join(home, ".avm", "tools", nodeToolName, normalized)
}

func nodePlatform() (string, string, string, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	var platform string
	switch goos {
	case "darwin":
		platform = "darwin"
	case "linux":
		platform = "linux"
	case "windows":
		platform = "win"
	default:
		return "", "", "", fmt.Errorf("unsupported platform: %s", goos)
	}

	var arch string
	switch goarch {
	case "amd64":
		arch = "x64"
	case "arm64":
		arch = "arm64"
	case "386":
		arch = "x86"
	default:
		return "", "", "", fmt.Errorf("unsupported architecture: %s", goarch)
	}

	ext := "tar.gz"
	if platform == "win" {
		ext = "zip"
	}
	return platform, arch, ext, nil
}

func toVersionTag(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

func downloadFile(url string, target string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func nodeChecksum(versionTag string, fileName string) (string, error) {
	checksumsURL := fmt.Sprintf("%s/%s/SHASUMS256.txt", nodeDistURL, versionTag)
	resp, err := http.Get(checksumsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("checksum request failed: %s", resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if fields[1] == fileName {
			return fields[0], nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("checksum entry not found for %s", fileName)
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func extractNodeArchive(archivePath string, stagingDir string, fileName string, ext string) error {
	switch ext {
	case "zip":
		return extractNodeZip(archivePath, stagingDir)
	case "tar.gz":
		_ = fileName
		return extractNodeTarGz(archivePath, stagingDir)
	default:
		return fmt.Errorf("unsupported archive extension: %s", ext)
	}
}

func extractNodeTarGz(archivePath, stagingDir string) error {
	in, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer in.Close()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(stagingDir, header.Name)
		target = filepath.Clean(target)

		if !strings.HasPrefix(target, filepath.Clean(stagingDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid tar path: %q", header.Name)
		}

		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}

		flags := os.O_CREATE | os.O_RDWR | os.O_TRUNC
		outFile, err := os.OpenFile(target, flags, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, tr); err != nil {
			outFile.Close()
			return err
		}
		outFile.Close()
	}

	return nil
}

func extractNodeZip(archivePath, stagingDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		targetPath := filepath.Join(stagingDir, f.Name)
		targetPath = filepath.Clean(targetPath)
		if !strings.HasPrefix(targetPath, filepath.Clean(stagingDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid zip path: %q", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		in, err := f.Open()
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		_, err = io.Copy(&buf, in)
		in.Close()
		if err != nil {
			return err
		}

		if err := os.WriteFile(targetPath, buf.Bytes(), f.FileInfo().Mode()); err != nil {
			return err
		}
	}

	return nil
}
