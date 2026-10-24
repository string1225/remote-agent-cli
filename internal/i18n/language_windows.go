package i18n

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func installedLanguage() string {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\RemoteAgent`, registry.QUERY_VALUE|registry.WOW64_64KEY)
	if err == nil {
		value, _, readErr := key.GetStringValue("Language")
		key.Close()
		if readErr == nil {
			return Normalize(value)
		}
	}
	languages, _ := windows.GetSystemPreferredUILanguages(windows.MUI_LANGUAGE_NAME)
	for _, language := range languages {
		if Normalize(language) == "zh" {
			return "zh"
		}
	}
	return "en"
}
