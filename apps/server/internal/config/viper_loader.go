package config

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// setViperDefaults reads `default:` struct tags of T and registers them as Viper defaults.
func setViperDefaults[T any](v *viper.Viper) {
	var zero T
	t := reflect.TypeOf(zero)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envKey := field.Tag.Get("env")
		defaultVal := field.Tag.Get("default")
		if envKey != "" && defaultVal != "" {
			v.SetDefault(envKey, defaultVal)
		}
	}
}

// parseEnvFileToMap reads a KEY=VALUE .env file and returns a flat map.
// Lines starting with # and blank lines are ignored.
// Inline comments and surrounding quotes are stripped.
func parseEnvFileToMap(filePath string) (map[string]interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	result := make(map[string]interface{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if idx := strings.Index(value, "#"); idx != -1 {
			value = strings.TrimSpace(value[:idx])
		}
		value = strings.Trim(value, `"'`)
		result[key] = value
	}
	return result, scanner.Err()
}

// LoadWithViper loads config of type T using the provided Viper instance.
//
// Priority (highest → lowest):
//  1. CLI flags bound with v.BindPFlag before calling this function
//  2. OS environment variables
//  3. .env file (auto-discovered or from envFilePath)
//  4. `default:` struct tag values
//
// Unmarshal uses the `env:` struct tag as the field name so that existing
// env-tag-based config structs work without adding mapstructure tags.
func LoadWithViper[T any](v *viper.Viper, envFilePath string) (T, error) {
	var cfg T

	// 1. Register struct-tag defaults.
	setViperDefaults[T](v)

	// 2. Merge .env file values at the "config file" precedence level
	//    (below OS env vars, above defaults).
	for _, candidate := range findEnvFiles(envFilePath) {
		envMap, err := parseEnvFileToMap(candidate)
		if err == nil && len(envMap) > 0 {
			if mergeErr := v.MergeConfigMap(envMap); mergeErr != nil {
				return cfg, fmt.Errorf("failed to merge .env file %s: %w", candidate, mergeErr)
			}
			break // first found wins
		}
	}

	// 3. OS environment variables override the .env file.
	v.AutomaticEnv()

	// 4. Unmarshal into T.
	//    - TagName "env" makes mapstructure read `env:"KEY"` as the field identifier.
	//    - MatchName is case-insensitive because Viper normalises all keys to lowercase.
	//    - WeaklyTypedInput converts string values (from env/file) to int, bool, etc.
	//    - StringToTimeDurationHookFunc handles "10s", "1m", "24h" → time.Duration.
	decodeHook := mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)
	err := v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "env"
		dc.WeaklyTypedInput = true
		dc.DecodeHook = decodeHook
		dc.MatchName = func(mapKey, fieldName string) bool {
			return strings.EqualFold(mapKey, fieldName)
		}
	})
	if err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}
