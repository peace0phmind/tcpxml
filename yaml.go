package tcpxml

import (
	"gopkg.in/yaml.v3"
	"strings"
)

type Commands []XmlCommand

func NewCommandsFromYaml(yamlName string) (Commands, error) {
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

func (cmds Commands) FindItemByName(itemName string) *XmlItem {
	for _, cmd := range cmds {
		for _, item := range cmd.Items {
			if strings.EqualFold(item.Name, itemName) {
				return item
			}
		}
	}

	return nil
}
