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

type XmlCommand struct {
	Name           string `yaml:"name"`
	Description    string `yaml:"description"`
	RequestFormat  string `yaml:"requestFormat"`
	ResponseXQuery string `yaml:"responseXQuery"`
	ResponseType   Type   `yaml:"responseType"`
}

type Client interface {
	Read(name string, params ...any) (any, error)
}
