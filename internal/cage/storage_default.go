package cage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sylvinus/agent-vm/internal/paths"
)

// InitialStorageDirectory supplies a fresh packaged Windows installation's
// data location when it is suitable for the VM's Unix sockets.
func InitialStorageDirectory(exe string) string {
	if runtime.GOOS != "windows" {
		return ""
	}
	resolved, err := paths.Resolve(exe)
	if err != nil {
		return ""
	}
	directory := initialStorageDirectory(resolved, runtime.GOOS)
	if directory == "" {
		return ""
	}
	target, err := StoragePhysicalPath(directory)
	if err != nil {
		return ""
	}
	// AF_UNIX connections and stale socket cleanup can fail in these trees.
	for _, root := range []string{os.Getenv("LOCALAPPDATA"), os.Getenv("APPDATA"), os.TempDir()} {
		if root == "" {
			continue
		}
		if StorageContains(root, directory) || StorageContains(root, target) {
			return ""
		}
		physical, err := StoragePhysicalPath(root)
		if err == nil && StorageContains(physical, target) {
			return ""
		}
	}
	return directory
}

func initialStorageDirectory(exe, platform string) string {
	if platform != "windows" || !strings.EqualFold(filepath.Base(filepath.Dir(exe)), "resources") {
		return ""
	}
	return windowsApplicationDataDirectory(filepath.Dir(filepath.Dir(exe)), platform)
}

func windowsApplicationDataDirectory(app, platform string) string {
	if platform != "windows" {
		return ""
	}
	marker, err := os.ReadFile(filepath.Join(app, "Cage.install-id"))
	if err != nil || string(marker) != "io.github.lee-xeo.claude-cage" {
		return ""
	}
	if info, err := os.Stat(filepath.Join(app, "Cage.exe")); err != nil || !info.Mode().IsRegular() {
		return ""
	}
	return filepath.Join(app, "CageData")
}
