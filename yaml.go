package tcpxml

import (
	"gopkg.in/yaml.v3"
	"strings"
)

func NewCommandsFromYaml(yamlName string) ([]XmlCommand, error) {
	yamlName = strings.ToLower(yamlName) + ".yaml"

	buf, err := yamlFiles.ReadFile(yamlName)
	if err != nil {
		return nil, err
	}

	var commands []XmlCommand
	err = yaml.Unmarshal(buf, &commands)
	if err != nil {
		return nil, err
	}

	return commands, nil
}
