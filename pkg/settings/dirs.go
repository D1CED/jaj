package settings

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Jguer/yay/v10/pkg/text"
)

// configFileName holds the name of the config file.
const configFileName = "config.json"

func getConfigPath() string {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		if err := initDir(configHome); err == nil {
			return filepath.Join(configHome, "yay", configFileName)
		}
	}

	if configHome := os.Getenv("HOME"); configHome != "" {
		if err := initDir(configHome); err == nil {
			return filepath.Join(configHome, ".config", "yay", configFileName)
		}
	}

	return ""
}

func getCacheHome() string {
	if cacheHome := os.Getenv("XDG_CACHE_HOME"); cacheHome != "" {
		if err := initDir(cacheHome); err == nil {
			return filepath.Join(cacheHome, "yay")
		}
	}

	if cacheHome := os.Getenv("HOME"); cacheHome != "" {
		if err := initDir(cacheHome); err == nil {
			return filepath.Join(cacheHome, ".cache", "yay")
		}
	}

	return "/tmp"
}

func initDir(dir string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return fmt.Errorf(text.Tf("failed to create config directory %q: %w", dir, err))
	}
	return nil
}
