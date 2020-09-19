package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

const (
	templatePath     = "/config/template/template.conf"
	configValuesPath = "/config/values/values.yaml"
	resultFilePath   = "/config/config.conf"
)

type values struct {
	LocalAddress string `yaml:"localaddress,omitempty"`
	SouthAddress string `yaml:"southaddress,omitempty"`
	NorthAddress string `yaml:"northaddress,omitempty"`
	UPFAddress   string `yaml:"upfaddress,omitempty"`
}

func main() {
	logrus.Info("Starting replacer")
	v := &values{}
	for v.LocalAddress == "" {
		if err := loadFile(v); err != nil {
			log.Fatalf("error loading file: %s", err)
		}
	}

	if err := replacer(v); err != nil {
		log.Fatalf("error replacing values: %s", err)
	}

	logrus.Info("Replacer finished!")
}

func loadFile(v *values) error {
	yamlFile, err := ioutil.ReadFile(configValuesPath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", configValuesPath, err)
	}

	err = yaml.Unmarshal(yamlFile, v)
	if err != nil {
		return fmt.Errorf("error unmarshaling file %s: %w", configValuesPath, err)
	}

	return nil
}

func replacer(v *values) error {
	// TODO: Use gardener struct instead of template file
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("error parsing template file: %w", err)
	}

	outputFile, err := os.Create(resultFilePath)
	if err != nil {
		return fmt.Errorf("error opening shoot file: %w", err)
	}

	defer outputFile.Close()

	err = tmpl.Execute(outputFile, v)
	if err != nil {
		return fmt.Errorf("error replacing shoot template files: %w", err)
	}

	return nil
}
