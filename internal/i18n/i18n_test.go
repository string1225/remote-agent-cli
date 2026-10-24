package i18n

import (
	"fmt"
	"strings"
	"testing"
)

func TestLanguageSelection(t *testing.T) {
	for _, language := range []string{"zh", "zh-CN", "zh-TW", "ZH_hant"} {
		if Normalize(language) != "zh" {
			t.Fatal(language)
		}
	}
	for _, language := range []string{"", "en-US", "de-DE", "ja-JP"} {
		if Normalize(language) != "en" {
			t.Fatal(language)
		}
	}
	t.Setenv("REMOTE_AGENT_LANGUAGE", "en")
	if Text("Paste Token and press Enter (input hidden): ") != "Paste Token and press Enter (input hidden): " {
		t.Fatal("English prompt changed")
	}
	t.Setenv("REMOTE_AGENT_LANGUAGE", "zh-CN")
	if !strings.Contains(Text("Previous configuration backed up to %s\n"), "备份") {
		t.Fatal("Chinese backup message missing")
	}
	if got := Error(fmt.Errorf("gateway verification returned HTTP %d; check token and server login", 401)); !strings.Contains(got, "401") || !strings.Contains(got, "检查") {
		t.Fatal(got)
	}
}

func TestCatalogPreservesFormatArguments(t *testing.T) {
	for en, zh := range messages {
		if strings.Join(formatPattern.FindAllString(en, -1), "|") != strings.Join(formatPattern.FindAllString(zh, -1), "|") {
			t.Errorf("format mismatch: %q -> %q", en, zh)
		}
	}
}
