package config

import (
	"encoding/json"
	"os"
	//	"github.com/davecgh/go-spew/spew"
	. "github.com/vision-it/webhookd/logging"
)

type MQConfig struct {
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Exchange string `json:"exchange"`
}

type HooksConfig struct {
	Github []struct {
		Route    string `json:"route"`
		Secret   string `json:"secret"`
		Exchange string `json:"exchange"`
	} `json:"github"`
	Travis []struct {
		Route    string `json:"route"`
		Exchange string `json:"exchange"`
	} `json:"travis"`
	Gitlab []struct {
		Route    string `json:"route"`
		Secret   string `json:"secret"`
		Exchange string `json:"exchange"`
	} `json:"gitlab"`
	Gitea []struct {
		Route    string `json:"route"`
		Secret   string `json:"secret"`
		Exchange string `json:"exchange"`
	} `json:"gitea"`
	Demo []struct {
		Route    string `json:"route"`
		Secret   string `json:"secret"`
		Exchange string `json:"exchange"`
	} `json:"demo"`
}

type Config struct {
	Address     string      `json:"address"`
	Port        int         `json:"port"`
	RoutePrefix string      `json:"route-prefix"`
	MQ          MQConfig    `json:"mq"`
	Hooks       HooksConfig `json:"hooks"`
}

func LoadConfig(file string) (config Config, err error) {
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

	//	spew.Dump(config)

	return config, nil
}

func ValidateConfig(c Config) (err error) {
	// if (Config{}) == c {
	// 	return errors.New("validateConfig: Empty Configuration struct")
	// }

	if c.Address == "" {
		Lg(1, "%s", "bind address not set in config, using 0.0.0.0")
		c.Address = "0.0.0.0"
	}

	if c.Port == 0 {
		Lg(1, "%s", "port not set in config, using 8080")
		c.Port = 8080
	}

	// to be continued ...

	return nil
}
