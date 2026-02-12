package utils

import (
	"os"
	"strings"
)

// EnvToLabels converts environment variables with the "TEST_LABEL_" prefix
// into a label map. The prefix is stripped and underscores are replaced with
// dots to form the label key; the env value becomes the label value.
func EnvToLabels() map[string]string {
	const prefix = "TEST_LABEL_"

	labels := make(map[string]string)

	for _, env := range os.Environ() {
		key, value, _ := strings.Cut(env, "=")
		if !strings.HasPrefix(key, prefix) {
			continue
		}

		labelKey := strings.ToLower(strings.ReplaceAll(key[len(prefix):], "_", "."))
		labels[labelKey] = value
	}

	return labels
}
