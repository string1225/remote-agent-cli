//go:build !windows

package i18n

import "os"

func installedLanguage() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if language := os.Getenv(key); language != "" {
			return Normalize(language)
		}
	}
	return "en"
}
