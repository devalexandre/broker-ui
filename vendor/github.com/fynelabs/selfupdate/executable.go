package selfupdate

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fynelabs/selfupdate/internal/osext"
)

var (
	exePath           string
	defaultOldExePath string
	exeErr            error
	once              sync.Once
)

// ExecutableRealPath returns the path to the original executable and an error if something went bad
func ExecutableRealPath() (string, error) {
	if loadPath() != nil {
		return "", exeErr
	}
	return exePath, nil
}

// ExecutableDefaultOldPath returns the path to the old executable and an error if something went bad
func ExecutableDefaultOldPath() (string, error) {
	if loadPath() != nil {
		return "", exeErr
	}
	return defaultOldExePath, nil
}

func loadPath() error {
	once.Do(func() {
		exePath, exeErr = getExecutableRealPath()
		if exeErr != nil {
			return
		}
		// get the directory the executable exists in
		updateDir := filepath.Dir(exePath)
		filename := filepath.Base(exePath)

		// get file path to the old executable
		defaultOldExePath = filepath.Join(updateDir, fmt.Sprintf(".%s.old", filename))
	})
	return exeErr
}

func lastModifiedExecutable() (time.Time, error) {
	exe, err := ExecutableRealPath()
	if err != nil {
		return time.Time{}, err
	}

	fi, err := os.Stat(exe)
	if err != nil {
		return time.Time{}, err
	}

	return fi.ModTime(), nil
}

func getExecutableRealPath() (string, error) {
	exe, err := osext.Executable()
	if err != nil {
		return "", err
	}

	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return "", err
	}

	return exe, nil
}
