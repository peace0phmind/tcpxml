package tcpxml

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
	"time"
)

func TestClient_Read(t *testing.T) {

	data, err := os.ReadFile("commands.yaml")
	if err != nil {
		panic(err)
	}

	var commands []XmlCommand
	err = yaml.Unmarshal(data, &commands)
	if err != nil {
		panic(err)
	}

	transporter := NewTcpTransport("127.0.0.1:62937")

	cc, err := NewClient(transporter, commands)
	assert.NoError(t, err)

	err = transporter.Open()
	assert.NoError(t, err)

	defer transporter.Close()

	for _, cmd := range commands {
		v, err1 := cc.Read(cmd.Name)
		assert.NoError(t, err1)
		t.Logf("%+v", v)
		//time.Sleep(2 * time.Second)
	}

	for i := 0; i < 2000; i++ {
		var buf [4096]byte
		l, err2 := transporter.Read(buf[:])
		if err2 != nil {
			t.Logf("%+v", err2)
		} else {
			t.Logf("%+v", buf[:l])
		}

		time.Sleep(2 * time.Second)
	}
}
