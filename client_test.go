package tcpxml

import (
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"testing"
)

func TestClient_ReadCNC(t *testing.T) {
	data, err := os.ReadFile("cmd_cnc.yaml")
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

	defer transporter.Close()

	for _, cmd := range commands {
		err1 := cc.Read(cmd.Name)
		assert.NoError(t, err1)
	}

	for i := 0; i < 2000; i++ {
		line := <-transporter.Lines()
		t.Logf("%s", line)
	}

	//v, err1 := cc.Read("pos")
	//assert.NoError(t, err1)
	//t.Logf("%+v", v)

	//var buf [4096]byte
	//for i := 0; i < 20; i++ {
	//	l, err2 := transporter.Read(buf[:])
	//	if err2 != nil {
	//		t.Logf("%+v", err2)
	//	} else {
	//		t.Logf("%+v", string(buf[:l]))
	//	}
	//}
}

func TestClient_ReadPLC(t *testing.T) {
	data, err := os.ReadFile("cmd_plc.yaml")
	if err != nil {
		panic(err)
	}

	var commands []XmlCommand
	err = yaml.Unmarshal(data, &commands)
	if err != nil {
		panic(err)
	}

	transporter := NewTcpTransport("127.0.0.1:62944")

	cc, err := NewClient(transporter, commands)
	assert.NoError(t, err)

	defer transporter.Close()

	//for _, cmd := range commands {
	//	v, err1 := cc.Read(cmd.Name)
	//	assert.NoError(t, err1)
	//	t.Logf("%+v", v)
	//	time.Sleep(2 * time.Second)
	//}

	err1 := cc.Read("INI_GAS")
	assert.NoError(t, err1)
}

func doXmlEval(t *testing.T, nodeStr string, expStr string, expect any) {
	doc, err := xmlquery.Parse(strings.NewReader(nodeStr))
	assert.NoError(t, err)

	navigator := xmlquery.CreateXPathNavigator(doc)

	exp, err := xpath.Compile(expStr)
	assert.NoError(t, err)

	ni := exp.Evaluate(navigator)
	assert.Equal(t, expect, ni)
}

func TestClient_XPath(t *testing.T) {
	doXmlEval(t, "<ncda><auto>yes</auto><status>6</status></ncda>", "count(/ncda/status) > 0", true)
	doXmlEval(t, "<ncda><auto>yes</auto><status>6</status></ncda>", "number(/ncda/status/text())", 6.0)

	doXmlEval(t, "<ncda><amode>0</amode><auto>yes</auto><mmode>0</mmode><mode>0</mode></ncda>", "count(/ncda/amode | mmode | mode) > 0", true)
	doXmlEval(t, "<ncda><amode>0</amode><auto>yes</auto><mmode>0</mmode><mode>0</mode></ncda>", "number(/ncda/amode/text())", 0.0)

	doXmlEval(t, "<axes><auto>yes</auto><ax1>+04483.533</ax1><ax2>+00000.000</ax2><ax3>+00024.000</ax3><ax4>+04413.335</ax4><ax5>+04413.335</ax5><ax6>+00103.128</ax6><ax7>+00111.000</ax7><sub>pos</sub></axes>", `count(/axes/sub[text() = "pos"]) > 0`, true)
	doXmlEval(t, "<axes><auto>yes</auto><ax1>+04483.533</ax1><ax2>+00000.000</ax2><ax3>+00024.000</ax3><ax4>+04413.335</ax4><ax5>+04413.335</ax5><ax6>+00103.128</ax6><ax7>+00111.000</ax7><sub>pos</sub></axes>", "string(/axes/ax1/text())", "+04483.533")

	doXmlEval(t, "<axes><auto>yes</auto><ax1>0</ax1><ax2>0</ax2><ax3>0</ax3><ax4>14</ax4><ax5>14</ax5><ax6>0</ax6><ax7>0</ax7><sub>vel</sub></axes>", `count(/axes/sub[text() = "vel"]) > 0`, true)
	doXmlEval(t, "<axes><auto>yes</auto><ax1>0</ax1><ax2>0</ax2><ax3>0</ax3><ax4>14</ax4><ax5>14</ax5><ax6>0</ax6><ax7>0</ax7><sub>vel</sub></axes>", "number(/axes/ax1/text())", 0.0)

	doXmlEval(t, "<alarm><auto>yes</auto><no>821</no><prio>5</prio><st>nc1</st><v1>C:/PAData/NCProg\\1Flower.nc</v1><v2>C:/PAData/NCProg\\1Flower.nc</v2></alarm>", "count(/alarm) > 0 and number(/alarm/no/text()) > 0", true)
	doXmlEval(t, "<alarm><auto>yes</auto><no>821</no><prio>5</prio><st>nc1</st><v1>C:/PAData/NCProg\\1Flower.nc</v1><v2>C:/PAData/NCProg\\1Flower.nc</v2></alarm>", "string-join(/alarm/*[starts-with(name(), 'v')]/text(), ';')", "C:/PAData/NCProg\\1Flower.nc;C:/PAData/NCProg\\1Flower.nc")
}
