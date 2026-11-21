package cage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestWindowsInitialDataDirectorySkipsAppData(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("packaged Windows storage selection")
	}
	_, root := newStorageFixture(t)
	other := filepath.Join(root, "other")
	if err := os.Mkdir(other, 0700); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"LOCALAPPDATA", "APPDATA", "TMP", "TEMP"} {
		t.Setenv(key, other)
	}
	appData := filepath.Join(root, "AppData", "Local")
	app := filepath.Join(appData, "Programs", "Cage")
	backend := filepath.Join(app, "resources", "cage.exe")
	writeStorageFixture(t, backend, []byte("backend"))
	writeStorageFixture(t, filepath.Join(app, "Cage.exe"), []byte("launcher"))
	writeStorageFixture(t, filepath.Join(app, "Cage.install-id"), []byte("io.github.lee-xeo.claude-cage"))
	if want := filepath.Join(app, "CageData"); InitialStorageDirectory(backend) != want {
		t.Fatalf("did not suggest valid install storage %q", want)
	}
	t.Setenv("LOCALAPPDATA", appData)
	if got := InitialStorageDirectory(backend); got != "" {
		t.Errorf("suggested storage inside AppData: %q", got)
	}

	alias := filepath.Join(root, "alias")
	if err := linkStorage(alias, appData); err != nil {
		t.Fatal(err)
	}
	aliasedBackend := filepath.Join(alias, "Programs", "Cage", "resources", "cage.exe")
	if got := InitialStorageDirectory(aliasedBackend); got != "" {
		t.Errorf("suggested storage through junction to AppData: %q", got)
	}
}

func TestWindowsInitialDataDirectoryAndApplicationBoundary(t *testing.T) {
	profile, root := newStorageFixture(t)
	app := filepath.Join(root, "Cage")
	backend := filepath.Join(app, "resources", "cage.exe")
	writeStorageFixture(t, backend, []byte("backend"))
	writeStorageFixture(t, filepath.Join(app, "Cage.exe"), []byte("launcher"))
	if got := initialStorageDirectory(backend, "windows"); got != "" {
		t.Fatalf("unmarked build default=%s", got)
	}
	writeStorageFixture(t, filepath.Join(app, "Cage.install-id"), []byte("io.github.lee-xeo.claude-cage"))
	data := filepath.Join(app, "CageData")
	if got := initialStorageDirectory(backend, "windows"); got != data {
		t.Fatalf("default=%s want=%s", got, data)
	}
	for _, platform := range []string{"darwin", "linux"} {
		if got := initialStorageDirectory(backend, platform); got != "" {
			t.Fatalf("%s default changed to %s", platform, got)
		}
	}
	if err := validateStorageTarget(profile, data, app, "windows"); err != nil {
		t.Fatalf("reserved Windows data folder rejected: %v", err)
	}
	for _, target := range []string{app, filepath.Join(app, "resources", "data"), filepath.Join(app, "OtherData")} {
		if err := validateStorageTarget(profile, target, app, "windows"); err == nil {
			t.Fatalf("program boundary opened for %s", target)
		}
	}
	if err := validateStorageTarget(profile, data, app, "darwin"); err == nil {
		t.Fatal("Mac application directory allowed")
	}
	if err := os.Remove(filepath.Join(app, "Cage.install-id")); err != nil {
		t.Fatal(err)
	}
	if err := validateStorageTarget(profile, data, app, "windows"); err == nil {
		t.Fatal("unmarked installation allowed data inside app")
	}
}
