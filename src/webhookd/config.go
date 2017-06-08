package main

import (
	"os"
	"encoding/json"
	"github.com/davecgh/go-spew/spew"
)

type MQConfig struct {
	Type string `json:"type"`
	Protocol string `json:"protocol"`
	Host string `json:"host"`
	Port int `json:"port"`
	User string `json:"user"`
	Password string `json:"password"`
	Exchange string `json:"exchange"`
}

type PostrunConfig struct {
	Path string `json:"path"`
}

type Config struct {
	Address string `json:"address"`
	Port int `json:"port"`
	MQ MQConfig `json:"mq"`
}

func loadConfig(file string) (config Config, err error) {
	configFile, err := os.Open(file)
    defer configFile.Close()
    if err != nil {
        return config, err
    }

    jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return config, err
	}

	spew.Dump(config)

    return config, nil
}

func validateConfig(c Config) (err error) {
	// if ! c {
	// 	return errors.New("validateConfig: Empty Configuration struct")
	// }

	if c.Address == "" {
		lg(1, "%s", "bind address not set in config, using 0.0.0.0")
		c.Address = "0.0.0.0"
	}

	if c.Port == 0 {
		lg(1, "%s", "port not set in config, using 8080")
		c.Port = 8080
	}

	// to be continued ...

	return nil
}
