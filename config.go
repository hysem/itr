package main

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Tenant        Person `yaml:"tenant"`
	Landlord      Person `yaml:"landlord"`
	FinancialYear uint16 `yaml:"financial_year"`
	Rent          uint64 `yaml:"rent"`
}

type Person struct {
	Name    string `yaml:"name"`
	PAN     string `yaml:"pan"`
	Address string `yaml:"address"`
}

func ParseConfig(filename string) (*Config, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open config file: %s", filename)
	}
	var config Config
	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, errors.Wrapf(err, "failed to read config file: %s", filename)
	}
	return &config, nil
}
