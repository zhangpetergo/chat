package logger

import (
	"runtime/debug"
	"strconv"
	"strings"
)

// BuildInfo logs information stored inside the Go binary.
func BuildInfo() {
	var values []any

	info, _ := debug.ReadBuildInfo()

	for _, s := range info.Settings {
		key := s.Key
		if quoteKey(key) {
			key = strconv.Quote(key)
		}

		value := s.Value
		if quoteValue(value) {
			value = strconv.Quote(value)
		}

		values = append(values, key, value)
	}

	values = append(values, "goversion", info.GoVersion)
	values = append(values, "modversion", info.Main.Version)

	Log.Infow("build info", values...)
}

// quoteKey reports whether key is required to be quoted.
func quoteKey(key string) bool {
	return len(key) == 0 || strings.ContainsAny(key, "= \t\r\n\"`")
}

// quoteValue reports whether value is required to be quoted.
func quoteValue(value string) bool {
	return strings.ContainsAny(value, " \t\r\n\"`")
}
