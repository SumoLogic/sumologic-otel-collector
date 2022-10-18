package sumologic_scripts_tests

import (
	"os"

	"gopkg.in/yaml.v2"
)

type config struct {
	Extensions extensions `yaml:"extensions"`
}

type extensions struct {
	Sumologic sumologicExtension `yaml:"sumologic"`
}

type sumologicExtension struct {
	InstallToken string            `yaml:"install_token"`
	Tags         map[string]string `yaml:"collector_fields"`
	APIBaseURL   string            `yaml:"api_base_url"`
}

func getConfig(path string) (config, error) {
	var conf config

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		return config{}, err
	}

	return conf, err
}

func saveConfig(path string, conf config) error {
	out, err := yaml.Marshal(conf)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, out, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
