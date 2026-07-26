package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	LocalSecretsFile = ".harness/secrets.local.env"
	HarnessGitignore = ".harness/.gitignore"
)

// LoadLocalSecrets loads per-project local secrets. The file is intended for
// developer machines only and is ignored by .harness/.gitignore.
func LoadLocalSecrets() map[string]string {
	secrets := map[string]string{}
	data, err := os.ReadFile(LocalSecretsFile)
	if err != nil {
		return secrets
	}
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if key != "" {
			secrets[key] = value
		}
	}
	return secrets
}

func SaveLocalSecret(key, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return fmt.Errorf("secret key cannot be empty")
	}
	if strings.ContainsAny(key, "\n\r\t =") {
		return fmt.Errorf("secret key contains invalid characters")
	}
	if strings.ContainsAny(value, "\n\r") {
		return fmt.Errorf("secret value cannot contain newlines")
	}

	secrets := LoadLocalSecrets()
	secrets[key] = value

	if err := os.MkdirAll(filepath.Dir(LocalSecretsFile), 0755); err != nil {
		return err
	}
	if err := ensureHarnessGitignoreIncludes("secrets.local.env"); err != nil {
		return err
	}

	var keys []string
	for existingKey := range secrets {
		keys = append(keys, existingKey)
	}
	sort.Strings(keys)

	var builder strings.Builder
	builder.WriteString("# Local machine secrets for Shipwright integrations.\n")
	builder.WriteString("# Do not commit this file.\n")
	for _, existingKey := range keys {
		builder.WriteString(existingKey)
		builder.WriteString("=")
		builder.WriteString(secrets[existingKey])
		builder.WriteString("\n")
	}
	return os.WriteFile(LocalSecretsFile, []byte(builder.String()), 0600)
}

func LocalSecretValue(key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return ""
	}
	return strings.TrimSpace(LoadLocalSecrets()[key])
}

func ensureHarnessGitignoreIncludes(entry string) error {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return nil
	}
	var existing string
	if data, err := os.ReadFile(HarnessGitignore); err == nil {
		existing = string(data)
	}
	for _, line := range strings.Split(existing, "\n") {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		existing += "\n"
	}
	existing += entry + "\n"
	return os.WriteFile(HarnessGitignore, []byte(existing), 0644)
}
