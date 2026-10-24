package workspace

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// SafeRoot verifies that dirPath is a safe, permissible location for a Vivechak research workspace.
// It explicitly rejects:
// - Empty paths
// - Root filesystem paths ('/', 'C:\', etc.)
// - User home directory directly ($HOME / %USERPROFILE%)
// - OS system directories (C:\Windows, /usr/bin, /etc, etc.)
// - Application directories and bundles (Program Files, .app bundles, AppData\Local\Programs)
func SafeRoot(dirPath string) error {
	trimmed := strings.TrimSpace(dirPath)
	if trimmed == "" {
		return fmt.Errorf("E-ROOT-UNSAFE: workspace path cannot be empty")
	}

	abs, err := filepath.Abs(trimmed)
	if err != nil {
		return fmt.Errorf("E-ROOT-UNSAFE: resolving absolute path for %q: %w", dirPath, err)
	}
	cleanAbs := filepath.Clean(abs)

	// Check for root directory
	if cleanAbs == "/" || cleanAbs == "\\" {
		return fmt.Errorf("E-ROOT-UNSAFE: filesystem root %q cannot be used as workspace root", dirPath)
	}

	// Check for Windows drive roots (e.g. C:\, D:\, c:)
	vol := filepath.VolumeName(cleanAbs)
	if vol != "" && (cleanAbs == vol || cleanAbs == vol+"\\" || cleanAbs == vol+"/") {
		return fmt.Errorf("E-ROOT-UNSAFE: drive root %q cannot be used as workspace root", dirPath)
	}

	// Check against user home directory
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		cleanHome := filepath.Clean(home)
		if strings.EqualFold(cleanAbs, cleanHome) {
			return fmt.Errorf("E-ROOT-UNSAFE: user home directory %q cannot be used directly as workspace root; please use a project subfolder", dirPath)
		}
	}

	// Check system or application directory
	if IsSystemOrAppDir(cleanAbs) {
		return fmt.Errorf("E-ROOT-UNSAFE: path %q is located inside a protected system or application directory", dirPath)
	}

	// Unix root-level directories (/etc, /bin, /usr, /sys, /proc, /dev, /root, /home)
	if runtime.GOOS != "windows" {
		topLevel := cleanAbs
		parts := strings.Split(cleanAbs, string(filepath.Separator))
		if len(parts) >= 2 && parts[1] != "" {
			switch parts[1] {
			case "bin", "sbin", "usr", "etc", "sys", "proc", "dev", "root", "home", "var", "tmp":
				if len(parts) == 2 {
					return fmt.Errorf("E-ROOT-UNSAFE: system root directory %q cannot be used as workspace root", dirPath)
				}
			}
		}
		_ = topLevel
	}

	return nil
}

func isSystemOrAppDirNormalized(checkPath string) bool {
	isDirOrChild := func(p, target string) bool {
		return p == target || strings.HasPrefix(p, target+"/")
	}

	// Windows application and system paths
	if strings.Contains(checkPath, "/appdata/local/programs") ||
		isDirOrChild(checkPath, "/program files") ||
		isDirOrChild(checkPath, "/program files (x86)") ||
		isDirOrChild(checkPath, "/programdata") ||
		isDirOrChild(checkPath, "/windows") {
		return true
	}

	// macOS application bundles (*.app or *.app/Contents/...) and system directories
	if strings.Contains(checkPath, ".app/") || strings.HasSuffix(checkPath, ".app") ||
		isDirOrChild(checkPath, "/applications") ||
		isDirOrChild(checkPath, "/system") ||
		isDirOrChild(checkPath, "/library") {
		return true
	}

	// Linux / Unix system binary and config paths
	if isDirOrChild(checkPath, "/usr/bin") ||
		isDirOrChild(checkPath, "/usr/local/bin") ||
		isDirOrChild(checkPath, "/usr/sbin") ||
		isDirOrChild(checkPath, "/usr/lib") ||
		isDirOrChild(checkPath, "/bin") ||
		isDirOrChild(checkPath, "/sbin") ||
		isDirOrChild(checkPath, "/etc") {
		return true
	}

	// Bare /opt root is a system root, but subdirectories like /opt/my-app are valid project targets
	if checkPath == "/opt" || checkPath == "/opt/" {
		return true
	}

	return false
}

// IsSystemOrAppDir returns true if dirPath appears to be an operating system binary,
// program installation, or application directory.
func IsSystemOrAppDir(dirPath string) bool {
	if dirPath == "" {
		return false
	}
	normalized := strings.ToLower(strings.ReplaceAll(dirPath, "\\", "/"))
	normalized = path.Clean(normalized)

	if isSystemOrAppDirNormalized(normalized) {
		return true
	}

	// Strip Windows drive letter (e.g. "c:") so root-based paths match uniformly
	checkPath := normalized
	if len(checkPath) >= 2 && checkPath[1] == ':' && checkPath[0] >= 'a' && checkPath[0] <= 'z' {
		checkPath = checkPath[2:]
		if isSystemOrAppDirNormalized(checkPath) {
			return true
		}
	}

	if idx := strings.Index(normalized, ":/"); idx >= 1 {
		sub := normalized[idx+1:]
		if isSystemOrAppDirNormalized(sub) {
			return true
		}
	}

	return false
}
