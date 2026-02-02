package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
)

type Config struct {
	LogFile   string `json:"logfile"`
	Intercept struct {
		Enabled bool   `json:"enabled"`
		Address string `json:"address"`
	} `json:"intercept"`
}

// TODO: Implement config file by precedence.
func getConfigFile() (string, error) {
	// TODO: MCPSHIM_CONFIGFILE env.
	// TODO: CWD
	// User config dir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	_, err = os.Stat(homeDir + "/.config/mcpshim/config.json")
	if err == nil {
		return homeDir + "/.config/mcpshim/config.json", nil
	}

	return "", errors.New("no config file found")

}

func NewConfig() Config {
	// TODO - Not returning an error as this should just panic if a failed config exists.
	cfg := Config{
		LogFile: "mcp_shim.log",
	}
	cfg.Intercept.Enabled = false

	// Determine which config file to use.
	cfgFile, err := getConfigFile()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(cfgFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		err := json.Unmarshal(data, &cfg)
		if err != nil {
			slog.Error("failed to unmarshal config data", "error", err)
		}
	}

	//

	return cfg
}
