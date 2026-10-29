package requestbeam

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestLocaleCatalogCompletenessAndAliases(t *testing.T) {
	want := []string{"en", "de", "tr", "it", "es", "fr", "pt-BR", "ja", "ko", "zh-CN", "zh-TW", "ru"}
	if !reflect.DeepEqual(SupportedLocales(), want) {
		t.Fatalf("locales=%v", SupportedLocales())
	}
	keys := []string{"missing_api_key", "invalid_base_url", "invalid_timeout", "invalid_path", "invalid_job_id", "invalid_poll_interval", "invalid_pagination", "invalid_response", "transport_error", "invalid_stream", "job_failed", "poll_timeout", "stream_error"}
	for _, locale := range want {
		t.Run(locale, func(t *testing.T) {
			for _, key := range keys {
				if strings.TrimSpace(messages[locale][key]) == "" {
					t.Errorf("missing translation %s", key)
				}
				if message(locale, key) != messages[locale][key] {
					t.Errorf("message lookup failed for %s", key)
				}
			}
		})
	}
	for input, want := range map[string]string{"ZH_cn": "zh-CN", " zh-TW ": "zh-TW", "cn": "zh-CN", "tw": "zh-TW", "pt": "pt-BR", "PT_br": "pt-BR", "unknown": "en", "": "en"} {
		if got := normalizeLocale(input); got != want {
			t.Errorf("normalize(%q)=%q want %q", input, got, want)
		}
	}
	locales := SupportedLocales()
	locales[0] = "mutated"
	if SupportedLocales()[0] != "en" {
		t.Error("SupportedLocales exposes mutable shared state")
	}
	t.Setenv("REQUESTBEAM_API_KEY", "")
	for _, locale := range want {
		_, err := NewClient("", WithLocale(locale))
		if err == nil || !strings.Contains(err.Error(), messages[locale]["missing_api_key"]) {
			t.Errorf("constructor not localized for %s: %v", locale, err)
		}
	}
}

func TestTranslatedREADMEAndRelativeLinks(t *testing.T) {
	linkPattern := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	goBlockPattern := regexp.MustCompile("(?s)```go\\s*\\n(.*?)\\n```")
	files := []string{"README.md"}
	for _, locale := range SupportedLocales() {
		if locale != "en" {
			files = append(files, filepath.Join("docs", "README."+locale+".md"))
		}
	}
	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			if !strings.Contains(text, "github.com/RequestBeamDev/golang_sdk") || !strings.Contains(text, "WithLocale(") {
				t.Error("localized installation/example missing")
			}
			blocks := goBlockPattern.FindAllStringSubmatch(text, -1)
			if len(blocks) == 0 {
				t.Error("missing Go example")
			}
			for index, block := range blocks {
				source := block[1]
				if !strings.HasPrefix(strings.TrimSpace(source), "package ") {
					source = "package example\nfunc example() {\n" + source + "\n}"
				}
				if _, err := parser.ParseFile(token.NewFileSet(), file, source, parser.AllErrors); err != nil {
					t.Errorf("invalid Go syntax in example %d: %v", index+1, err)
				}
			}
			for _, match := range linkPattern.FindAllStringSubmatch(text, -1) {
				target := strings.Split(match[1], "#")[0]
				if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
					continue
				}
				if _, err := os.Stat(filepath.Join(filepath.Dir(file), target)); err != nil {
					t.Errorf("broken relative link %q: %v", target, err)
				}
			}
		})
	}
}
