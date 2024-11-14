package tcpxml

import (
	"errors"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/expgo/factory"
	"github.com/expgo/log"
	"math"
	"strings"
)

type client struct {
	log.InnerLog
	transporter Transporter
	cmdMap      map[string]XmlCommand
}

func NewClient(transporter Transporter, commands []XmlCommand) (Client, error) {
	c := factory.New[client]()
	c.transporter = transporter
	c.cmdMap = make(map[string]XmlCommand)

	for _, cmd := range commands {
		if _, ok := c.cmdMap[cmd.Name]; ok {
			return nil, fmt.Errorf("duplicate command: %s", cmd.Name)
		} else {
			ifExpr, err := xpath.Compile(cmd.ResponseIf)
			if err != nil {
				return nil, fmt.Errorf("cmd %s parse expr err: %v", cmd.Name, err)
			}
			cmd.ResponseIfExpr = ifExpr

			for _, item := range cmd.Items {
				expr, err1 := xpath.Compile(item.XQuery)
				if err1 != nil {
					return nil, fmt.Errorf("cmd %s value %s parse expr err: %v", cmd.Name, item.Name, err1)
				}
				item.XQueryExpr = expr
			}

			c.cmdMap[cmd.Name] = cmd
		}
	}

	return c, nil
}

func (c *client) Read(cmdName string, params ...any) error {
	if cmd, ok := c.cmdMap[cmdName]; ok {
		request := cmd.RequestFormat
		if len(params) > 0 {
			request = fmt.Sprintf(cmd.RequestFormat, params...)
		}

		_, err := c.transporter.Write([]byte(request))
		if err != nil {
			return err
		}

		return nil
	} else {
		return fmt.Errorf("unknown command: %s", cmdName)
	}
}

func (c *client) doLine(line string) (map[string]any, error) {
	if len(line) == 0 {
		return nil, errors.New("empty line")
	}

	doc, err := xmlquery.Parse(strings.NewReader(line))
	if err != nil {
		return nil, err
	}

	nav := xmlquery.CreateXPathNavigator(doc)

	for _, cmd := range c.cmdMap {
		ev := cmd.ResponseIfExpr.Evaluate(nav)
		if bv, ok := ev.(bool); ok {
			if bv {
				var ret = make(map[string]any)
				for _, item := range cmd.Items {
					v := item.XQueryExpr.Evaluate(nav)
					switch item.Type {
					case TypeInt, TypeUint, TypeFloat:
						fv := v.(float64)
						if !math.IsNaN(fv) {
							ret[item.Name] = fv
						}
					case TypeString:
						sv := v.(string)
						if len(sv) > 0 {
							ret[item.Name] = sv
						}
					}
				}

				return ret, nil
			}
		}
	}

	c.L.Infof("count't parse line: %s", line)
	return nil, nil
}
