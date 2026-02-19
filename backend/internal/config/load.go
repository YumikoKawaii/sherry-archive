package config

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/spf13/viper"
)

// Load builds configuration by layering:
//  1. Compiled-in defaults (loadDefault)
//  2. config.yaml in the working directory (optional)
//  3. Environment variables — nested keys use __ as separator
//     e.g. DB__HOST overrides db.host, JWT__ACCESS_SECRET overrides jwt.access_secret
func Load() (*Application, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// Map env vars: DB__HOST → db.host
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()

	// Seed viper with compiled defaults so all keys are registered
	// before MergeInConfig / env resolution.
	c := loadDefault()
	configBuffer, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	if err := viper.ReadConfig(bytes.NewBuffer(configBuffer)); err != nil {
		return nil, err
	}

	// Merge config.yaml on top of defaults (ignore if not found)
	_ = viper.MergeInConfig()

	if err := viper.Unmarshal(c); err != nil {
		return nil, err
	}
	return c, nil
}
