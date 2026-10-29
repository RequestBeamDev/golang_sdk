package requestbeam

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed locales/messages.json
var localeData []byte

var supportedLocales = []string{"en", "de", "tr", "it", "es", "fr", "pt-BR", "ja", "ko", "zh-CN", "zh-TW", "ru"}

var messages = func() map[string]map[string]string {
	var catalog map[string]map[string]string
	if err := json.Unmarshal(localeData, &catalog); err != nil {
		panic("requestbeam: invalid embedded locale catalog")
	}
	return catalog
}()

// SupportedLocales returns an independent copy of the supported locale codes.
func SupportedLocales() []string {
	return append([]string(nil), supportedLocales...)
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(locale), "_", "-"))
	switch locale {
	case "cn", "zh", "zh-hans":
		return "zh-CN"
	case "tw", "zh-hant":
		return "zh-TW"
	case "pt":
		return "pt-BR"
	}
	for _, code := range supportedLocales {
		if strings.ToLower(code) == locale {
			return code
		}
	}
	return "en"
}

func message(locale, key string) string {
	if value := messages[normalizeLocale(locale)][key]; value != "" {
		return value
	}
	if value := messages["en"][key]; value != "" {
		return value
	}
	return key
}
