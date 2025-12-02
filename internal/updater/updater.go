package updater

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Updater interface {
	GetVersion(keyPath string) (string, error)
	SetVersion(keyPath, version string) error
}

func New(filePath string) (Updater, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		return NewJSONUpdater(filePath), nil
	case ".yaml", ".yml":
		return NewYAMLUpdater(filePath), nil
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

type JSONUpdater struct {
	filePath string
}

func NewJSONUpdater(path string) *JSONUpdater {
	return &JSONUpdater{filePath: path}
}

func (u *JSONUpdater) GetVersion(keyPath string) (string, error) {
	data, err := u.read()
	if err != nil {
		return "", err
	}

	value := getNestedValue(data, strings.Split(keyPath, "."))
	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("version key '%s' not found or not a string", keyPath)
}

func (u *JSONUpdater) SetVersion(keyPath, version string) error {
	data, err := u.read()
	if err != nil {
		return err
	}

	if err := setNestedValue(data, strings.Split(keyPath, "."), version); err != nil {
		return err
	}

	return u.write(data)
}

func (u *JSONUpdater) read() (map[string]interface{}, error) {
	file, err := os.ReadFile(u.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return data, nil
}

func (u *JSONUpdater) write(data map[string]interface{}) error {
	file, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Add newline at end
	file = append(file, '\n')

	if err := os.WriteFile(u.filePath, file, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

type YAMLUpdater struct {
	filePath string
}

func NewYAMLUpdater(path string) *YAMLUpdater {
	return &YAMLUpdater{filePath: path}
}

func (u *YAMLUpdater) GetVersion(keyPath string) (string, error) {
	data, err := u.read()
	if err != nil {
		return "", err
	}

	value := getNestedValue(data, strings.Split(keyPath, "."))
	if str, ok := value.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("version key '%s' not found or not a string", keyPath)
}

func (u *YAMLUpdater) SetVersion(keyPath, version string) error {
	data, err := u.read()
	if err != nil {
		return err
	}

	if err := setNestedValue(data, strings.Split(keyPath, "."), version); err != nil {
		return err
	}

	return u.write(data)
}

func (u *YAMLUpdater) read() (map[string]interface{}, error) {
	file, err := os.ReadFile(u.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return data, nil
}

func (u *YAMLUpdater) write(data map[string]interface{}) error {
	file, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(u.filePath, file, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func getNestedValue(data map[string]interface{}, keys []string) interface{} {
	if len(keys) == 0 {
		return nil
	}

	if len(keys) == 1 {
		return data[keys[0]]
	}

	if nested, ok := data[keys[0]].(map[string]interface{}); ok {
		return getNestedValue(nested, keys[1:])
	}

	return nil
}

func setNestedValue(data map[string]interface{}, keys []string, value string) error {
	if len(keys) == 0 {
		return fmt.Errorf("empty key path")
	}

	if len(keys) == 1 {
		data[keys[0]] = value
		return nil
	}

	if _, ok := data[keys[0]]; !ok {
		data[keys[0]] = make(map[string]interface{})
	}

	nested, ok := data[keys[0]].(map[string]interface{})
	if !ok {
		return fmt.Errorf("key '%s' is not a map", keys[0])
	}

	return setNestedValue(nested, keys[1:], value)
}
