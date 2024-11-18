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
	Name           string `yaml:"name"`
	Description    string `yaml:"description"`
	RequestFormat  string `yaml:"requestFormat"`
	ResponseIf     string `yaml:"responseIf"`
	ResponseIfExpr *xpath.Expr
	Items          []*XmlItem `yaml:"items"`
}

type XmlItem struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	XQuery      string `yaml:"xQuery"`
	XQueryExpr  *xpath.Expr
	Type        Type `yaml:"type"`
}

type Client interface {
	Read(name string, params ...any) (map[string]any, error)
	doLine(cmd XmlCommand, line string) (map[string]any, error)
}
