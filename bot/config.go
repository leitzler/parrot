package bot

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Config is the representation of the dynamic config.json
type Config struct {
	Admin     string              `json:"admin"`
	Notifiers map[string][]string `json:"notifiers"`
	Debug     bool                `json:"debug"`
}

// NewConfig returns a new instance of Config
func NewConfig() *Config {
	cfg := &Config{
		Notifiers: make(map[string][]string),
	}
	return cfg
}

func (cfg *Config) fromJSON(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(file, &cfg); err != nil {
		return err
	}
	log.Println("Loaded config.")
	return nil
}

func (cfg *Config) toJSON(filename string) error {
	jsonBytes, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filename, jsonBytes, 0600); err != nil {
		return err
	}
	log.Println("Saved config.")
	return nil
}
