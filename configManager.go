package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	Database struct {
		//Username string `yaml:"username"`
		//Password string `yaml:"password"`
		ConnStr string `yaml:"connStr"`
		DbName  string `yaml:"dbName"`
	} `yaml:"database"`
}

func ParseConfigFile() *Config {
	f, err := os.Open(".config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var conf *Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}
	return conf
}
