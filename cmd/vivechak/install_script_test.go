package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallScriptChecksum(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not found on PATH, skipping shell script test")
	}

	installShPath := filepath.Join("..", "..", "install.sh")
	data, err := os.ReadFile(installShPath)
	if err != nil {
		t.Fatalf("failed to read install.sh: %v", err)
	}

	content := string(data)
	startMarker := "# >>> verify-checksum\n"
	endMarker := "# <<< verify-checksum"

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		t.Fatalf("start marker %q not found in install.sh", startMarker)
	}
	startIdx += len(startMarker)

	endIdx := strings.Index(content[startIdx:], endMarker)
	if endIdx == -1 {
		t.Fatalf("end marker %q not found in install.sh", endMarker)
	}
	endIdx += startIdx

	block := content[startIdx:endIdx]

	hashFile := func(path string) string {
		d, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read %s: %v", path, err)
		}
		h := sha256.Sum256(d)
		return hex.EncodeToString(h[:])
	}

	runCase := func(t *testing.T, archiveContent string, tampered bool, includeInChecksums bool) error {
		tmpDir := t.TempDir()
		archiveName := "vivechak_0.1.0_linux_amd64.tar.gz"
		archivePath := filepath.Join(tmpDir, archiveName)

		if err := os.WriteFile(archivePath, []byte(archiveContent), 0o644); err != nil {
			t.Fatal(err)
		}

		realHash := hashFile(archivePath)
		var checksumLines []string
		if includeInChecksums {
			checksumLines = append(checksumLines, realHash+"  "+archiveName)
		}
		checksumLines = append(checksumLines, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  "+archiveName+".cdx.json")
		checksumLines = append(checksumLines, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  vivechak_0.1.0_darwin_all.tar.gz")

		checksumsPath := filepath.Join(tmpDir, "checksums.txt")
		if err := os.WriteFile(checksumsPath, []byte(strings.Join(checksumLines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		if tampered {
			if err := os.WriteFile(archivePath, []byte(archiveContent+"-tampered"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		scriptFile := filepath.Join(tmpDir, "verify.sh")
		if err := os.WriteFile(scriptFile, []byte(block), 0o755); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command("sh", "-e", scriptFile)
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "ARCHIVE="+archiveName)
		return cmd.Run()
	}

	t.Run("ValidChecksumWithCDXAndOtherArchives", func(t *testing.T) {
		if err := runCase(t, "good-archive-content", false, true); err != nil {
			t.Fatalf("expected valid checksum to pass, got: %v", err)
		}
	})

	t.Run("TamperedArchiveFails", func(t *testing.T) {
		if err := runCase(t, "good-archive-content", true, true); err == nil {
			t.Fatal("expected tampered archive to fail verification, but it succeeded")
		}
	})

	t.Run("MissingFromChecksumsFails", func(t *testing.T) {
		if err := runCase(t, "good-archive-content", false, false); err == nil {
			t.Fatal("expected archive not in checksums.txt to fail verification, but it succeeded")
		}
	})
}
