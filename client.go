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
			cmd.responseIfExpr = ifExpr

			for _, item := range cmd.Items {
				expr, err1 := xpath.Compile(item.XQuery)
				if err1 != nil {
					return nil, fmt.Errorf("cmd %s value %s parse expr err: %v", cmd.Name, item.Name, err1)
				}
				item.xQueryExpr = expr
			}

			c.cmdMap[cmd.Name] = cmd
		}
	}

	return c, nil
}

func (c *client) Read(cmdName string, params ...any) (map[string]any, error) {
	if cmd, ok := c.cmdMap[cmdName]; ok {
		request := cmd.RequestFormat
		if len(params) > 0 {
			request = fmt.Sprintf(cmd.RequestFormat, params...)
		}

		c.L.Debugf("Request: %s", request)

		_, err := c.transporter.Write([]byte(request))
		if err != nil {
			return nil, err
		}

		var buf [4096]byte
		l, err := c.transporter.Read(buf[:])
		if err != nil {
			return nil, err
		}

		response := string(buf[:l])

		c.L.Debugf("Response: %s", response)

		return c.parseLine(cmd, response)
	} else {
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}
}

func (c *client) parseLine(cmd XmlCommand, line string) (map[string]any, error) {
	if len(line) == 0 {
		return nil, errors.New("empty line")
	}

	if strings.Contains(line, "<.") {
		line = strings.ReplaceAll(line, "<.", "<_")
		line = strings.ReplaceAll(line, "</.", "</_")

		line = strings.ReplaceAll(line, "<_P[", "<_P_")
		line = strings.ReplaceAll(line, "</_P[", "</_P_")
		line = strings.ReplaceAll(line, "]>", "_>")
	}

	doc, err := xmlquery.Parse(strings.NewReader(line))
	if err != nil {
		return nil, err
	}

	nav := xmlquery.CreateXPathNavigator(doc)

	ev := cmd.responseIfExpr.Evaluate(nav)
	if bv, ok := ev.(bool); ok {
		if bv {
			var ret = make(map[string]any)

			for _, item := range cmd.Items {
				v := item.xQueryExpr.Evaluate(nav)
				switch item.Type {
				case TypeInt, TypeUint, TypeFloat:
					fv := v.(float64)
					if !math.IsNaN(fv) {
						switch item.Type {
						case TypeInt:
							ret[item.Name] = int32(fv)
						case TypeUint:
							ret[item.Name] = uint32(fv)
						case TypeFloat:
							ret[item.Name] = float32(fv)
						default:
							panic("unhandled default case")
						}
					}
				default:
					ret[item.Name] = v
				}
			}

			return ret, nil
		}
	}

	c.L.Infof("count't parse line: %s", line)
	return nil, nil
}
