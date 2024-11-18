package tcpxml

import "github.com/antchfx/xpath"

/*
Type

	@EnumConfig(marshal)
	@Enum {
		Int
		Uint
		Float
		String
	}
*/
type Type int

type XmlCommand struct {
	Name           string     `yaml:"name"`
	Description    string     `yaml:"description"`
	RequestFormat  string     `yaml:"requestFormat"`
	ResponseIf     string     `yaml:"responseIf"`
	Items          []*XmlItem `yaml:"items"`
	responseIfExpr *xpath.Expr
}

type XmlItem struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	XQuery      string `yaml:"xQuery"`
	Type        Type   `yaml:"type"`
	Scale       int    `yaml:"scale"`
	xQueryExpr  *xpath.Expr
}

type Client interface {
	Read(name string, params ...any) (map[string]any, error)
	parseLine(cmd XmlCommand, line string) (map[string]any, error)
}
