package tcpxml

import "fmt"

type client struct {
	transporter Transporter
	cmdMap      map[string]XmlCommand
}

func NewClient(transporter Transporter, commands []XmlCommand) (Client, error) {
	c := &client{
		transporter: transporter,
		cmdMap:      make(map[string]XmlCommand),
	}

	for _, cmd := range commands {
		if _, ok := c.cmdMap[cmd.Name]; ok {
			return nil, fmt.Errorf("duplicate command: %s", cmd.Name)
		} else {
			c.cmdMap[cmd.Name] = cmd
		}
	}

	return c, nil
}

func (c *client) Read(cmdName string, params ...any) (any, error) {
	if cmd, ok := c.cmdMap[cmdName]; ok {
		request := cmd.RequestFormat
		if len(params) > 0 {
			request = fmt.Sprintf(cmd.RequestFormat, params...)
		}

		_, err := c.transporter.Write([]byte(request))
		if err != nil {
			return nil, err
		}

		var buf [4096]byte
		l, err := c.transporter.Read(buf[:])
		if err != nil {
			return nil, err
		}

		return string(buf[:l]), nil
	} else {
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}
}