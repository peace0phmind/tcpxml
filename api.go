package tcpxml

/*
Type

	@EnumConfig(marshal)
	@Enum {
		Bool
		Int
		Uint
		String
	}
*/
type Type int

type XmlValue struct {
	Name   string `yaml:"name"`
	XQuery string `yaml:"xQuery"`
	Type   Type   `yaml:"type"`
}

type XmlCommand struct {
	Name          string     `yaml:"name"`
	Description   string     `yaml:"description"`
	RequestFormat string     `yaml:"requestFormat"`
	ResponseIf    string     `yaml:"responseIf"`
	Values        []XmlValue `yaml:"values"`
}

type Client interface {
	Read(name string, params ...any) error
}
