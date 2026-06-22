//go:build integration

// Package integration tests the generated Pulumi ZITADEL provider against
// a real Zitadel Cloud Pro instance using Pulumi's automation API with a
// local file backend.
//
// Prerequisites (same as zitadel-operator):
//   - JWT key stored in system keyring: service="zitadel-operator", username="jwt-key"
//     Store:  secret-tool store --label='zitadel-operator jwt-key' service zitadel-operator username jwt-key < /path/to/key.json
//   - Config at ~/.config/zitadel-operator/config.yaml with domain, port, insecure
//   - Provider binary built: just provider
//
// Run: go test -tags=integration -v ./tests/integration/... -count=1 -timeout=120s
package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

// testConfig holds Zitadel connection parameters loaded from the operator config.
type testConfig struct {
	Domain                string `yaml:"domain"`
	Port                  string `yaml:"port"`
	Insecure              bool   `yaml:"insecure"`
	DefaultOrganizationId string `yaml:"defaultOrganizationId"`
}

var (
	cfg    testConfig
	jwtKey string
)

func TestMain(m *testing.M) {
	// Load config from ~/.config/zitadel-operator/config.yaml (same as operator tests).
	home, err := os.UserHomeDir()
	if err != nil {
		panic("cannot determine home dir: " + err.Error())
	}

	configPath := filepath.Join(home, ".config", "zitadel-operator", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic("failed to read config " + configPath + ": " + err.Error())
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic("failed to parse config: " + err.Error())
	}

	if cfg.Port == "" {
		cfg.Port = "443"
	}

	// Load JWT key from system keyring (same keyring entry as zitadel-operator).
	jwtKey, err = keyring.Get("zitadel-operator", "jwt-key")
	if err != nil {
		panic("failed to read JWT key from keyring: " + err.Error() +
			"\nhint: secret-tool store --label='zitadel-operator jwt-key' service zitadel-operator username jwt-key < /path/to/key.json")
	}

	os.Exit(m.Run())
}
