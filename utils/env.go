package utils

import (
	"net/url"
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

// DockerHost returns the IP/hostname to use for Kafka's advertised listeners.
// It mirrors testcontainers' own host resolution so the advertised address
// matches what container.Host() will return after startup.
func DockerHost() string {
	if v := os.Getenv("TESTCONTAINERS_HOST_OVERRIDE"); v != "" {
		return v
	}

	dh := os.Getenv("DOCKER_HOST")
	if dh == "" || strings.HasPrefix(dh, "unix://") || strings.HasPrefix(dh, "npipe://") {
		return "localhost"
	}

	if u, err := url.Parse(dh); err == nil && u.Hostname() != "" {
		return u.Hostname()
	}

	return "localhost"
}
