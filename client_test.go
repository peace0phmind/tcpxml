package tcpxml

import (
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	"math"
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
	if fv, ok := expect.(float64); ok {
		if math.IsNaN(fv) {
			assert.True(t, math.IsNaN(ni.(float64)))
		} else {
			assert.Equal(t, fv, ni)
		}
	} else {
		assert.Equal(t, expect, ni)
	}
}

func TestClient_XPath(t *testing.T) {
	doXmlEval(t, "<ncda><auto>yes</auto><status>6</status></ncda>", "count(/ncda/status) > 0", true)
	doXmlEval(t, "<ncda><auto>yes</auto><status>6</status></ncda>", "number(/ncda/status/text())", 6.0)

	doXmlEval(t, "<ncda><amode>0</amode><auto>yes</auto><mmode>0</mmode><mode>0</mode></ncda>", "count(/ncda/amode | mmode | mode) > 0", true)
	doXmlEval(t, "<ncda><amode>0</amode><auto>yes</auto><mmode>0</mmode><mode>0</mode></ncda>", "number(/ncda/amode/text())", 0.0)

	doXmlEval(t, "<axes><auto>yes</auto><ax1>+04483.533</ax1><ax2>+00000.000</ax2><ax3>+00024.000</ax3><ax4>+04413.335</ax4><ax5>+04413.335</ax5><ax6>+00103.128</ax6><ax7>+00111.000</ax7><sub>pos</sub></axes>", `count(/axes/sub[text() = "pos"]) > 0`, true)
	doXmlEval(t, "<axes><auto>yes</auto><ax1>+04483.533</ax1><ax2>+00000.000</ax2><ax3>+00024.000</ax3><ax4>+04413.335</ax4><ax5>+04413.335</ax5><ax6>+00103.128</ax6><ax7>+00111.000</ax7><sub>pos</sub></axes>", "string(/axes/ax1/text())", "+04483.533")
	doXmlEval(t, "<axes><auto>yes</auto><ax1>+04483.533</ax1><ax2>+00000.000</ax2><ax3>+00024.000</ax3><ax4>+04413.335</ax4><ax5>+04413.335</ax5><ax6>+00103.128</ax6><ax7>+00111.000</ax7><sub>pos</sub></axes>", "string(/axes/ax8/text())", "")

	doXmlEval(t, "<axes><auto>yes</auto><ax1>0</ax1><ax2>0</ax2><ax3>0</ax3><ax4>14</ax4><ax5>14</ax5><ax6>0</ax6><ax7>0</ax7><sub>vel</sub></axes>", `count(/axes/sub[text() = "vel"]) > 0`, true)
	doXmlEval(t, "<axes><auto>yes</auto><ax1>0</ax1><ax2>0</ax2><ax3>0</ax3><ax4>14</ax4><ax5>14</ax5><ax6>0</ax6><ax7>0</ax7><sub>vel</sub></axes>", "number(/axes/ax1/text())", 0.0)
	doXmlEval(t, "<axes><auto>yes</auto><ax1>0</ax1><ax2>0</ax2><ax3>0</ax3><ax4>14</ax4><ax5>14</ax5><ax6>0</ax6><ax7>0</ax7><sub>vel</sub></axes>", "number(/axes/ax8/text())", math.NaN())

	doXmlEval(t, "<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>", "count(/alarm) > 0 and number(/alarm/no/text()) > 0", false)
	doXmlEval(t, "<alarm><auto>yes</auto><no>821</no><prio>5</prio><st>nc1</st><v1>C:/PAData/NCProg\\1Flower.nc</v1><v2>C:/PAData/NCProg\\1Flower.nc</v2></alarm>", "count(/alarm) > 0 and number(/alarm/no/text()) > 0", true)
	doXmlEval(t, "<alarm><auto>yes</auto><no>821</no><prio>5</prio><st>nc1</st><v1>C:/PAData/NCProg\\1Flower.nc</v1><v2>C:/PAData/NCProg\\1Flower.nc</v2></alarm>", "string-join(/alarm/*[starts-with(name(), 'v')]/text(), ';')", "C:/PAData/NCProg\\1Flower.nc;C:/PAData/NCProg\\1Flower.nc")

	//doXmlEval(t, "<get><.P085>80</.P085><auto>yes</auto></get>", "count(/get) > 0", true)
	//doXmlEval(t, "<get><.P085>80</.P085><auto>yes</auto></get>", "", "")
}

func TestClient_DoLine(t *testing.T) {
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

	Lines := strings.Split(lines, "\n")

	for i, line := range Lines {
		values, err1 := cc.doLine(line)
		assert.NoError(t, err1)
		fmt.Printf("%d: %+v\n", i, values)
	}
}

const lines = `<ncda><amode>0</amode><auto>yes</auto><mmode>0</mmode><mode>0</mode></ncda>
<alarm><auto>yes</auto></alarm>
<axes><auto>yes</auto><ax1>+04483.533</ax1><ax2>+00000.000</ax2><ax3>+00024.000</ax3><ax4>+04413.335</ax4><ax5>+04413.335</ax5><ax6>+00103.128</ax6><ax7>+00111.000</ax7><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax2>0</ax2><ax3>0</ax3><ax4>14</ax4><ax5>14</ax5><ax6>0</ax6><ax7>0</ax7><sub>vel</sub></axes>
<blocks><act></act><auto>yes</auto><pas></pas><sub>basis</sub><temp></temp></blocks>
<laser><act1>9994</act1><act2>9995</act2><act3>0</act3><auto>yes</auto><preset1>9994</preset1><preset2>9995</preset2><preset3>0</preset3></laser>
<blocks><act>N17200 G01 X3534.4 Y0 Z24 A91.837</act><auto>yes</auto><pas>N17205 G01 X3534.4 Y0 Z24 A90</pas><sub>basis</sub><temp>N17210 G01 X3534.4 Y0 Z24 A88.163</temp></blocks>
<dir><eof>yes</eof><name>RO38_3 - ms014415.CNC</name><nmb>38</nmb><path>D:\NCPRog_\11-14</path><sub>exe</sub></dir>
<axes><auto>yes</auto><ax4>+04409.957</ax4><ax5>+04409.957</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04406.580</ax4><ax5>+04406.580</ax5><sub>pos</sub></axes>
<blocks><act>N17220 G01 X3534.4 Y0 Z24 A84.49</act><auto>yes</auto><pas>N17225 G01 X3534.4 Y0 Z24 A82.653</pas><sub>basis</sub><temp>N17230 G01 X3534.4 Y0 Z24 A80.816</temp></blocks>
<axes><auto>yes</auto><ax4>+04403.202</ax4><ax5>+04403.202</ax5><sub>pos</sub></axes>
<blocks><act>N17230 G01 X3534.4 Y0 Z24 A80.816</act><auto>yes</auto><pas>N17235 G01 X3534.4 Y0 Z24 A78.98</pas><sub>basis</sub><temp>N17240 G01 X3534.4 Y0 Z24 A77.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04399.825</ax4><ax5>+04399.825</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04396.869</ax4><ax5>+04396.869</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04393.492</ax4><ax5>+04393.492</ax5><sub>pos</sub></axes>
<blocks><act>N17255 G01 X3534.4 Y0 Z24 A71.633</act><auto>yes</auto><pas>N17260 G01 X3534.4 Y0 Z24 A69.796</pas><sub>basis</sub><temp>N17265 G01 X3534.4 Y0 Z24 A67.959</temp></blocks>
<axes><auto>yes</auto><ax4>+04390.114</ax4><ax5>+04390.114</ax5><sub>pos</sub></axes>
<blocks><act>N17265 G01 X3534.4 Y0 Z24 A67.959</act><auto>yes</auto><pas>N17270 G01 X3534.4 Y0 Z24 A66.122</pas><sub>basis</sub><temp>N17275 G01 X3534.4 Y0 Z24 A64.286</temp></blocks>
<axes><auto>yes</auto><ax4>+04386.737</ax4><ax5>+04386.737</ax5><sub>pos</sub></axes>
<blocks><act>N17275 G01 X3534.4 Y0 Z24 A64.286</act><auto>yes</auto><pas>N17280 G01 X3534.4 Y0 Z24 A62.449</pas><sub>basis</sub><temp>N17285 G01 X3534.4 Y0 Z24 A60.612</temp></blocks>
<axes><auto>yes</auto><ax4>+04383.359</ax4><ax5>+04383.359</ax5><sub>pos</sub></axes>
<blocks><act>N17285 G01 X3534.4 Y0 Z24 A60.612</act><auto>yes</auto><pas>N17290 G01 X3534.4 Y0 Z24 A58.775</pas><sub>basis</sub><temp>N17295 G01 X3534.4 Y0 Z24 A56.939</temp></blocks>
<axes><auto>yes</auto><ax4>+04380.404</ax4><ax5>+04380.404</ax5><sub>pos</sub></axes>
<blocks><act>N17290 G01 X3534.4 Y0 Z24 A58.775</act><auto>yes</auto><pas>N17295 G01 X3534.4 Y0 Z24 A56.939</pas><sub>basis</sub><temp>N17300 G01 X3534.4 Y0 Z24 A55.102</temp></blocks>
<axes><auto>yes</auto><ax4>+04377.026</ax4><ax5>+04377.026</ax5><sub>pos</sub></axes>
<blocks><act>N17300 G01 X3534.4 Y0 Z24 A55.102</act><auto>yes</auto><pas>N17305 G01 X3534.4 Y0 Z24 A53.265</pas><sub>basis</sub><temp>N17310 G01 X3534.4 Y0 Z24 A51.429</temp></blocks>
<axes><auto>yes</auto><ax4>+04373.649</ax4><ax5>+04373.649</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04370.271</ax4><ax5>+04370.271</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04366.894</ax4><ax5>+04366.894</ax5><sub>pos</sub></axes>
<blocks><act>N17330 G01 X3534.4 Y0 Z24 A44.082</act><auto>yes</auto><pas>N17335 G01 X3534.4 Y0 Z24 A42.245</pas><sub>basis</sub><temp>N17340 G01 X3534.4 Y0 Z24 A40.408</temp></blocks>
<axes><auto>yes</auto><ax4>+04363.516</ax4><ax5>+04363.516</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04360.561</ax4><ax5>+04360.561</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04357.183</ax4><ax5>+04357.183</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04353.806</ax4><ax5>+04353.806</ax5><sub>pos</sub></axes>
<blocks><act>N17365 G01 X3534.4 Y0 Z24 A31.225</act><auto>yes</auto><pas>N17370 G01 X3534.4 Y0 Z24 A29.388</pas><sub>basis</sub><temp>N17375 G01 X3534.4 Y0 Z24 A27.551</temp></blocks>
<axes><auto>yes</auto><ax4>+04350.428</ax4><ax5>+04350.428</ax5><sub>pos</sub></axes>
<blocks><act>N17375 G01 X3534.4 Y0 Z24 A27.551</act><auto>yes</auto><pas>N17380 G01 X3534.4 Y0 Z24 A25.714</pas><sub>basis</sub><temp>N17385 G01 X3534.4 Y0 Z24 A23.878</temp></blocks>
<axes><auto>yes</auto><ax4>+04347.473</ax4><ax5>+04347.473</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04344.095</ax4><ax5>+04344.095</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04340.718</ax4><ax5>+04340.718</ax5><sub>pos</sub></axes>
<blocks><act>N17400 G01 X3534.4 Y0 Z24 A18.367</act><auto>yes</auto><pas>N17405 G01 X3534.4 Y0 Z24 A16.531</pas><sub>basis</sub><temp>N17410 G01 X3534.4 Y0 Z24 A14.694</temp></blocks>
<axes><auto>yes</auto><ax4>+04337.340</ax4><ax5>+04337.340</ax5><sub>pos</sub></axes>
<blocks><act>N17410 G01 X3534.4 Y0 Z24 A14.694</act><auto>yes</auto><pas>N17415 G01 X3534.4 Y0 Z24 A13.243</pas><sub>basis</sub><temp>N17420 G01 X3534.4 Y0 Z24 A13.242</temp></blocks>
<axes><auto>yes</auto><ax4>+04334.080</ax4><ax5>+04334.080</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04333.017</ax4><ax5>+04333.017</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>6</ax4><ax5>6</ax5><sub>vel</sub></axes>
<blocks><act>N17425 G01 X3534.4 Y0 Z24 A11.02</act><auto>yes</auto><pas>N17430 G01 X3534.4 Y0 Z24 A9.184</pas><sub>basis</sub><temp>N17435 G01 X3534.4 Y0 Z24 A7.347</temp></blocks>
<axes><auto>yes</auto><ax4>+04330.342</ax4><ax5>+04330.342</ax5><sub>pos</sub></axes>
<blocks><act>N17435 G01 X3534.4 Y0 Z24 A7.347</act><auto>yes</auto><pas>N17440 G01 X3534.4 Y0 Z24 A5.51</pas><sub>basis</sub><temp>N17445 G01 X3534.4 Y0 Z24 A3.673</temp></blocks>
<axes><auto>yes</auto><ax4>+04326.964</ax4><ax5>+04326.964</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N17440 G01 X3534.4 Y0 Z24 A5.51</act><auto>yes</auto><pas>N17445 G01 X3534.4 Y0 Z24 A3.673</pas><sub>basis</sub><temp>N17450 G01 X3534.4 Y0 Z24 A1.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04323.587</ax4><ax5>+04323.587</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04320.212</ax4><ax5>+04320.212</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04318.165</ax4><ax5>+04318.165</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax4>+04318.163</ax4><ax5>+04318.163</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00022.079</ax3><sub>pos</sub></axes>
<blocks><act>N1100 M02</act><auto>yes</auto><pas>N17475 G00 Z29</pas><sub>basis</sub><temp>N17480 Q990039</temp></blocks>
<axes><auto>yes</auto><ax3>+00022.463</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>6</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00025.584</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00028.727</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>5</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00029.000</ax3><sub>pos</sub></axes>
<blocks><act>N1010 (GONGJIANXIALIAO_E6)</act><auto>yes</auto><pas>N1020 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<blocks><act>N1050 M223 (UD SP BOARD DOWN)</act><auto>yes</auto><pas>N1060 G4 F200</pas><sub>basis</sub><temp>N1070 G10</temp></blocks>
<blocks><act>N1060 G4 F200</act><auto>yes</auto><pas>N1070 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1070 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<blocks><act>N1160 (M254 LASER ON  CHECK)</act><auto>yes</auto><pas>N1170 M02</pas><sub>basis</sub><temp>N17490 (====PART 7 ====)</temp></blocks>
<blocks><act>N1020 M206</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<laser><act1>9363</act1><act2>9997</act2><auto>yes</auto><preset1>9363</preset1><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+04484.914</ax1><ax4>+04319.514</ax4><ax5>+04319.514</ax5><sub>pos</sub></axes>
<blocks><act>N17515 G00 X3536.6 Y0 A0</act><auto>yes</auto><pas>N17520 G10</pas><sub>basis</sub></blocks>
<laser><act1>9994</act1><act2>9998</act2><auto>yes</auto><preset1>9994</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+04485.733</ax1><ax4>+04320.000</ax4><ax5>+04320.000</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax4>+04320.443</ax4><ax5>+04320.443</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04321.806</ax4><ax5>+04321.806</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00028.208</ax3><ax4>+04321.837</ax4><ax5>+04321.837</ax5><sub>pos</sub></axes>
<blocks><act>N17530 G01 X3536.6 Y0 Z24 F10000</act><auto>yes</auto><pas>N17535 Q990051</pas><sub>basis</sub><temp>N1010 G10 (LCUT-1)</temp></blocks>
<axes><auto>yes</auto><ax3>+00024.884</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>10</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>*N1060 P117=P217,P118=P218,P121=P221,P122=P222,P123=P223*P191*P203</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><auto>yes</auto><pas>N1130 G110 X0 Y0 A0</pas><sub>basis</sub><temp>N1140 G111 V= P163 F100</temp></blocks>
<blocks><act>N1200 M110</act><auto>yes</auto><pas>N1210 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1010 G10 (STA-PO-PI-LI-P6018DA_A1)</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><sub>basis</sub><temp>N2260 M30</temp></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>N1020 U0 M118</temp></blocks>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<ncda><auto>yes</auto><status>4</status></ncda>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<blocks><act>N1090 M02</act><auto>yes</auto><pas>N17545 G01 X3539.6 Y0 Z24 A1.837 F3000</pas><sub>basis</sub><temp>N17550 G01 X3539.6 Y0 Z24 A3.673 F9047</temp></blocks>
<alarm><auto>yes</auto><no>1048</no><prio>4</prio><st>plc</st><v1> </v1></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<alarm><auto>yes</auto><no>1046</no><prio>4</prio><st>plc</st><v1> </v1></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<ncda><auto>yes</auto><status>6</status></ncda>
<axes><auto>yes</auto><ax1>+04486.208</ax1><sub>pos</sub></axes>
<blocks><act>N17545 G01 X3539.6 Y0 Z24 A1.837 F3000</act><auto>yes</auto><pas>N17550 G01 X3539.6 Y0 Z24 A3.673 F9047</pas><sub>basis</sub><temp>N17555 G01 X3539.6 Y0 Z24 A5.51</temp></blocks>
<laser><act1>9395</act1><act2>9998</act2><auto>yes</auto><preset1>9395</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+04487.328</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04488.448</ax1><sub>pos</sub></axes>
<laser><act1>7938</act1><act2>9996</act2><auto>yes</auto><preset1>7938</preset1><preset2>9996</preset2></laser>
<blocks><act>N17550 G01 X3539.6 Y0 Z24 A3.673 F9047</act><auto>yes</auto><pas>N17555 G01 X3539.6 Y0 Z24 A5.51</pas><sub>basis</sub><temp>N17560 G01 X3539.6 Y0 Z24 A7.347</temp></blocks>
<axes><auto>yes</auto><ax1>+04488.733</ax1><ax4>+04323.653</ax4><ax5>+04323.653</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N17555 G01 X3539.6 Y0 Z24 A5.51</act><auto>yes</auto><pas>N17560 G01 X3539.6 Y0 Z24 A7.347</pas><sub>basis</sub><temp>N17565 G01 X3539.6 Y0 Z24 A9.184</temp></blocks>
<laser><act1>9994</act1><act2>9995</act2><auto>yes</auto><preset1>9994</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax4>+04326.609</ax4><ax5>+04326.609</ax5><sub>pos</sub></axes>
<blocks><act>N17565 G01 X3539.6 Y0 Z24 A9.184</act><auto>yes</auto><pas>N17570 G01 X3539.6 Y0 Z24 A11.02</pas><sub>basis</sub><temp>N17575 G01 X3539.6 Y0 Z24 A12.857</temp></blocks>
<axes><auto>yes</auto><ax4>+04329.986</ax4><ax5>+04329.986</ax5><sub>pos</sub></axes>
<blocks><act>N17575 G01 X3539.6 Y0 Z24 A12.857</act><auto>yes</auto><pas>N17580 G01 X3539.6 Y0 Z24 A14.694</pas><sub>basis</sub><temp>N17585 G01 X3539.6 Y0 Z24 A16.915</temp></blocks>
<axes><auto>yes</auto><ax4>+04333.364</ax4><ax5>+04333.364</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04336.463</ax4><ax5>+04336.463</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>8</ax4><ax5>8</ax5><sub>vel</sub></axes>
<laser><act1>9462</act1><auto>yes</auto><preset1>9462</preset1></laser>
<axes><auto>yes</auto><ax4>+04337.167</ax4><ax5>+04337.167</ax5><sub>pos</sub></axes>
<blocks><act>N17595 G01 X3539.6 Y0 Z24 A18.367</act><auto>yes</auto><pas>N17600 G01 X3539.6 Y0 Z24 A20.204</pas><sub>basis</sub><temp>N17605 G01 X3539.6 Y0 Z24 A22.041</temp></blocks>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04340.712</ax4><ax5>+04340.712</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04343.667</ax4><ax5>+04343.667</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04347.045</ax4><ax5>+04347.045</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04350.422</ax4><ax5>+04350.422</ax5><sub>pos</sub></axes>
<blocks><act>N17635 G01 X3539.6 Y0 Z24 A33.061</act><auto>yes</auto><pas>N17640 G01 X3539.6 Y0 Z24 A34.898</pas><sub>basis</sub><temp>N17645 G01 X3539.6 Y0 Z24 A36.735</temp></blocks>
<axes><auto>yes</auto><ax4>+04353.800</ax4><ax5>+04353.800</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04356.755</ax4><ax5>+04356.755</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04360.132</ax4><ax5>+04360.132</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04363.510</ax4><ax5>+04363.510</ax5><sub>pos</sub></axes>
<blocks><act>N17670 G01 X3539.6 Y0 Z24 A45.918</act><auto>yes</auto><pas>N17675 G01 X3539.6 Y0 Z24 A47.755</pas><sub>basis</sub><temp>N17680 G01 X3539.6 Y0 Z24 A49.592</temp></blocks>
<axes><auto>yes</auto><ax4>+04366.888</ax4><ax5>+04366.888</ax5><sub>pos</sub></axes>
<blocks><act>N17680 G01 X3539.6 Y0 Z24 A49.592</act><auto>yes</auto><pas>N17685 G01 X3539.6 Y0 Z24 A51.429</pas><sub>basis</sub><temp>N17690 G01 X3539.6 Y0 Z24 A53.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04370.265</ax4><ax5>+04370.265</ax5><sub>pos</sub></axes>
<blocks><act>N17690 G01 X3539.6 Y0 Z24 A53.265</act><auto>yes</auto><pas>N17695 G01 X3539.6 Y0 Z24 A55.102</pas><sub>basis</sub><temp>N17700 G01 X3539.6 Y0 Z24 A56.939</temp></blocks>
<axes><auto>yes</auto><ax4>+04373.220</ax4><ax5>+04373.220</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04376.598</ax4><ax5>+04376.598</ax5><sub>pos</sub></axes>
<blocks><act>N17705 G01 X3539.6 Y0 Z24 A58.775</act><auto>yes</auto><pas>N17710 G01 X3539.6 Y0 Z24 A60.612</pas><sub>basis</sub><temp>N17715 G01 X3539.6 Y0 Z24 A62.449</temp></blocks>
<axes><auto>yes</auto><ax4>+04379.976</ax4><ax5>+04379.976</ax5><sub>pos</sub></axes>
<blocks><act>N17715 G01 X3539.6 Y0 Z24 A62.449</act><auto>yes</auto><pas>N17720 G01 X3539.6 Y0 Z24 A64.286</pas><sub>basis</sub><temp>N17725 G01 X3539.6 Y0 Z24 A66.122</temp></blocks>
<axes><auto>yes</auto><ax4>+04383.353</ax4><ax5>+04383.353</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04386.731</ax4><ax5>+04386.731</ax5><sub>pos</sub></axes>
<blocks><act>N17735 G01 X3539.6 Y0 Z24 A69.796</act><auto>yes</auto><pas>N17740 G01 X3539.6 Y0 Z24 A71.633</pas><sub>basis</sub><temp>N17745 G01 X3539.6 Y0 Z24 A73.469</temp></blocks>
<axes><auto>yes</auto><ax4>+04389.686</ax4><ax5>+04389.686</ax5><sub>pos</sub></axes>
<blocks><act>N17740 G01 X3539.6 Y0 Z24 A71.633</act><auto>yes</auto><pas>N17745 G01 X3539.6 Y0 Z24 A73.469</pas><sub>basis</sub><temp>N17750 G01 X3539.6 Y0 Z24 A75.306</temp></blocks>
<axes><auto>yes</auto><ax4>+04393.064</ax4><ax5>+04393.064</ax5><sub>pos</sub></axes>
<blocks><act>N17750 G01 X3539.6 Y0 Z24 A75.306</act><auto>yes</auto><pas>N17755 G01 X3539.6 Y0 Z24 A77.143</pas><sub>basis</sub><temp>N17760 G01 X3539.6 Y0 Z24 A78.98</temp></blocks>
<axes><auto>yes</auto><ax4>+04396.441</ax4><ax5>+04396.441</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04399.819</ax4><ax5>+04399.819</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04403.196</ax4><ax5>+04403.196</ax5><sub>pos</sub></axes>
<blocks><act>N17775 G01 X3539.6 Y0 Z24 A84.49</act><auto>yes</auto><pas>N17780 G01 X3539.6 Y0 Z24 A86.327</pas><sub>basis</sub><temp>N17785 G01 X3539.6 Y0 Z24 A88.163</temp></blocks>
<axes><auto>yes</auto><ax4>+04406.574</ax4><ax5>+04406.574</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04409.529</ax4><ax5>+04409.529</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04412.907</ax4><ax5>+04412.907</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04416.284</ax4><ax5>+04416.284</ax5><sub>pos</sub></axes>
<blocks><act>N17815 G01 X3539.6 Y0 Z24 A99.184</act><auto>yes</auto><pas>N17820 G01 X3539.6 Y0 Z24 A101.02</pas><sub>basis</sub><temp>N17825 G01 X3539.6 Y0 Z24 A102.857</temp></blocks>
<axes><auto>yes</auto><ax4>+04419.662</ax4><ax5>+04419.662</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04422.617</ax4><ax5>+04422.617</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04425.995</ax4><ax5>+04425.995</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04429.372</ax4><ax5>+04429.372</ax5><sub>pos</sub></axes>
<blocks><act>N17850 G01 X3539.6 Y0 Z24 A112.041</act><auto>yes</auto><pas>N17855 G01 X3539.6 Y0 Z24 A113.878</pas><sub>basis</sub><temp>N17860 G01 X3539.6 Y0 Z24 A115.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04432.750</ax4><ax5>+04432.750</ax5><sub>pos</sub></axes>
<blocks><act>N17860 G01 X3539.6 Y0 Z24 A115.714</act><auto>yes</auto><pas>N17865 G01 X3539.6 Y0 Z24 A117.551</pas><sub>basis</sub><temp>N17870 G01 X3539.6 Y0 Z24 A119.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04436.127</ax4><ax5>+04436.127</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04439.083</ax4><ax5>+04439.083</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04442.460</ax4><ax5>+04442.460</ax5><sub>pos</sub></axes>
<blocks><act>N17885 G01 X3539.6 Y0 Z24 A124.898</act><auto>yes</auto><pas>N17890 G01 X3539.6 Y0 Z24 A126.735</pas><sub>basis</sub><temp>N17895 G01 X3539.6 Y0 Z24 A128.571</temp></blocks>
<axes><auto>yes</auto><ax4>+04445.838</ax4><ax5>+04445.838</ax5><sub>pos</sub></axes>
<blocks><act>N17895 G01 X3539.6 Y0 Z24 A128.571</act><auto>yes</auto><pas>N17900 G01 X3539.6 Y0 Z24 A130.408</pas><sub>basis</sub><temp>N17905 G01 X3539.6 Y0 Z24 A132.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04449.215</ax4><ax5>+04449.215</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04452.593</ax4><ax5>+04452.593</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04455.548</ax4><ax5>+04455.548</ax5><sub>pos</sub></axes>
<blocks><act>N17920 G01 X3539.6 Y0 Z24 A137.755</act><auto>yes</auto><pas>N17925 G01 X3539.6 Y0 Z24 A139.592</pas><sub>basis</sub><temp>N17930 G01 X3539.6 Y0 Z24 A141.429</temp></blocks>
<axes><auto>yes</auto><ax4>+04459.348</ax4><ax5>+04459.348</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04462.303</ax4><ax5>+04462.303</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04465.681</ax4><ax5>+04465.681</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04469.058</ax4><ax5>+04469.058</ax5><sub>pos</sub></axes>
<blocks><act>N17955 G01 X3539.6 Y0 Z24 A150.612</act><auto>yes</auto><pas>N17960 G01 X3539.6 Y0 Z24 A152.449</pas><sub>basis</sub><temp>N17965 G01 X3539.6 Y0 Z24 A154.286</temp></blocks>
<axes><auto>yes</auto><ax4>+04472.436</ax4><ax5>+04472.436</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04475.391</ax4><ax5>+04475.391</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04478.769</ax4><ax5>+04478.769</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04482.146</ax4><ax5>+04482.146</ax5><sub>pos</sub></axes>
<blocks><act>N17990 G01 X3539.6 Y0 Z24 A163.469</act><auto>yes</auto><pas>N17995 G01 X3539.6 Y0 Z24 A165.306</pas><sub>basis</sub><temp>N18000 G01 X3539.6 Y0 Z24 A167.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04485.524</ax4><ax5>+04485.524</ax5><sub>pos</sub></axes>
<blocks><act>N18000 G01 X3539.6 Y0 Z24 A167.143</act><auto>yes</auto><pas>N18005 G01 X3539.6 Y0 Z24 A168.98</pas><sub>basis</sub><temp>N18010 G01 X3539.6 Y0 Z24 A170.816</temp></blocks>
<axes><auto>yes</auto><ax4>+04488.901</ax4><ax5>+04488.901</ax5><sub>pos</sub></axes>
<blocks><act>N18010 G01 X3539.6 Y0 Z24 A170.816</act><auto>yes</auto><pas>N18015 G01 X3539.6 Y0 Z24 A172.653</pas><sub>basis</sub><temp>N18020 G01 X3539.6 Y0 Z24 A174.49</temp></blocks>
<axes><auto>yes</auto><ax4>+04491.857</ax4><ax5>+04491.857</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04495.234</ax4><ax5>+04495.234</ax5><sub>pos</sub></axes>
<blocks><act>N18030 G01 X3539.6 Y0 Z24 A178.163</act><auto>yes</auto><pas>N18035 G01 X3539.6 Y0 Z24 A180</pas><sub>basis</sub><temp>N18040 G01 X3539.6 Y0 Z24 A181.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04498.612</ax4><ax5>+04498.612</ax5><sub>pos</sub></axes>
<blocks><act>N18035 G01 X3539.6 Y0 Z24 A180</act><auto>yes</auto><pas>N18040 G01 X3539.6 Y0 Z24 A181.837</pas><sub>basis</sub><temp>N18045 G01 X3539.6 Y0 Z24 A183.673</temp></blocks>
<axes><auto>yes</auto><ax4>+04501.989</ax4><ax5>+04501.989</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04505.367</ax4><ax5>+04505.367</ax5><sub>pos</sub></axes>
<blocks><act>N18055 G01 X3539.6 Y0 Z24 A187.347</act><auto>yes</auto><pas>N18060 G01 X3539.6 Y0 Z24 A189.184</pas><sub>basis</sub><temp>N18065 G01 X3539.6 Y0 Z24 A191.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04508.322</ax4><ax5>+04508.322</ax5><sub>pos</sub></axes>
<blocks><act>N18065 G01 X3539.6 Y0 Z24 A191.02</act><auto>yes</auto><pas>N18070 G01 X3539.6 Y0 Z24 A192.857</pas><sub>basis</sub><temp>N18075 G01 X3539.6 Y0 Z24 A194.694</temp></blocks>
<axes><auto>yes</auto><ax4>+04511.700</ax4><ax5>+04511.700</ax5><sub>pos</sub></axes>
<blocks><act>N18075 G01 X3539.6 Y0 Z24 A194.694</act><auto>yes</auto><pas>N18080 G01 X3539.6 Y0 Z24 A196.531</pas><sub>basis</sub><temp>N18085 G01 X3539.6 Y0 Z24 A198.367</temp></blocks>
<axes><auto>yes</auto><ax4>+04515.077</ax4><ax5>+04515.077</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04518.455</ax4><ax5>+04518.455</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04521.833</ax4><ax5>+04521.833</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04525.210</ax4><ax5>+04525.210</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04528.165</ax4><ax5>+04528.165</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04531.543</ax4><ax5>+04531.543</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04534.921</ax4><ax5>+04534.921</ax5><sub>pos</sub></axes>
<blocks><act>N18135 G01 X3539.6 Y0 Z24 A216.735</act><auto>yes</auto><pas>N18140 G01 X3539.6 Y0 Z24 A218.571</pas><sub>basis</sub><temp>N18145 G01 X3539.6 Y0 Z24 A220.408</temp></blocks>
<axes><auto>yes</auto><ax4>+04538.298</ax4><ax5>+04538.298</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04541.253</ax4><ax5>+04541.253</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04544.631</ax4><ax5>+04544.631</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04548.009</ax4><ax5>+04548.009</ax5><sub>pos</sub></axes>
<blocks><act>N18170 G01 X3539.6 Y0 Z24 A229.592</act><auto>yes</auto><pas>N18175 G01 X3539.6 Y0 Z24 A231.429</pas><sub>basis</sub><temp>N18180 G01 X3539.6 Y0 Z24 A233.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04551.386</ax4><ax5>+04551.386</ax5><sub>pos</sub></axes>
<blocks><act>N18180 G01 X3539.6 Y0 Z24 A233.265</act><auto>yes</auto><pas>N18185 G01 X3539.6 Y0 Z24 A235.102</pas><sub>basis</sub><temp>N18190 G01 X3539.6 Y0 Z24 A236.939</temp></blocks>
<axes><auto>yes</auto><ax4>+04554.764</ax4><ax5>+04554.764</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04558.141</ax4><ax5>+04558.141</ax5><sub>pos</sub></axes>
<blocks><act>N18200 G01 X3539.6 Y0 Z24 A240.612</act><auto>yes</auto><pas>N18205 G01 X3539.6 Y0 Z24 A242.449</pas><sub>basis</sub><temp>N18210 G01 X3539.6 Y0 Z24 A244.286</temp></blocks>
<axes><auto>yes</auto><ax4>+04561.097</ax4><ax5>+04561.097</ax5><sub>pos</sub></axes>
<blocks><act>N18205 G01 X3539.6 Y0 Z24 A242.449</act><auto>yes</auto><pas>N18210 G01 X3539.6 Y0 Z24 A244.286</pas><sub>basis</sub><temp>N18215 G01 X3539.6 Y0 Z24 A246.122</temp></blocks>
<axes><auto>yes</auto><ax4>+04564.474</ax4><ax5>+04564.474</ax5><sub>pos</sub></axes>
<blocks><act>N18215 G01 X3539.6 Y0 Z24 A246.122</act><auto>yes</auto><pas>N18220 G01 X3539.6 Y0 Z24 A247.959</pas><sub>basis</sub><temp>N18225 G01 X3539.6 Y0 Z24 A249.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04567.852</ax4><ax5>+04567.852</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04571.229</ax4><ax5>+04571.229</ax5><sub>pos</sub></axes>
<blocks><act>N18235 G01 X3539.6 Y0 Z24 A253.469</act><auto>yes</auto><pas>N18240 G01 X3539.6 Y0 Z24 A255.306</pas><sub>basis</sub><temp>N18245 G01 X3539.6 Y0 Z24 A257.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04574.607</ax4><ax5>+04574.607</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04577.562</ax4><ax5>+04577.562</ax5><sub>pos</sub></axes>
<blocks><act>N18250 G01 X3539.6 Y0 Z24 A258.98</act><auto>yes</auto><pas>N18255 G01 X3539.6 Y0 Z24 A260.816</pas><sub>basis</sub><temp>N18260 G01 X3539.6 Y0 Z24 A262.653</temp></blocks>
<axes><auto>yes</auto><ax4>+04580.940</ax4><ax5>+04580.940</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04584.317</ax4><ax5>+04584.317</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04587.695</ax4><ax5>+04587.695</ax5><sub>pos</sub></axes>
<blocks><act>N18280 G01 X3539.6 Y0 Z24 A270</act><auto>yes</auto><pas>N18285 G01 X3539.6 Y0 Z24 A271.837</pas><sub>basis</sub><temp>N18290 G01 X3539.6 Y0 Z24 A273.673</temp></blocks>
<axes><auto>yes</auto><ax4>+04591.072</ax4><ax5>+04591.072</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04594.028</ax4><ax5>+04594.028</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04597.405</ax4><ax5>+04597.405</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04600.783</ax4><ax5>+04600.783</ax5><sub>pos</sub></axes>
<blocks><act>N18315 G01 X3539.6 Y0 Z24 A282.857</act><auto>yes</auto><pas>N18320 G01 X3539.6 Y0 Z24 A284.694</pas><sub>basis</sub><temp>N18325 G01 X3539.6 Y0 Z24 A286.531</temp></blocks>
<axes><auto>yes</auto><ax4>+04604.160</ax4><ax5>+04604.160</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04607.538</ax4><ax5>+04607.538</ax5><sub>pos</sub></axes>
<blocks><act>N18335 G01 X3539.6 Y0 Z24 A290.204</act><auto>yes</auto><pas>N18340 G01 X3539.6 Y0 Z24 A292.041</pas><sub>basis</sub><temp>N18345 G01 X3539.6 Y0 Z24 A293.878</temp></blocks>
<axes><auto>yes</auto><ax4>+04610.493</ax4><ax5>+04610.493</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04613.871</ax4><ax5>+04613.871</ax5><sub>pos</sub></axes>
<blocks><act>N18350 G01 X3539.6 Y0 Z24 A295.714</act><auto>yes</auto><pas>N18355 G01 X3539.6 Y0 Z24 A297.551</pas><sub>basis</sub><temp>N18360 G01 X3539.6 Y0 Z24 A299.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04617.248</ax4><ax5>+04617.248</ax5><sub>pos</sub></axes>
<blocks><act>N18360 G01 X3539.6 Y0 Z24 A299.388</act><auto>yes</auto><pas>N18365 G01 X3539.6 Y0 Z24 A301.225</pas><sub>basis</sub><temp>N18370 G01 X3539.6 Y0 Z24 A303.061</temp></blocks>
<axes><auto>yes</auto><ax4>+04620.626</ax4><ax5>+04620.626</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04624.003</ax4><ax5>+04624.003</ax5><sub>pos</sub></axes>
<blocks><act>N18380 G01 X3539.6 Y0 Z24 A306.735</act><auto>yes</auto><pas>N18385 G01 X3539.6 Y0 Z24 A308.571</pas><sub>basis</sub><temp>N18390 G01 X3539.6 Y0 Z24 A310.408</temp></blocks>
<axes><auto>yes</auto><ax4>+04626.959</ax4><ax5>+04626.959</ax5><sub>pos</sub></axes>
<blocks><act>N18385 G01 X3539.6 Y0 Z24 A308.571</act><auto>yes</auto><pas>N18390 G01 X3539.6 Y0 Z24 A310.408</pas><sub>basis</sub><temp>N18395 G01 X3539.6 Y0 Z24 A312.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04630.336</ax4><ax5>+04630.336</ax5><sub>pos</sub></axes>
<blocks><act>N18395 G01 X3539.6 Y0 Z24 A312.245</act><auto>yes</auto><pas>N18400 G01 X3539.6 Y0 Z24 A314.082</pas><sub>basis</sub><temp>N18405 G01 X3539.6 Y0 Z24 A315.918</temp></blocks>
<axes><auto>yes</auto><ax4>+04633.714</ax4><ax5>+04633.714</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04637.091</ax4><ax5>+04637.091</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04640.469</ax4><ax5>+04640.469</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04643.424</ax4><ax5>+04643.424</ax5><sub>pos</sub></axes>
<blocks><act>N18430 G01 X3539.6 Y0 Z24 A325.102</act><auto>yes</auto><pas>N18435 G01 X3539.6 Y0 Z24 A326.939</pas><sub>basis</sub><temp>N18440 G01 X3539.6 Y0 Z24 A328.775</temp></blocks>
<axes><auto>yes</auto><ax4>+04646.802</ax4><ax5>+04646.802</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04650.179</ax4><ax5>+04650.179</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04653.557</ax4><ax5>+04653.557</ax5><sub>pos</sub></axes>
<blocks><act>N18460 G01 X3539.6 Y0 Z24 A336.122</act><auto>yes</auto><pas>N18465 G01 X3539.6 Y0 Z24 A337.959</pas><sub>basis</sub><temp>N18470 G01 X3539.6 Y0 Z24 A339.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04656.934</ax4><ax5>+04656.934</ax5><sub>pos</sub></axes>
<blocks><act>N18470 G01 X3539.6 Y0 Z24 A339.796</act><auto>yes</auto><pas>N18475 G01 X3539.6 Y0 Z24 A341.633</pas><sub>basis</sub><temp>N18480 G01 X3539.6 Y0 Z24 A343.469</temp></blocks>
<axes><auto>yes</auto><ax4>+04659.890</ax4><ax5>+04659.890</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04663.267</ax4><ax5>+04663.267</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04666.322</ax4><ax5>+04666.322</ax5><sub>pos</sub></axes>
<blocks><act>N18490 G01 X3539.6 Y0 Z24 A346.757</act><auto>yes</auto><pas>N18495 G01 X3539.6 Y0 Z24 A346.758</pas><sub>basis</sub><temp>N18500 G01 X3539.6 Y0 Z24 A348.98</temp></blocks>
<laser><act1>9406</act1><auto>yes</auto><preset1>9406</preset1></laser>
<axes><auto>yes</auto><ax4>+04667.310</ax4><ax5>+04667.310</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>10</ax4><ax5>10</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04670.618</ax4><ax5>+04670.618</ax5><sub>pos</sub></axes>
<blocks><act>N18510 G01 X3539.6 Y0 Z24 A352.653</act><auto>yes</auto><pas>N18515 G01 X3539.6 Y0 Z24 A354.49</pas><sub>basis</sub><temp>N18520 G01 X3539.6 Y0 Z24 A356.327</temp></blocks>
<axes><auto>yes</auto><ax4>+04673.995</ax4><ax5>+04673.995</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N18520 G01 X3539.6 Y0 Z24 A356.327</act><auto>yes</auto><pas>N18525 G01 X3539.6 Y0 Z24 A358.163</pas><sub>basis</sub><temp>N18530 G01 X3539.6 Y0 Z24 A360</temp></blocks>
<axes><auto>yes</auto><ax4>+04676.951</ax4><ax5>+04676.951</ax5><sub>pos</sub></axes>
<blocks><act>N18530 G01 X3539.6 Y0 Z24 A360</act><auto>yes</auto><pas>N18535 G01 X3539.6 Y0 Z24 A361.837</pas><sub>basis</sub><temp>N18545 Q= P11 (LASER OFF)</temp></blocks>
<axes><auto>yes</auto><ax4>+04680.311</ax4><ax5>+04680.311</ax5><sub>pos</sub></axes>
<blocks><act>N18535 G01 X3539.6 Y0 Z24 A361.837</act><auto>yes</auto><pas>N18545 Q= P11 (LASER OFF)</pas><sub>basis</sub><temp>N1010 G10 (END-LI-PO_W)</temp></blocks>
<axes><auto>yes</auto><ax4>+04681.837</ax4><ax5>+04681.837</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00022.016</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<blocks><act>N18090 G01 X3539.6 Y0 Z24 A200.204</act><auto>yes</auto><pas>N18090 G01 X3539.6 Y0 Z24 A200.204</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1080 G90</act><auto>yes</auto><pas>N1090 G10</pas><sub>basis</sub></blocks>
<axes><auto>yes</auto><ax3>+00022.976</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>10</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00027.117</ax3><sub>pos</sub></axes>
<blocks><act>N18550 G00 Z29</act><auto>yes</auto><pas>N18555 (====CONTOUR 2 ====)</pas><sub>basis</sub><temp>N18560 G00 X3832.291 Y0 A360</temp></blocks>
<axes><auto>yes</auto><ax3>+00028.997</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>1</ax3><sub>vel</sub></axes>
<laser><act1>8287</act1><act2>9997</act2><auto>yes</auto><preset1>8287</preset1><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+04489.910</ax1><ax3>+00029.000</ax3><ax4>+04681.827</ax4><ax5>+04681.827</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9999</act2><auto>yes</auto><preset1>9994</preset1><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+04495.871</ax1><ax4>+04681.785</ax4><ax5>+04681.785</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>15</ax1><ax3>0</ax3><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04506.851</ax1><ax4>+04681.712</ax4><ax5>+04681.712</ax5><sub>pos</sub></axes>
<blocks><act>N18560 G00 X3832.291 Y0 A360</act><auto>yes</auto><pas>N18565 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax1>+04522.851</ax1><ax4>+04681.608</ax4><ax5>+04681.608</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>34</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04540.969</ax1><ax4>+04681.491</ax4><ax5>+04681.491</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04566.380</ax1><ax4>+04681.328</ax4><ax5>+04681.328</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>52</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04596.812</ax1><ax4>+04681.133</ax4><ax5>+04681.133</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04632.223</ax1><ax4>+04680.907</ax4><ax5>+04680.907</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>69</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04667.938</ax1><ax4>+04680.686</ax4><ax5>+04680.686</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04698.839</ax1><ax4>+04680.496</ax4><ax5>+04680.496</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>53</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04721.857</ax1><ax4>+04680.355</ax4><ax5>+04680.355</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04743.569</ax1><ax4>+04680.223</ax4><ax5>+04680.223</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>36</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04760.381</ax1><ax4>+04680.121</ax4><ax5>+04680.121</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04772.292</ax1><ax4>+04680.050</ax4><ax5>+04680.050</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>18</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04779.303</ax1><ax4>+04680.010</ax4><ax5>+04680.010</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04781.424</ax1><ax4>+04680.000</ax4><ax5>+04680.000</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax4>+04679.734</ax4><ax5>+04679.734</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04677.076</ax4><ax5>+04677.076</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>17</ax4><ax5>17</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04671.582</ax4><ax5>+04671.582</ax5><sub>pos</sub></axes>
<blocks><act>N18570 G00 X3832.291 Y0 A270.506</act><auto>yes</auto><pas>N18575 G01 X3832.291 Y0 Z24 F10000</pas><sub>basis</sub><temp>N18580 Q= P10 (LASER ON)</temp></blocks>
<axes><auto>yes</auto><ax4>+04663.252</ax4><ax5>+04663.252</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>41</ax4><ax5>41</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04653.637</ax4><ax5>+04653.637</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04639.990</ax4><ax5>+04639.990</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>65</ax4><ax5>65</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04625.072</ax4><ax5>+04625.072</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04612.706</ax4><ax5>+04612.706</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>48</ax4><ax5>48</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04604.123</ax4><ax5>+04604.123</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04596.154</ax4><ax5>+04596.154</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>24</ax4><ax5>24</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04592.342</ax4><ax5>+04592.342</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04590.541</ax4><ax5>+04590.541</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>2</ax4><ax5>2</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<blocks><act>N18575 G01 X3832.291 Y0 Z24 F10000</act><auto>yes</auto><pas>N18580 Q= P10 (LASER ON)</pas><sub>basis</sub><temp>N1010 G10 (STA-PO-PI-LI-P6018DA_A1)</temp></blocks>
<axes><auto>yes</auto><ax3>+00028.208</ax3><ax4>+04590.506</ax4><ax5>+04590.506</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00024.884</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>10</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N2250 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N1020 U0 M118</pas><sub>basis</sub><temp>N1022 G10</temp></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N18585 G01 X3835.288 Y0 Z24 A270.881 F3021</act><auto>yes</auto><pas>N18590 G01 X3835.276 Y0 Z24 A271.762 F9040</pas><sub>basis</sub><temp>N18595 G01 X3835.202 Y0 Z24 A273.521 F8984</temp></blocks>
<laser><act1>9412</act1><act2>9998</act2><auto>yes</auto><preset1>9412</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+04782.388</ax1><ax4>+04590.644</ax4><ax5>+04590.644</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04783.507</ax1><ax4>+04590.784</ax4><ax5>+04590.784</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04784.421</ax1><ax4>+04590.881</ax4><ax5>+04590.881</ax5><sub>pos</sub></axes>
<laser><act1>9268</act1><act2>9995</act2><auto>yes</auto><preset1>9268</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+04784.365</ax1><ax4>+04593.228</ax4><ax5>+04593.228</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N18610 G01 X3834.706 Y0 Z24 A278.492 F8483</act><auto>yes</auto><pas>N18615 G01 X3834.454 Y0 Z24 A280.018 F8204</pas><sub>basis</sub><temp>N18620 G01 X3834.164 Y0 Z24 A281.441 F7866</temp></blocks>
<axes><auto>yes</auto><ax1>+04784.169</ax1><ax4>+04596.100</ax4><ax5>+04596.100</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+04783.781</ax1><ax4>+04599.222</ax4><ax5>+04599.222</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04783.223</ax1><ax4>+04602.079</ax4><ax5>+04602.079</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+04782.571</ax1><ax4>+04604.206</ax4><ax5>+04604.206</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>4</ax4><ax5>4</ax5><sub>vel</sub></axes>
<laser><act1>7782</act1><act2>9995</act2><auto>yes</auto><preset1>7782</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+04782.328</ax1><ax4>+04605.014</ax4><ax5>+04605.014</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9997</act2><auto>yes</auto><preset1>9994</preset1><preset2>9997</preset2></laser>
<blocks><act>N18650 G01 X3832.664 Y0 Z24 A285.984 F5900</act><auto>yes</auto><pas>N18655 G01 X3832.22 Y0 Z24 A286.771 F5263</pas><sub>basis</sub><temp>N18660 G01 X3831.757 Y0 Z24 A287.394 F4594</temp></blocks>
<axes><auto>yes</auto><ax1>+04781.547</ax1><ax4>+04606.638</ax4><ax5>+04606.638</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+04780.544</ax1><ax4>+04607.840</ax4><ax5>+04607.840</ax5><sub>pos</sub></axes>
<laser><act1>9715</act1><auto>yes</auto><preset1>9715</preset1></laser>
<axes><auto>yes</auto><ax1>+04779.463</ax1><ax4>+04608.191</ax4><ax5>+04608.191</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<blocks><act>N18685 G01 X3829.32 Y0 Z24 A287.845 F3383</act><auto>yes</auto><pas>N18690 G01 X3828.84 Y0 Z24 A287.39 F3941</pas><sub>basis</sub><temp>N18695 G01 X3828.379 Y0 Z24 A286.769 F4596</temp></blocks>
<laser><act1>9433</act1><auto>yes</auto><preset1>9433</preset1></laser>
<blocks><act>N18695 G01 X3828.379 Y0 Z24 A286.769 F4596</act><auto>yes</auto><pas>N18700 G01 X3827.939 Y0 Z24 A285.989 F5261</pas><sub>basis</sub><temp>N18705 G01 X3827.518 Y0 Z24 A285.051 F5899</temp></blocks>
<axes><auto>yes</auto><ax1>+04778.366</ax1><ax4>+04607.636</ax4><ax5>+04607.636</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+04777.466</ax1><ax4>+04606.472</ax4><ax5>+04606.472</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04776.449</ax1><ax4>+04604.209</ax4><ax5>+04604.209</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+04775.790</ax1><ax4>+04601.990</ax4><ax5>+04601.990</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<blocks><act>N18725 G01 X3826.146 Y0 Z24 A280.014 F7864</act><auto>yes</auto><pas>N18730 G01 X3825.894 Y0 Z24 A278.496 F8203</pas><sub>basis</sub><temp>N18735 G01 X3825.684 Y0 Z24 A276.901 F8483</temp></blocks>
<blocks><act>N18735 G01 X3825.684 Y0 Z24 A276.901 F8483</act><auto>yes</auto><pas>N18740 G01 X3825.518 Y0 Z24 A275.229 F8706</pas><sub>basis</sub><temp>N18745 G01 X3825.397 Y0 Z24 A273.512 F8872</temp></blocks>
<axes><auto>yes</auto><ax1>+04775.192</ax1><ax4>+04599.111</ax4><ax5>+04599.111</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+04774.765</ax1><ax4>+04595.972</ax4><ax5>+04595.972</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04774.513</ax1><ax4>+04592.674</ax4><ax5>+04592.674</ax5><sub>pos</sub></axes>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+04774.435</ax1><ax4>+04589.728</ax4><ax5>+04589.728</ax5><sub>pos</sub></axes>
<blocks><act>N18760 G01 X3825.324 Y0 Z24 A268.229</act><auto>yes</auto><pas>N18765 G01 X3825.397 Y0 Z24 A266.485 F8984</pas><sub>basis</sub><temp>N18770 G01 X3825.518 Y0 Z24 A264.766 F8872</temp></blocks>
<blocks><act>N18770 G01 X3825.518 Y0 Z24 A264.766 F8872</act><auto>yes</auto><pas>N18775 G01 X3825.685 Y0 Z24 A263.096 F8705</pas><sub>basis</sub><temp>N18780 G01 X3825.894 Y0 Z24 A261.503 F8482</temp></blocks>
<axes><auto>yes</auto><ax1>+04774.518</ax1><ax4>+04586.372</ax4><ax5>+04586.372</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04774.777</ax1><ax4>+04583.112</ax4><ax5>+04583.112</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+04775.208</ax1><ax4>+04580.044</ax4><ax5>+04580.044</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04775.724</ax1><ax4>+04577.597</ax4><ax5>+04577.597</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<blocks><act>N18800 G01 X3827.125 Y0 Z24 A256.028 F7008</act><auto>yes</auto><pas>N18805 G01 X3827.518 Y0 Z24 A254.95 F6484</pas><sub>basis</sub><temp>N18810 G01 X3827.938 Y0 Z24 A254.011 F5899</temp></blocks>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+04776.569</ax1><ax4>+04574.911</ax4><ax5>+04574.911</ax5><sub>pos</sub></axes>
<blocks><act>N18810 G01 X3827.938 Y0 Z24 A254.011 F5899</act><auto>yes</auto><pas>N18815 G01 X3828.38 Y0 Z24 A253.229 F5261</pas><sub>basis</sub><temp>N18820 G01 X3828.843 Y0 Z24 A252.607 F4592</temp></blocks>
<axes><auto>yes</auto><ax1>+04777.356</ax1><ax4>+04573.303</ax4><ax5>+04573.303</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>6</ax4><ax5>6</ax5><sub>vel</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+04778.364</ax1><ax4>+04572.136</ax4><ax5>+04572.136</ax5><sub>pos</sub></axes>
<laser><act1>9715</act1><auto>yes</auto><preset1>9715</preset1></laser>
<axes><auto>yes</auto><ax1>+04779.442</ax1><ax4>+04571.818</ax4><ax5>+04571.818</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<blocks><act>N18845 G01 X3831.283 Y0 Z24 A252.157 F3385</act><auto>yes</auto><pas>N18850 G01 X3831.758 Y0 Z24 A252.607 F3941</pas><sub>basis</sub><temp>N18855 G01 X3832.22 Y0 Z24 A253.23 F4595</temp></blocks>
<laser><act1>9433</act1><auto>yes</auto><preset1>9433</preset1></laser>
<blocks><act>N18855 G01 X3832.22 Y0 Z24 A253.23 F4595</act><auto>yes</auto><pas>N18860 G01 X3832.664 Y0 Z24 A254.016 F5263</pas><sub>basis</sub><temp>N18865 G01 X3833.082 Y0 Z24 A254.949 F5900</temp></blocks>
<axes><auto>yes</auto><ax1>+04780.538</ax1><ax4>+04572.399</ax4><ax5>+04572.399</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+04781.436</ax1><ax4>+04573.590</ax4><ax5>+04573.590</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04782.345</ax1><ax4>+04575.590</ax4><ax5>+04575.590</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+04782.874</ax1><ax4>+04576.917</ax4><ax5>+04576.917</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<blocks><act>N18890 G01 X3834.163 Y0 Z24 A258.556 F7467</act><auto>yes</auto><pas>N18895 G01 X3834.453 Y0 Z24 A259.979 F7865</pas><sub>basis</sub><temp>N18900 G01 X3834.706 Y0 Z24 A261.505 F8203</temp></blocks>
<laser><act1>9373</act1><act2>9996</act2><auto>yes</auto><preset1>9373</preset1><preset2>9996</preset2></laser>
<blocks><act>N18900 G01 X3834.706 Y0 Z24 A261.505 F8203</act><auto>yes</auto><pas>N18905 G01 X3834.917 Y0 Z24 A263.109 F8482</pas><sub>basis</sub><temp>N18910 G01 X3835.082 Y0 Z24 A264.768 F8706</temp></blocks>
<axes><auto>yes</auto><ax1>+04783.312</ax1><ax4>+04578.991</ax4><ax5>+04578.991</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+04783.790</ax1><ax4>+04581.585</ax4><ax5>+04581.585</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04784.216</ax1><ax4>+04585.184</ax4><ax5>+04585.184</ax5><sub>pos</sub></axes>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+04784.386</ax1><ax4>+04588.097</ax4><ax5>+04588.097</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N18925 G01 X3835.3 Y0 Z24 A270 F9040</act><auto>yes</auto><pas>N18930 G01 X3835.288 Y0 Z24 A270.881</pas><sub>basis</sub><temp>N18940 Q= P11 (LASER OFF)</temp></blocks>
<blocks><act>N1010 G10 (END-LI-PO_W)</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax1>+04784.425</ax1><ax4>+04590.753</ax4><ax5>+04590.753</ax5><sub>pos</sub></axes>
<laser><act1>7341</act1><auto>yes</auto><preset1>7341</preset1></laser>
<axes><auto>yes</auto><ax1>+04784.421</ax1><ax4>+04590.881</ax4><ax5>+04590.881</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00022.135</ax3><sub>pos</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><sub>basis</sub></blocks>
<blocks><auto>yes</auto><pas>N1100 M02</pas><sub>basis</sub><temp>N18945 G00 Z29</temp></blocks>
<axes><auto>yes</auto><ax3>+00022.327</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>4</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00025.640</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00028.753</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>5</ax3><sub>vel</sub></axes>
<blocks><act>N18955 G00 X4104.793 Y0 A270.328</act><auto>yes</auto><pas>N18960 G01 X4104.793 Y0 Z24 F10000</pas><sub>basis</sub><temp>N18965 Q= P10 (LASER ON)</temp></blocks>
<axes><auto>yes</auto><ax1>+04785.039</ax1><ax3>+00029.000</ax3><ax4>+04590.879</ax4><ax5>+04590.879</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9999</act2><auto>yes</auto><preset1>9994</preset1><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+04789.157</ax1><ax4>+04590.869</ax4><ax5>+04590.869</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>12</ax1><ax3>0</ax3><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04798.568</ax1><ax4>+04590.849</ax4><ax5>+04590.849</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04813.000</ax1><ax4>+04590.818</ax4><ax5>+04590.818</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>31</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04832.451</ax1><ax4>+04590.777</ax4><ax5>+04590.777</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04853.588</ax1><ax4>+04590.732</ax4><ax5>+04590.732</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>49</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04882.451</ax1><ax4>+04590.672</ax4><ax5>+04590.672</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04916.333</ax1><ax4>+04590.601</ax4><ax5>+04590.601</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>67</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04950.841</ax1><ax4>+04590.531</ax4><ax5>+04590.531</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+04976.969</ax1><ax4>+04590.479</ax4><ax5>+04590.479</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>52</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+04993.298</ax1><ax4>+04590.446</ax4><ax5>+04590.446</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05012.891</ax1><ax4>+04590.407</ax4><ax5>+04590.407</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>38</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05030.616</ax1><ax4>+04590.372</ax4><ax5>+04590.372</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05043.362</ax1><ax4>+04590.347</ax4><ax5>+04590.347</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>19</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05051.129</ax1><ax4>+04590.332</ax4><ax5>+04590.332</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05053.917</ax1><ax4>+04590.328</ax4><ax5>+04590.328</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05053.926</ax1><ax3>+00027.856</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00024.377</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax3>6</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<blocks><act>N1010 G10 ( RECTANGULAR TUBE DETECTION_20161226 )</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N1070 Q= P005</pas><sub>basis</sub><temp>N1010 G10 (PIE1-NORM)</temp></blocks>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N1040 M02</pas><sub>basis</sub><temp>N1080 Q= P006</temp></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N1030 G90 O= P113 U2020 F= P101 M02</pas><sub>basis</sub><temp>N1090 M02</temp></blocks>
<axes><auto>yes</auto><ax1>+05054.051</ax1><ax4>+04590.361</ax4><ax5>+04590.361</ax5><sub>pos</sub></axes>
<laser><act1>9412</act1><act2>9998</act2><auto>yes</auto><preset1>9412</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05055.030</ax1><ax4>+04590.484</ax4><ax5>+04590.484</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<blocks><act>N18970 G01 X4107.79 Y0 Z24 A270.705 F3021</act><auto>yes</auto><pas>N18975 G01 X4107.78 Y0 Z24 A271.411 F9040</pas><sub>basis</sub><temp>N18980 G01 X4107.722 Y0 Z24 A272.812 F8984</temp></blocks>
<axes><auto>yes</auto><ax1>+05056.149</ax1><ax4>+04590.625</ax4><ax5>+04590.625</ax5><sub>pos</sub></axes>
<blocks><act>N18975 G01 X4107.78 Y0 Z24 A271.411 F9040</act><auto>yes</auto><pas>N18980 G01 X4107.722 Y0 Z24 A272.812 F8984</pas><sub>basis</sub><temp>N18985 G01 X4107.627 Y0 Z24 A274.176 F8873</temp></blocks>
<axes><auto>yes</auto><ax1>+05056.923</ax1><ax4>+04590.738</ax4><ax5>+04590.738</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>2</ax4><ax5>2</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><act2>9995</act2><auto>yes</auto><preset1>9994</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05056.839</ax1><ax4>+04593.459</ax4><ax5>+04593.459</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05056.577</ax1><ax4>+04596.273</ax4><ax5>+04596.273</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<blocks><act>N18990 G01 X4107.494 Y0 Z24 A275.499 F8705</act><auto>yes</auto><pas>N18995 G01 X4107.326 Y0 Z24 A276.78 F8481</pas><sub>basis</sub><temp>N19000 G01 X4107.124 Y0 Z24 A277.995 F8198</temp></blocks>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05056.077</ax1><ax4>+04599.218</ax4><ax5>+04599.218</ax5><sub>pos</sub></axes>
<blocks><act>N19015 G01 X4106.338 Y0 Z24 A281.14 F6981</act><auto>yes</auto><pas>N19020 G01 X4106.025 Y0 Z24 A281.992 F6450</pas><sub>basis</sub><temp>N19025 G01 X4105.689 Y0 Z24 A282.731 F5858</temp></blocks>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<blocks><act>N19025 G01 X4105.689 Y0 Z24 A282.731 F5858</act><auto>yes</auto><pas>N19030 G01 X4105.41 Y0 Z24 A283.217 F5219</pas><sub>basis</sub><temp>N19035 G01 X4105.41 Y0 Z24 A283.217</temp></blocks>
<axes><auto>yes</auto><ax1>+05055.367</ax1><ax4>+04601.706</ax4><ax5>+04601.706</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>9</ax4><ax5>9</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05054.571</ax1><ax4>+04603.214</ax4><ax5>+04603.214</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05054.007</ax1><ax4>+04604.047</ax4><ax5>+04604.047</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>4</ax4><ax5>4</ax5><sub>vel</sub></axes>
<laser><act1>9703</act1><act2>9998</act2><auto>yes</auto><preset1>9703</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05052.934</ax1><ax4>+04604.453</ax4><ax5>+04604.453</ax5><sub>pos</sub></axes>
<laser><act1>9431</act1><auto>yes</auto><preset1>9431</preset1></laser>
<blocks><act>N19080 G01 X4102.265 Y0 Z24 A283.348 F4554</act><auto>yes</auto><pas>N19085 G01 X4101.912 Y0 Z24 A282.733 F5218</pas><sub>basis</sub><temp>N19090 G01 X4101.575 Y0 Z24 A281.992 F5858</temp></blocks>
<axes><auto>yes</auto><ax1>+05051.976</ax1><ax4>+04603.906</ax4><ax5>+04603.906</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05050.982</ax1><ax4>+04602.345</ax4><ax5>+04602.345</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05050.157</ax1><ax4>+04600.035</ax4><ax5>+04600.035</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05049.538</ax1><ax4>+04597.188</ax4><ax5>+04597.188</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05049.132</ax1><ax4>+04594.023</ax4><ax5>+04594.023</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<blocks><act>N19135 G01 X4099.819 Y0 Z24 A271.414 F8984</act><auto>yes</auto><pas>N19140 G01 X4099.8 Y0 Z24 A270 F9040</pas><sub>basis</sub><temp>N19145 G01 X4099.82 Y0 Z24 A268.585</temp></blocks>
<axes><auto>yes</auto><ax1>+05048.957</ax1><ax4>+04591.106</ax4><ax5>+04591.106</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05048.971</ax1><ax4>+04587.739</ax4><ax5>+04587.739</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05049.202</ax1><ax4>+04584.467</ax4><ax5>+04584.467</ax5><sub>pos</sub></axes>
<blocks><act>N19165 G01 X4100.275 Y0 Z24 A263.215 F8479</act><auto>yes</auto><pas>N19170 G01 X4100.477 Y0 Z24 A262.003 F8196</pas><sub>basis</sub><temp>N19175 G01 X4100.711 Y0 Z24 A260.862 F7852</temp></blocks>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<blocks><act>N19180 G01 X4100.973 Y0 Z24 A259.812 F7447</act><auto>yes</auto><pas>N19185 G01 X4101.263 Y0 Z24 A258.857 F6978</pas><sub>basis</sub><temp>N19190 G01 X4101.578 Y0 Z24 A258.002 F6447</temp></blocks>
<axes><auto>yes</auto><ax1>+05049.651</ax1><ax4>+04581.444</ax4><ax5>+04581.444</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>12</ax4><ax5>12</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05050.314</ax1><ax4>+04578.831</ax4><ax5>+04578.831</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05051.047</ax1><ax4>+04577.054</ax4><ax5>+04577.054</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05052.041</ax1><ax4>+04575.801</ax4><ax5>+04575.801</ax5><sub>pos</sub></axes>
<blocks><act>N19220 G01 X4103.801 Y0 Z24 A255.522 F3043</act><auto>yes</auto><pas>N19225 G01 X4104.196 Y0 Z24 A255.595 F3045</pas><sub>basis</sub><temp>N19230 G01 X4104.585 Y0 Z24 A255.81 F3370</temp></blocks>
<laser><act1>9702</act1><auto>yes</auto><preset1>9702</preset1></laser>
<blocks><act>N19230 G01 X4104.585 Y0 Z24 A255.81 F3370</act><auto>yes</auto><pas>N19235 G01 X4104.965 Y0 Z24 A256.163 F3913</pas><sub>basis</sub><temp>N19240 G01 X4105.336 Y0 Z24 A256.654 F4557</temp></blocks>
<axes><auto>yes</auto><ax1>+05053.118</ax1><ax4>+04575.582</ax4><ax5>+04575.582</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9705</act1><auto>yes</auto><preset1>9705</preset1></laser>
<axes><auto>yes</auto><ax1>+05054.195</ax1><ax4>+04576.461</ax4><ax5>+04576.461</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05054.933</ax1><ax4>+04577.529</ax4><ax5>+04577.529</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><sub>vel</sub></axes>
<laser><act1>7752</act1><act2>9995</act2><auto>yes</auto><preset1>7752</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05055.324</ax1><ax4>+04578.741</ax4><ax5>+04578.741</ax5><sub>pos</sub></axes>
<blocks><act>N19270 G01 X4106.63 Y0 Z24 A259.822 F6981</act><auto>yes</auto><pas>N19275 G01 X4106.891 Y0 Z24 A260.871 F7450</pas><sub>basis</sub><temp>N19280 G01 X4107.124 Y0 Z24 A262.006 F7855</temp></blocks>
<laser><act1>9994</act1><act2>9997</act2><auto>yes</auto><preset1>9994</preset1><preset2>9997</preset2></laser>
<blocks><act>N19285 G01 X4107.326 Y0 Z24 A263.223 F8199</act><auto>yes</auto><pas>N19290 G01 X4107.494 Y0 Z24 A264.501 F8481</pas><sub>basis</sub><temp>N19295 G01 X4107.627 Y0 Z24 A265.825 F8706</temp></blocks>
<axes><auto>yes</auto><ax1>+05056.049</ax1><ax4>+04581.352</ax4><ax5>+04581.352</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>12</ax4><ax5>12</ax5><sub>vel</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05056.560</ax1><ax4>+04584.378</ax4><ax5>+04584.378</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05056.858</ax1><ax4>+04587.651</ax4><ax5>+04587.651</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N19310 G01 X4107.8 Y0 Z24 A270 F9040</act><auto>yes</auto><pas>N19315 G01 X4107.79 Y0 Z24 A270.705</pas><sub>basis</sub><temp>N19325 Q= P11 (LASER OFF)</temp></blocks>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05056.931</ax1><ax4>+04590.402</ax4><ax5>+04590.402</ax5><sub>pos</sub></axes>
<blocks><act>N19315 G01 X4107.79 Y0 Z24 A270.705</act><auto>yes</auto><pas>N19325 Q= P11 (LASER OFF)</pas><sub>basis</sub><temp>N1010 G10 (END-LI-PO_W)</temp></blocks>
<laser><act1>8678</act1><auto>yes</auto><preset1>8678</preset1></laser>
<blocks><act>N1020 M116</act><auto>yes</auto><pas>N1030 U0000 O= P100</pas><sub>basis</sub><temp>N1040 G10</temp></blocks>
<axes><auto>yes</auto><ax1>+05056.923</ax1><ax4>+04590.705</ax4><ax5>+04590.705</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00022.152</ax3><sub>pos</sub></axes>
<blocks><act>N1090 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N19330 G00 Z29</act><auto>yes</auto><pas>N19335 Q990036</pas><sub>basis</sub><temp>N1010 (UNLOAD ONE ENTER)</temp></blocks>
<axes><auto>yes</auto><ax3>+00023.496</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00027.657</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>12</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00029.000</ax3><sub>pos</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>N1040 M222</temp></blocks>
<blocks><act>N1040 M222</act><auto>yes</auto><pas>N1050 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N1060 G4 F200</act><auto>yes</auto><pas>N1070 M02</pas><sub>basis</sub><temp>N19340 (====CONTOUR 4 ====)</temp></blocks>
<blocks><act>N19345 G00 X4124 Y0 A358.163</act><auto>yes</auto><pas>N19350 G01 X4124 Y0 Z24 F10000</pas><sub>basis</sub><temp>N19355 Q= P10 (LASER ON)</temp></blocks>
<laser><act1>8497</act1><auto>yes</auto><preset1>8497</preset1></laser>
<axes><auto>yes</auto><ax1>+05057.153</ax1><ax4>+04592.300</ax4><ax5>+04592.300</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9997</act2><auto>yes</auto><preset1>9994</preset1><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05057.909</ax1><ax4>+04596.731</ax4><ax5>+04596.731</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>25</ax4><ax5>25</ax5><sub>vel</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05059.190</ax1><ax4>+04603.997</ax4><ax5>+04603.997</ax5><sub>pos</sub></axes>
<laser><act2>9999</act2><auto>yes</auto><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+05060.742</ax1><ax4>+04612.682</ax4><ax5>+04612.682</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>3</ax1><ax4>48</ax4><ax5>48</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05063.009</ax1><ax4>+04625.265</ax4><ax5>+04625.265</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05065.750</ax1><ax4>+04640.152</ax4><ax5>+04640.152</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>5</ax1><ax4>62</ax4><ax5>62</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05068.230</ax1><ax4>+04653.190</ax4><ax5>+04653.190</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05070.204</ax1><ax4>+04663.500</ax4><ax5>+04663.500</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>3</ax1><ax4>39</ax4><ax5>39</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05071.673</ax1><ax4>+04671.081</ax4><ax5>+04671.081</ax5><sub>pos</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05072.543</ax1><ax4>+04675.476</ax4><ax5>+04675.476</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>16</ax4><ax5>16</ax5><sub>vel</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05073.063</ax1><ax4>+04677.941</ax4><ax5>+04677.941</ax5><sub>pos</sub></axes>
<laser><act1>8224</act1><act2>9995</act2><auto>yes</auto><preset1>8224</preset1><preset2>9995</preset2></laser>
<blocks><act>N19350 G01 X4124 Y0 Z24 F10000</act><auto>yes</auto><pas>N19355 Q= P10 (LASER ON)</pas><sub>basis</sub><temp>N1010 G10 (STA-PO-PI-LI-P6018DA_A1)</temp></blocks>
<axes><auto>yes</auto><ax1>+05073.133</ax1><ax3>+00028.720</ax3><ax4>+04678.163</ax4><ax5>+04678.163</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax3>5</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00025.605</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00024.015</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>1</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>N2250 G10</temp></blocks>
<blocks><act>N1055 G10</act><auto>yes</auto><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>N1030 G90 O= P113 U2020 F= P101 M02</temp></blocks>
<laser><act1>9395</act1><act2>9998</act2><auto>yes</auto><preset1>9395</preset1><preset2>9998</preset2></laser>
<blocks><act>N19360 G01 X4121 Y0 Z24 A358.163 F3000</act><auto>yes</auto><pas>N19365 G01 X4121 Y0 Z24 A356.327 F9047</pas><sub>basis</sub><temp>N19370 G01 X4121 Y0 Z24 A354.49</temp></blocks>
<axes><auto>yes</auto><ax1>+05072.029</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05070.909</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05070.133</ax1><sub>pos</sub></axes>
<blocks><act>N19365 G01 X4121 Y0 Z24 A356.327 F9047</act><auto>yes</auto><pas>N19370 G01 X4121 Y0 Z24 A354.49</pas><sub>basis</sub><temp>N19375 G01 X4121 Y0 Z24 A352.653</temp></blocks>
<laser><act1>9267</act1><act2>9995</act2><auto>yes</auto><preset1>9267</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax4>+04675.805</ax4><ax5>+04675.805</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N19370 G01 X4121 Y0 Z24 A354.49</act><auto>yes</auto><pas>N19375 G01 X4121 Y0 Z24 A352.653</pas><sub>basis</sub><temp>N19380 G01 X4121 Y0 Z24 A350.816</temp></blocks>
<axes><auto>yes</auto><ax4>+04672.427</ax4><ax5>+04672.427</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04669.050</ax4><ax5>+04669.050</ax5><sub>pos</sub></axes>
<blocks><act>N19390 G01 X4121 Y0 Z24 A347.143</act><auto>yes</auto><pas>N19395 G01 X4121 Y0 Z24 A345.306</pas><sub>basis</sub><temp>N19400 G01 X4121 Y0 Z24 A343.085</temp></blocks>
<axes><auto>yes</auto><ax4>+04665.672</ax4><ax5>+04665.672</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04663.257</ax4><ax5>+04663.257</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>5</ax4><ax5>5</ax5><sub>vel</sub></axes>
<laser><act1>7795</act1><auto>yes</auto><preset1>7795</preset1></laser>
<blocks><act>N19405 G01 X4121 Y0 Z24 A343.083</act><auto>yes</auto><pas>N19410 G01 X4121 Y0 Z24 A341.633</pas><sub>basis</sub><temp>N19415 G01 X4121 Y0 Z24 A339.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04662.007</ax4><ax5>+04662.007</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04658.629</ax4><ax5>+04658.629</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04655.252</ax4><ax5>+04655.252</ax5><sub>pos</sub></axes>
<blocks><act>N19435 G01 X4121 Y0 Z24 A332.449</act><auto>yes</auto><pas>N19440 G01 X4121 Y0 Z24 A330.612</pas><sub>basis</sub><temp>N19445 G01 X4121 Y0 Z24 A328.775</temp></blocks>
<axes><auto>yes</auto><ax4>+04651.874</ax4><ax5>+04651.874</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04648.919</ax4><ax5>+04648.919</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04645.541</ax4><ax5>+04645.541</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04642.164</ax4><ax5>+04642.164</ax5><sub>pos</sub></axes>
<blocks><act>N19470 G01 X4121 Y0 Z24 A319.592</act><auto>yes</auto><pas>N19475 G01 X4121 Y0 Z24 A317.755</pas><sub>basis</sub><temp>N19480 G01 X4121 Y0 Z24 A315.918</temp></blocks>
<axes><auto>yes</auto><ax4>+04638.786</ax4><ax5>+04638.786</ax5><sub>pos</sub></axes>
<blocks><act>N19480 G01 X4121 Y0 Z24 A315.918</act><auto>yes</auto><pas>N19485 G01 X4121 Y0 Z24 A314.082</pas><sub>basis</sub><temp>N19490 G01 X4121 Y0 Z24 A312.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04635.409</ax4><ax5>+04635.409</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04632.031</ax4><ax5>+04632.031</ax5><sub>pos</sub></axes>
<blocks><act>N19495 G01 X4121 Y0 Z24 A310.408</act><auto>yes</auto><pas>N19500 G01 X4121 Y0 Z24 A308.571</pas><sub>basis</sub><temp>N19505 G01 X4121 Y0 Z24 A306.735</temp></blocks>
<axes><auto>yes</auto><ax4>+04629.076</ax4><ax5>+04629.076</ax5><sub>pos</sub></axes>
<blocks><act>N19505 G01 X4121 Y0 Z24 A306.735</act><auto>yes</auto><pas>N19510 G01 X4121 Y0 Z24 A304.898</pas><sub>basis</sub><temp>N19515 G01 X4121 Y0 Z24 A303.061</temp></blocks>
<axes><auto>yes</auto><ax4>+04625.698</ax4><ax5>+04625.698</ax5><sub>pos</sub></axes>
<blocks><act>N19515 G01 X4121 Y0 Z24 A303.061</act><auto>yes</auto><pas>N19520 G01 X4121 Y0 Z24 A301.225</pas><sub>basis</sub><temp>N19525 G01 X4121 Y0 Z24 A299.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04622.321</ax4><ax5>+04622.321</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04618.943</ax4><ax5>+04618.943</ax5><sub>pos</sub></axes>
<blocks><act>N19530 G01 X4121 Y0 Z24 A297.551</act><auto>yes</auto><pas>N19535 G01 X4121 Y0 Z24 A295.714</pas><sub>basis</sub><temp>N19540 G01 X4121 Y0 Z24 A293.878</temp></blocks>
<axes><auto>yes</auto><ax4>+04615.988</ax4><ax5>+04615.988</ax5><sub>pos</sub></axes>
<blocks><act>N19540 G01 X4121 Y0 Z24 A293.878</act><auto>yes</auto><pas>N19545 G01 X4121 Y0 Z24 A292.041</pas><sub>basis</sub><temp>N19550 G01 X4121 Y0 Z24 A290.204</temp></blocks>
<axes><auto>yes</auto><ax4>+04612.610</ax4><ax5>+04612.610</ax5><sub>pos</sub></axes>
<blocks><act>N19550 G01 X4121 Y0 Z24 A290.204</act><auto>yes</auto><pas>N19555 G01 X4121 Y0 Z24 A288.367</pas><sub>basis</sub><temp>N19560 G01 X4121 Y0 Z24 A286.531</temp></blocks>
<axes><auto>yes</auto><ax4>+04609.233</ax4><ax5>+04609.233</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04605.855</ax4><ax5>+04605.855</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04602.477</ax4><ax5>+04602.477</ax5><sub>pos</sub></axes>
<blocks><act>N19575 G01 X4121 Y0 Z24 A281.02</act><auto>yes</auto><pas>N19580 G01 X4121 Y0 Z24 A279.184</pas><sub>basis</sub><temp>N19585 G01 X4121 Y0 Z24 A277.347</temp></blocks>
<axes><auto>yes</auto><ax4>+04599.100</ax4><ax5>+04599.100</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04596.145</ax4><ax5>+04596.145</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04592.767</ax4><ax5>+04592.767</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04589.389</ax4><ax5>+04589.389</ax5><sub>pos</sub></axes>
<blocks><act>N19615 G01 X4121 Y0 Z24 A266.327</act><auto>yes</auto><pas>N19620 G01 X4121 Y0 Z24 A264.49</pas><sub>basis</sub><temp>N19625 G01 X4121 Y0 Z24 A262.653</temp></blocks>
<axes><auto>yes</auto><ax4>+04586.012</ax4><ax5>+04586.012</ax5><sub>pos</sub></axes>
<blocks><act>N19620 G01 X4121 Y0 Z24 A264.49</act><auto>yes</auto><pas>N19625 G01 X4121 Y0 Z24 A262.653</pas><sub>basis</sub><temp>N19630 G01 X4121 Y0 Z24 A260.816</temp></blocks>
<axes><auto>yes</auto><ax4>+04583.057</ax4><ax5>+04583.057</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04579.679</ax4><ax5>+04579.679</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04576.301</ax4><ax5>+04576.301</ax5><sub>pos</sub></axes>
<blocks><act>N19650 G01 X4121 Y0 Z24 A253.469</act><auto>yes</auto><pas>N19655 G01 X4121 Y0 Z24 A251.633</pas><sub>basis</sub><temp>N19660 G01 X4121 Y0 Z24 A249.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04572.924</ax4><ax5>+04572.924</ax5><sub>pos</sub></axes>
<blocks><act>N19655 G01 X4121 Y0 Z24 A251.633</act><auto>yes</auto><pas>N19660 G01 X4121 Y0 Z24 A249.796</pas><sub>basis</sub><temp>N19665 G01 X4121 Y0 Z24 A247.959</temp></blocks>
<axes><auto>yes</auto><ax4>+04569.546</ax4><ax5>+04569.546</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04566.169</ax4><ax5>+04566.169</ax5><sub>pos</sub></axes>
<blocks><act>N19675 G01 X4121 Y0 Z24 A244.286</act><auto>yes</auto><pas>N19680 G01 X4121 Y0 Z24 A242.449</pas><sub>basis</sub><temp>N19685 G01 X4121 Y0 Z24 A240.612</temp></blocks>
<axes><auto>yes</auto><ax4>+04563.213</ax4><ax5>+04563.213</ax5><sub>pos</sub></axes>
<blocks><act>N19685 G01 X4121 Y0 Z24 A240.612</act><auto>yes</auto><pas>N19690 G01 X4121 Y0 Z24 A238.775</pas><sub>basis</sub><temp>N19695 G01 X4121 Y0 Z24 A236.939</temp></blocks>
<axes><auto>yes</auto><ax4>+04559.836</ax4><ax5>+04559.836</ax5><sub>pos</sub></axes>
<blocks><act>N19695 G01 X4121 Y0 Z24 A236.939</act><auto>yes</auto><pas>N19700 G01 X4121 Y0 Z24 A235.102</pas><sub>basis</sub><temp>N19705 G01 X4121 Y0 Z24 A233.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04556.458</ax4><ax5>+04556.458</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04553.081</ax4><ax5>+04553.081</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04550.125</ax4><ax5>+04550.125</ax5><sub>pos</sub></axes>
<blocks><act>N19720 G01 X4121 Y0 Z24 A227.755</act><auto>yes</auto><pas>N19725 G01 X4121 Y0 Z24 A225.918</pas><sub>basis</sub><temp>N19730 G01 X4121 Y0 Z24 A224.082</temp></blocks>
<axes><auto>yes</auto><ax4>+04546.326</ax4><ax5>+04546.326</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04543.370</ax4><ax5>+04543.370</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04539.993</ax4><ax5>+04539.993</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04536.615</ax4><ax5>+04536.615</ax5><sub>pos</sub></axes>
<blocks><act>N19755 G01 X4121 Y0 Z24 A214.898</act><auto>yes</auto><pas>N19760 G01 X4121 Y0 Z24 A213.061</pas><sub>basis</sub><temp>N19765 G01 X4121 Y0 Z24 A211.225</temp></blocks>
<axes><auto>yes</auto><ax4>+04533.238</ax4><ax5>+04533.238</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04530.282</ax4><ax5>+04530.282</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04526.905</ax4><ax5>+04526.905</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04523.527</ax4><ax5>+04523.527</ax5><sub>pos</sub></axes>
<blocks><act>N19790 G01 X4121 Y0 Z24 A202.041</act><auto>yes</auto><pas>N19795 G01 X4121 Y0 Z24 A200.204</pas><sub>basis</sub><temp>N19800 G01 X4121 Y0 Z24 A198.367</temp></blocks>
<axes><auto>yes</auto><ax4>+04520.150</ax4><ax5>+04520.150</ax5><sub>pos</sub></axes>
<blocks><act>N19800 G01 X4121 Y0 Z24 A198.367</act><auto>yes</auto><pas>N19805 G01 X4121 Y0 Z24 A196.531</pas><sub>basis</sub><temp>N19810 G01 X4121 Y0 Z24 A194.694</temp></blocks>
<axes><auto>yes</auto><ax4>+04516.772</ax4><ax5>+04516.772</ax5><sub>pos</sub></axes>
<blocks><act>N19810 G01 X4121 Y0 Z24 A194.694</act><auto>yes</auto><pas>N19815 G01 X4121 Y0 Z24 A192.857</pas><sub>basis</sub><temp>N19820 G01 X4121 Y0 Z24 A191.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04513.395</ax4><ax5>+04513.395</ax5><sub>pos</sub></axes>
<blocks><act>N19820 G01 X4121 Y0 Z24 A191.02</act><auto>yes</auto><pas>N19825 G01 X4121 Y0 Z24 A189.184</pas><sub>basis</sub><temp>N19830 G01 X4121 Y0 Z24 A187.347</temp></blocks>
<axes><auto>yes</auto><ax4>+04510.439</ax4><ax5>+04510.439</ax5><sub>pos</sub></axes>
<blocks><act>N19830 G01 X4121 Y0 Z24 A187.347</act><auto>yes</auto><pas>N19835 G01 X4121 Y0 Z24 A185.51</pas><sub>basis</sub><temp>N19840 G01 X4121 Y0 Z24 A183.673</temp></blocks>
<axes><auto>yes</auto><ax4>+04507.062</ax4><ax5>+04507.062</ax5><sub>pos</sub></axes>
<blocks><act>N19835 G01 X4121 Y0 Z24 A185.51</act><auto>yes</auto><pas>N19840 G01 X4121 Y0 Z24 A183.673</pas><sub>basis</sub><temp>N19845 G01 X4121 Y0 Z24 A181.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04503.684</ax4><ax5>+04503.684</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04500.307</ax4><ax5>+04500.307</ax5><sub>pos</sub></axes>
<blocks><act>N19855 G01 X4121 Y0 Z24 A178.163</act><auto>yes</auto><pas>N19860 G01 X4121 Y0 Z24 A176.327</pas><sub>basis</sub><temp>N19865 G01 X4121 Y0 Z24 A174.49</temp></blocks>
<axes><auto>yes</auto><ax4>+04497.351</ax4><ax5>+04497.351</ax5><sub>pos</sub></axes>
<blocks><act>N19865 G01 X4121 Y0 Z24 A174.49</act><auto>yes</auto><pas>N19870 G01 X4121 Y0 Z24 A172.653</pas><sub>basis</sub><temp>N19875 G01 X4121 Y0 Z24 A170.816</temp></blocks>
<axes><auto>yes</auto><ax4>+04493.974</ax4><ax5>+04493.974</ax5><sub>pos</sub></axes>
<blocks><act>N19870 G01 X4121 Y0 Z24 A172.653</act><auto>yes</auto><pas>N19875 G01 X4121 Y0 Z24 A170.816</pas><sub>basis</sub><temp>N19880 G01 X4121 Y0 Z24 A168.98</temp></blocks>
<axes><auto>yes</auto><ax4>+04490.596</ax4><ax5>+04490.596</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04487.219</ax4><ax5>+04487.219</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04483.841</ax4><ax5>+04483.841</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04480.464</ax4><ax5>+04480.464</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04477.508</ax4><ax5>+04477.508</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04474.131</ax4><ax5>+04474.131</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04470.753</ax4><ax5>+04470.753</ax5><sub>pos</sub></axes>
<blocks><act>N19935 G01 X4121 Y0 Z24 A148.775</act><auto>yes</auto><pas>N19940 G01 X4121 Y0 Z24 A146.939</pas><sub>basis</sub><temp>N19945 G01 X4121 Y0 Z24 A145.102</temp></blocks>
<axes><auto>yes</auto><ax4>+04467.376</ax4><ax5>+04467.376</ax5><sub>pos</sub></axes>
<blocks><act>N19945 G01 X4121 Y0 Z24 A145.102</act><auto>yes</auto><pas>N19950 G01 X4121 Y0 Z24 A143.265</pas><sub>basis</sub><temp>N19955 G01 X4121 Y0 Z24 A141.429</temp></blocks>
<axes><auto>yes</auto><ax4>+04464.420</ax4><ax5>+04464.420</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04461.043</ax4><ax5>+04461.043</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04457.665</ax4><ax5>+04457.665</ax5><sub>pos</sub></axes>
<blocks><act>N19970 G01 X4121 Y0 Z24 A135.918</act><auto>yes</auto><pas>N19975 G01 X4121 Y0 Z24 A134.082</pas><sub>basis</sub><temp>N19980 G01 X4121 Y0 Z24 A132.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04454.288</ax4><ax5>+04454.288</ax5><sub>pos</sub></axes>
<blocks><act>N19980 G01 X4121 Y0 Z24 A132.245</act><auto>yes</auto><pas>N19985 G01 X4121 Y0 Z24 A130.408</pas><sub>basis</sub><temp>N19990 G01 X4121 Y0 Z24 A128.571</temp></blocks>
<axes><auto>yes</auto><ax4>+04450.910</ax4><ax5>+04450.910</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04447.532</ax4><ax5>+04447.532</ax5><sub>pos</sub></axes>
<blocks><act>N20000 G01 X4121 Y0 Z24 A124.898</act><auto>yes</auto><pas>N20005 G01 X4121 Y0 Z24 A123.061</pas><sub>basis</sub><temp>N20010 G01 X4121 Y0 Z24 A121.225</temp></blocks>
<axes><auto>yes</auto><ax4>+04444.577</ax4><ax5>+04444.577</ax5><sub>pos</sub></axes>
<blocks><act>N20005 G01 X4121 Y0 Z24 A123.061</act><auto>yes</auto><pas>N20010 G01 X4121 Y0 Z24 A121.225</pas><sub>basis</sub><temp>N20015 G01 X4121 Y0 Z24 A119.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04441.200</ax4><ax5>+04441.200</ax5><sub>pos</sub></axes>
<blocks><act>N20015 G01 X4121 Y0 Z24 A119.388</act><auto>yes</auto><pas>N20020 G01 X4121 Y0 Z24 A117.551</pas><sub>basis</sub><temp>N20025 G01 X4121 Y0 Z24 A115.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04437.822</ax4><ax5>+04437.822</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04434.444</ax4><ax5>+04434.444</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04431.067</ax4><ax5>+04431.067</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04428.112</ax4><ax5>+04428.112</ax5><sub>pos</sub></axes>
<blocks><act>N20050 G01 X4121 Y0 Z24 A106.531</act><auto>yes</auto><pas>N20055 G01 X4121 Y0 Z24 A104.694</pas><sub>basis</sub><temp>N20060 G01 X4121 Y0 Z24 A102.857</temp></blocks>
<axes><auto>yes</auto><ax4>+04424.734</ax4><ax5>+04424.734</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04421.356</ax4><ax5>+04421.356</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04417.979</ax4><ax5>+04417.979</ax5><sub>pos</sub></axes>
<blocks><act>N20080 G01 X4121 Y0 Z24 A95.51</act><auto>yes</auto><pas>N20085 G01 X4121 Y0 Z24 A93.673</pas><sub>basis</sub><temp>N20090 G01 X4121 Y0 Z24 A91.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04414.601</ax4><ax5>+04414.601</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04411.646</ax4><ax5>+04411.646</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04408.268</ax4><ax5>+04408.268</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04404.891</ax4><ax5>+04404.891</ax5><sub>pos</sub></axes>
<blocks><act>N20115 G01 X4121 Y0 Z24 A82.653</act><auto>yes</auto><pas>N20120 G01 X4121 Y0 Z24 A80.816</pas><sub>basis</sub><temp>N20125 G01 X4121 Y0 Z24 A78.98</temp></blocks>
<axes><auto>yes</auto><ax4>+04401.513</ax4><ax5>+04401.513</ax5><sub>pos</sub></axes>
<blocks><act>N20125 G01 X4121 Y0 Z24 A78.98</act><auto>yes</auto><pas>N20130 G01 X4121 Y0 Z24 A77.143</pas><sub>basis</sub><temp>N20135 G01 X4121 Y0 Z24 A75.306</temp></blocks>
<axes><auto>yes</auto><ax4>+04398.136</ax4><ax5>+04398.136</ax5><sub>pos</sub></axes>
<blocks><act>N20135 G01 X4121 Y0 Z24 A75.306</act><auto>yes</auto><pas>N20140 G01 X4121 Y0 Z24 A73.469</pas><sub>basis</sub><temp>N20145 G01 X4121 Y0 Z24 A71.633</temp></blocks>
<axes><auto>yes</auto><ax4>+04394.758</ax4><ax5>+04394.758</ax5><sub>pos</sub></axes>
<blocks><act>N20140 G01 X4121 Y0 Z24 A73.469</act><auto>yes</auto><pas>N20145 G01 X4121 Y0 Z24 A71.633</pas><sub>basis</sub><temp>N20150 G01 X4121 Y0 Z24 A69.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04391.803</ax4><ax5>+04391.803</ax5><sub>pos</sub></axes>
<blocks><act>N20150 G01 X4121 Y0 Z24 A69.796</act><auto>yes</auto><pas>N20155 G01 X4121 Y0 Z24 A67.959</pas><sub>basis</sub><temp>N20160 G01 X4121 Y0 Z24 A66.122</temp></blocks>
<axes><auto>yes</auto><ax4>+04388.425</ax4><ax5>+04388.425</ax5><sub>pos</sub></axes>
<blocks><act>N20160 G01 X4121 Y0 Z24 A66.122</act><auto>yes</auto><pas>N20165 G01 X4121 Y0 Z24 A64.286</pas><sub>basis</sub><temp>N20170 G01 X4121 Y0 Z24 A62.449</temp></blocks>
<axes><auto>yes</auto><ax4>+04385.048</ax4><ax5>+04385.048</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04381.670</ax4><ax5>+04381.670</ax5><sub>pos</sub></axes>
<blocks><act>N20180 G01 X4121 Y0 Z24 A58.775</act><auto>yes</auto><pas>N20185 G01 X4121 Y0 Z24 A56.939</pas><sub>basis</sub><temp>N20190 G01 X4121 Y0 Z24 A55.102</temp></blocks>
<axes><auto>yes</auto><ax4>+04378.293</ax4><ax5>+04378.293</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04375.337</ax4><ax5>+04375.337</ax5><sub>pos</sub></axes>
<blocks><act>N20195 G01 X4121 Y0 Z24 A53.265</act><auto>yes</auto><pas>N20200 G01 X4121 Y0 Z24 A51.429</pas><sub>basis</sub><temp>N20205 G01 X4121 Y0 Z24 A49.592</temp></blocks>
<axes><auto>yes</auto><ax4>+04371.960</ax4><ax5>+04371.960</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04368.582</ax4><ax5>+04368.582</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04365.205</ax4><ax5>+04365.205</ax5><sub>pos</sub></axes>
<blocks><act>N20225 G01 X4121 Y0 Z24 A42.245</act><auto>yes</auto><pas>N20230 G01 X4121 Y0 Z24 A40.408</pas><sub>basis</sub><temp>N20235 G01 X4121 Y0 Z24 A38.571</temp></blocks>
<axes><auto>yes</auto><ax4>+04361.827</ax4><ax5>+04361.827</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04358.872</ax4><ax5>+04358.872</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04355.494</ax4><ax5>+04355.494</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04352.117</ax4><ax5>+04352.117</ax5><sub>pos</sub></axes>
<blocks><act>N20260 G01 X4121 Y0 Z24 A29.388</act><auto>yes</auto><pas>N20265 G01 X4121 Y0 Z24 A27.551</pas><sub>basis</sub><temp>N20270 G01 X4121 Y0 Z24 A25.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04348.739</ax4><ax5>+04348.739</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04345.362</ax4><ax5>+04345.362</ax5><sub>pos</sub></axes>
<blocks><act>N20275 G01 X4121 Y0 Z24 A23.878</act><auto>yes</auto><pas>N20280 G01 X4121 Y0 Z24 A22.041</pas><sub>basis</sub><temp>N20285 G01 X4121 Y0 Z24 A20.204</temp></blocks>
<axes><auto>yes</auto><ax4>+04342.406</ax4><ax5>+04342.406</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04339.029</ax4><ax5>+04339.029</ax5><sub>pos</sub></axes>
<blocks><act>N20295 G01 X4121 Y0 Z24 A16.531</act><auto>yes</auto><pas>N20300 G01 X4121 Y0 Z24 A14.694</pas><sub>basis</sub><temp>N20305 G01 X4121 Y0 Z24 A13.243</temp></blocks>
<axes><auto>yes</auto><ax4>+04335.651</ax4><ax5>+04335.651</ax5><sub>pos</sub></axes>
<blocks><act>N20305 G01 X4121 Y0 Z24 A13.243</act><auto>yes</auto><pas>N20310 G01 X4121 Y0 Z24 A13.242</pas><sub>basis</sub><temp>N20315 G01 X4121 Y0 Z24 A11.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04333.272</ax4><ax5>+04333.272</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax4>+04331.609</ax4><ax5>+04331.609</ax5><sub>pos</sub></axes>
<blocks><act>N20320 G01 X4121 Y0 Z24 A9.184</act><auto>yes</auto><pas>N20325 G01 X4121 Y0 Z24 A7.347</pas><sub>basis</sub><temp>N20330 G01 X4121 Y0 Z24 A5.51</temp></blocks>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04328.653</ax4><ax5>+04328.653</ax5><sub>pos</sub></axes>
<blocks><act>N20325 G01 X4121 Y0 Z24 A7.347</act><auto>yes</auto><pas>N20330 G01 X4121 Y0 Z24 A5.51</pas><sub>basis</sub><temp>N20335 G01 X4121 Y0 Z24 A3.673</temp></blocks>
<axes><auto>yes</auto><ax4>+04325.276</ax4><ax5>+04325.276</ax5><sub>pos</sub></axes>
<blocks><act>N20335 G01 X4121 Y0 Z24 A3.673</act><auto>yes</auto><pas>N20340 G01 X4121 Y0 Z24 A1.837</pas><sub>basis</sub><temp>N20345 G01 X4121 Y0 Z24 A0</temp></blocks>
<axes><auto>yes</auto><ax4>+04321.898</ax4><ax5>+04321.898</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04318.732</ax4><ax5>+04318.732</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>9</ax4><ax5>9</ax5><sub>vel</sub></axes>
<blocks><act>N20350 G01 X4121 Y0 Z24 A-1.837</act><auto>yes</auto><pas>N20360 Q= P11 (LASER OFF)</pas><sub>basis</sub><temp>N1010 G10 (END-LI-PO_W)</temp></blocks>
<laser><act1>9987</act1><auto>yes</auto><preset1>9987</preset1></laser>
<axes><auto>yes</auto><ax4>+04318.163</ax4><ax5>+04318.163</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00022.062</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<blocks><act>N20250 G01 X4121 Y0 Z24 A33.061</act><auto>yes</auto><pas>N20250 G01 X4121 Y0 Z24 A33.061</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>+00023.854</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>14</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00027.581</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00029.000</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>N1040 G10</temp></blocks>
<blocks><act>N1050 M223 (UD SP BOARD DOWN)</act><auto>yes</auto><pas>N1060 G4 F200</pas><sub>basis</sub><temp>N1070 G10</temp></blocks>
<blocks><act>N1060 G4 F200</act><auto>yes</auto><pas>N1070 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1100 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><sub>basis</sub></blocks>
<blocks><act>N1020 M206</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<axes><auto>yes</auto><ax1>+05070.452</ax1><ax4>+04318.606</ax4><ax5>+04318.606</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>6</ax4><ax5>6</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><act2>9999</act2><auto>yes</auto><preset1>9994</preset1><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+05072.220</ax1><ax4>+04319.969</ax4><ax5>+04319.969</ax5><sub>pos</sub></axes>
<blocks><act>N20410 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05072.333</ax1><ax4>+04320.044</ax4><ax5>+04320.044</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9267</act1><auto>yes</auto><preset1>9267</preset1></laser>
<axes><auto>yes</auto><ax4>+04321.351</ax4><ax5>+04321.351</ax5><sub>pos</sub></axes>
<laser><act1>9225</act1><auto>yes</auto><preset1>9225</preset1></laser>
<axes><auto>yes</auto><ax3>+00028.976</ax3><ax4>+04321.837</ax4><ax5>+04321.837</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>1</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00026.511</ax3><sub>pos</sub></axes>
<blocks><act>N20420 G01 X4123.2 Y0 Z24 F10000</act><auto>yes</auto><pas>N20425 Q990051</pas><sub>basis</sub><temp>N1010 G10 (LCUT-1)</temp></blocks>
<axes><auto>yes</auto><ax3>+00024.083</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>3</ax3><sub>vel</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N1160 G113 V= P141</act><auto>yes</auto><pas>N1170 G210 X0 Y0</pas><sub>basis</sub><temp>N1180 G211 V= P800 F100</temp></blocks>
<blocks><act>N1200 M110</act><auto>yes</auto><pas>N1210 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N1050 Q999992</pas><sub>basis</sub><temp>N1010 G10 ( RECTANGULAR TUBE DETECTION_20161226 )</temp></blocks>
<blocks><auto>yes</auto><pas>N2250 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1055 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>N20435 G01 X4126.2 Y0 Z24 A1.837 F3000</act><auto>yes</auto><pas>N20440 G01 X4126.2 Y0 Z24 A3.673 F9047</pas><sub>basis</sub><temp>N20445 G01 X4126.2 Y0 Z24 A5.51</temp></blocks>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05073.018</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><sub>vel</sub></axes>
<laser><act1>9395</act1><act2>9998</act2><auto>yes</auto><preset1>9395</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05073.998</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05075.118</ax1><sub>pos</sub></axes>
<laser><act1>7348</act1><act2>9995</act2><auto>yes</auto><preset1>7348</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05075.333</ax1><ax4>+04322.928</ax4><ax5>+04322.928</ax5><sub>pos</sub></axes>
<blocks><act>N20445 G01 X4126.2 Y0 Z24 A5.51</act><auto>yes</auto><pas>N20450 G01 X4126.2 Y0 Z24 A7.347</pas><sub>basis</sub><temp>N20455 G01 X4126.2 Y0 Z24 A9.184</temp></blocks>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04326.306</ax4><ax5>+04326.306</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N20455 G01 X4126.2 Y0 Z24 A9.184</act><auto>yes</auto><pas>N20460 G01 X4126.2 Y0 Z24 A11.02</pas><sub>basis</sub><temp>N20465 G01 X4126.2 Y0 Z24 A12.857</temp></blocks>
<axes><auto>yes</auto><ax4>+04329.683</ax4><ax5>+04329.683</ax5><sub>pos</sub></axes>
<blocks><act>N20460 G01 X4126.2 Y0 Z24 A11.02</act><auto>yes</auto><pas>N20465 G01 X4126.2 Y0 Z24 A12.857</pas><sub>basis</sub><temp>N20470 G01 X4126.2 Y0 Z24 A14.694</temp></blocks>
<axes><auto>yes</auto><ax4>+04332.639</ax4><ax5>+04332.639</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04335.983</ax4><ax5>+04335.983</ax5><sub>pos</sub></axes>
<blocks><act>N20475 G01 X4126.2 Y0 Z24 A16.915</act><auto>yes</auto><pas>N20480 G01 X4126.2 Y0 Z24 A16.917</pas><sub>basis</sub><temp>N20485 G01 X4126.2 Y0 Z24 A18.367</temp></blocks>
<axes><auto>yes</auto><ax4>+04337.083</ax4><ax5>+04337.083</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>5</ax4><ax5>5</ax5><sub>vel</sub></axes>
<blocks><act>N20485 G01 X4126.2 Y0 Z24 A18.367</act><auto>yes</auto><pas>N20490 G01 X4126.2 Y0 Z24 A20.204</pas><sub>basis</sub><temp>N20495 G01 X4126.2 Y0 Z24 A22.041</temp></blocks>
<axes><auto>yes</auto><ax4>+04340.104</ax4><ax5>+04340.104</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04343.482</ax4><ax5>+04343.482</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N20505 G01 X4126.2 Y0 Z24 A25.714</act><auto>yes</auto><pas>N20510 G01 X4126.2 Y0 Z24 A27.551</pas><sub>basis</sub><temp>N20515 G01 X4126.2 Y0 Z24 A29.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04346.437</ax4><ax5>+04346.437</ax5><sub>pos</sub></axes>
<blocks><act>N20515 G01 X4126.2 Y0 Z24 A29.388</act><auto>yes</auto><pas>N20520 G01 X4126.2 Y0 Z24 A31.225</pas><sub>basis</sub><temp>N20525 G01 X4126.2 Y0 Z24 A33.061</temp></blocks>
<axes><auto>yes</auto><ax4>+04349.815</ax4><ax5>+04349.815</ax5><sub>pos</sub></axes>
<blocks><act>N20520 G01 X4126.2 Y0 Z24 A31.225</act><auto>yes</auto><pas>N20525 G01 X4126.2 Y0 Z24 A33.061</pas><sub>basis</sub><temp>N20530 G01 X4126.2 Y0 Z24 A34.898</temp></blocks>
<axes><auto>yes</auto><ax4>+04353.192</ax4><ax5>+04353.192</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04356.570</ax4><ax5>+04356.570</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04359.947</ax4><ax5>+04359.947</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04362.903</ax4><ax5>+04362.903</ax5><sub>pos</sub></axes>
<blocks><act>N20560 G01 X4126.2 Y0 Z24 A45.918</act><auto>yes</auto><pas>N20565 G01 X4126.2 Y0 Z24 A47.755</pas><sub>basis</sub><temp>N20570 G01 X4126.2 Y0 Z24 A49.592</temp></blocks>
<axes><auto>yes</auto><ax4>+04366.280</ax4><ax5>+04366.280</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04369.658</ax4><ax5>+04369.658</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04373.035</ax4><ax5>+04373.035</ax5><sub>pos</sub></axes>
<blocks><act>N20585 G01 X4126.2 Y0 Z24 A55.102</act><auto>yes</auto><pas>N20590 G01 X4126.2 Y0 Z24 A56.939</pas><sub>basis</sub><temp>N20595 G01 X4126.2 Y0 Z24 A58.775</temp></blocks>
<axes><auto>yes</auto><ax4>+04376.413</ax4><ax5>+04376.413</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04379.368</ax4><ax5>+04379.368</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04382.746</ax4><ax5>+04382.746</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04386.123</ax4><ax5>+04386.123</ax5><sub>pos</sub></axes>
<blocks><act>N20620 G01 X4126.2 Y0 Z24 A67.959</act><auto>yes</auto><pas>N20625 G01 X4126.2 Y0 Z24 A69.796</pas><sub>basis</sub><temp>N20630 G01 X4126.2 Y0 Z24 A71.633</temp></blocks>
<axes><auto>yes</auto><ax4>+04389.501</ax4><ax5>+04389.501</ax5><sub>pos</sub></axes>
<blocks><act>N20630 G01 X4126.2 Y0 Z24 A71.633</act><auto>yes</auto><pas>N20635 G01 X4126.2 Y0 Z24 A73.469</pas><sub>basis</sub><temp>N20640 G01 X4126.2 Y0 Z24 A75.306</temp></blocks>
<axes><auto>yes</auto><ax4>+04392.878</ax4><ax5>+04392.878</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04396.256</ax4><ax5>+04396.256</ax5><sub>pos</sub></axes>
<blocks><act>N20650 G01 X4126.2 Y0 Z24 A78.98</act><auto>yes</auto><pas>N20655 G01 X4126.2 Y0 Z24 A80.816</pas><sub>basis</sub><temp>N20660 G01 X4126.2 Y0 Z24 A82.653</temp></blocks>
<axes><auto>yes</auto><ax4>+04399.211</ax4><ax5>+04399.211</ax5><sub>pos</sub></axes>
<blocks><act>N20655 G01 X4126.2 Y0 Z24 A80.816</act><auto>yes</auto><pas>N20660 G01 X4126.2 Y0 Z24 A82.653</pas><sub>basis</sub><temp>N20665 G01 X4126.2 Y0 Z24 A84.49</temp></blocks>
<axes><auto>yes</auto><ax4>+04402.589</ax4><ax5>+04402.589</ax5><sub>pos</sub></axes>
<blocks><act>N20665 G01 X4126.2 Y0 Z24 A84.49</act><auto>yes</auto><pas>N20670 G01 X4126.2 Y0 Z24 A86.327</pas><sub>basis</sub><temp>N20675 G01 X4126.2 Y0 Z24 A88.163</temp></blocks>
<axes><auto>yes</auto><ax4>+04405.966</ax4><ax5>+04405.966</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04409.344</ax4><ax5>+04409.344</ax5><sub>pos</sub></axes>
<blocks><act>N20685 G01 X4126.2 Y0 Z24 A91.837</act><auto>yes</auto><pas>N20690 G01 X4126.2 Y0 Z24 A93.673</pas><sub>basis</sub><temp>N20695 G01 X4126.2 Y0 Z24 A95.51</temp></blocks>
<axes><auto>yes</auto><ax4>+04412.299</ax4><ax5>+04412.299</ax5><sub>pos</sub></axes>
<blocks><act>N20690 G01 X4126.2 Y0 Z24 A93.673</act><auto>yes</auto><pas>N20695 G01 X4126.2 Y0 Z24 A95.51</pas><sub>basis</sub><temp>N20700 G01 X4126.2 Y0 Z24 A97.347</temp></blocks>
<axes><auto>yes</auto><ax4>+04415.677</ax4><ax5>+04415.677</ax5><sub>pos</sub></axes>
<blocks><act>N20700 G01 X4126.2 Y0 Z24 A97.347</act><auto>yes</auto><pas>N20705 G01 X4126.2 Y0 Z24 A99.184</pas><sub>basis</sub><temp>N20710 G01 X4126.2 Y0 Z24 A101.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04419.054</ax4><ax5>+04419.054</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04422.432</ax4><ax5>+04422.432</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04425.810</ax4><ax5>+04425.810</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04429.187</ax4><ax5>+04429.187</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04432.142</ax4><ax5>+04432.142</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04435.520</ax4><ax5>+04435.520</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04438.898</ax4><ax5>+04438.898</ax5><sub>pos</sub></axes>
<blocks><act>N20765 G01 X4126.2 Y0 Z24 A121.225</act><auto>yes</auto><pas>N20770 G01 X4126.2 Y0 Z24 A123.061</pas><sub>basis</sub><temp>N20775 G01 X4126.2 Y0 Z24 A124.898</temp></blocks>
<axes><auto>yes</auto><ax4>+04442.275</ax4><ax5>+04442.275</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04445.653</ax4><ax5>+04445.653</ax5><sub>pos</sub></axes>
<blocks><act>N20785 G01 X4126.2 Y0 Z24 A128.571</act><auto>yes</auto><pas>N20790 G01 X4126.2 Y0 Z24 A130.408</pas><sub>basis</sub><temp>N20795 G01 X4126.2 Y0 Z24 A132.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04448.608</ax4><ax5>+04448.608</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04451.986</ax4><ax5>+04451.986</ax5><sub>pos</sub></axes>
<blocks><act>N20800 G01 X4126.2 Y0 Z24 A134.082</act><auto>yes</auto><pas>N20805 G01 X4126.2 Y0 Z24 A135.918</pas><sub>basis</sub><temp>N20810 G01 X4126.2 Y0 Z24 A137.755</temp></blocks>
<axes><auto>yes</auto><ax4>+04455.363</ax4><ax5>+04455.363</ax5><sub>pos</sub></axes>
<blocks><act>N20810 G01 X4126.2 Y0 Z24 A137.755</act><auto>yes</auto><pas>N20815 G01 X4126.2 Y0 Z24 A139.592</pas><sub>basis</sub><temp>N20820 G01 X4126.2 Y0 Z24 A141.429</temp></blocks>
<axes><auto>yes</auto><ax4>+04458.741</ax4><ax5>+04458.741</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04462.118</ax4><ax5>+04462.118</ax5><sub>pos</sub></axes>
<blocks><act>N20830 G01 X4126.2 Y0 Z24 A145.102</act><auto>yes</auto><pas>N20835 G01 X4126.2 Y0 Z24 A146.939</pas><sub>basis</sub><temp>N20840 G01 X4126.2 Y0 Z24 A148.775</temp></blocks>
<axes><auto>yes</auto><ax4>+04465.074</ax4><ax5>+04465.074</ax5><sub>pos</sub></axes>
<blocks><act>N20835 G01 X4126.2 Y0 Z24 A146.939</act><auto>yes</auto><pas>N20840 G01 X4126.2 Y0 Z24 A148.775</pas><sub>basis</sub><temp>N20845 G01 X4126.2 Y0 Z24 A150.612</temp></blocks>
<axes><auto>yes</auto><ax4>+04468.451</ax4><ax5>+04468.451</ax5><sub>pos</sub></axes>
<blocks><act>N20845 G01 X4126.2 Y0 Z24 A150.612</act><auto>yes</auto><pas>N20850 G01 X4126.2 Y0 Z24 A152.449</pas><sub>basis</sub><temp>N20855 G01 X4126.2 Y0 Z24 A154.286</temp></blocks>
<axes><auto>yes</auto><ax4>+04471.829</ax4><ax5>+04471.829</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04475.206</ax4><ax5>+04475.206</ax5><sub>pos</sub></axes>
<blocks><act>N20865 G01 X4126.2 Y0 Z24 A157.959</act><auto>yes</auto><pas>N20870 G01 X4126.2 Y0 Z24 A159.796</pas><sub>basis</sub><temp>N20875 G01 X4126.2 Y0 Z24 A161.633</temp></blocks>
<axes><auto>yes</auto><ax4>+04478.584</ax4><ax5>+04478.584</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04481.539</ax4><ax5>+04481.539</ax5><sub>pos</sub></axes>
<blocks><act>N20880 G01 X4126.2 Y0 Z24 A163.469</act><auto>yes</auto><pas>N20885 G01 X4126.2 Y0 Z24 A165.306</pas><sub>basis</sub><temp>N20890 G01 X4126.2 Y0 Z24 A167.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04484.917</ax4><ax5>+04484.917</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04488.294</ax4><ax5>+04488.294</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04491.672</ax4><ax5>+04491.672</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04495.049</ax4><ax5>+04495.049</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04498.005</ax4><ax5>+04498.005</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04501.382</ax4><ax5>+04501.382</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04504.760</ax4><ax5>+04504.760</ax5><sub>pos</sub></axes>
<blocks><act>N20945 G01 X4126.2 Y0 Z24 A187.347</act><auto>yes</auto><pas>N20950 G01 X4126.2 Y0 Z24 A189.184</pas><sub>basis</sub><temp>N20955 G01 X4126.2 Y0 Z24 A191.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04508.137</ax4><ax5>+04508.137</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04511.515</ax4><ax5>+04511.515</ax5><sub>pos</sub></axes>
<blocks><act>N20960 G01 X4126.2 Y0 Z24 A192.857</act><auto>yes</auto><pas>N20965 G01 X4126.2 Y0 Z24 A194.694</pas><sub>basis</sub><temp>N20970 G01 X4126.2 Y0 Z24 A196.531</temp></blocks>
<axes><auto>yes</auto><ax4>+04514.470</ax4><ax5>+04514.470</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04517.848</ax4><ax5>+04517.848</ax5><sub>pos</sub></axes>
<blocks><act>N20980 G01 X4126.2 Y0 Z24 A200.204</act><auto>yes</auto><pas>N20985 G01 X4126.2 Y0 Z24 A202.041</pas><sub>basis</sub><temp>N20990 G01 X4126.2 Y0 Z24 A203.878</temp></blocks>
<axes><auto>yes</auto><ax4>+04521.225</ax4><ax5>+04521.225</ax5><sub>pos</sub></axes>
<blocks><act>N20990 G01 X4126.2 Y0 Z24 A203.878</act><auto>yes</auto><pas>N20995 G01 X4126.2 Y0 Z24 A205.714</pas><sub>basis</sub><temp>N21000 G01 X4126.2 Y0 Z24 A207.551</temp></blocks>
<axes><auto>yes</auto><ax4>+04524.603</ax4><ax5>+04524.603</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04527.980</ax4><ax5>+04527.980</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04530.936</ax4><ax5>+04530.936</ax5><sub>pos</sub></axes>
<blocks><act>N21015 G01 X4126.2 Y0 Z24 A213.061</act><auto>yes</auto><pas>N21020 G01 X4126.2 Y0 Z24 A214.898</pas><sub>basis</sub><temp>N21025 G01 X4126.2 Y0 Z24 A216.735</temp></blocks>
<axes><auto>yes</auto><ax4>+04534.313</ax4><ax5>+04534.313</ax5><sub>pos</sub></axes>
<blocks><act>N21025 G01 X4126.2 Y0 Z24 A216.735</act><auto>yes</auto><pas>N21030 G01 X4126.2 Y0 Z24 A218.571</pas><sub>basis</sub><temp>N21035 G01 X4126.2 Y0 Z24 A220.408</temp></blocks>
<axes><auto>yes</auto><ax4>+04537.691</ax4><ax5>+04537.691</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04541.068</ax4><ax5>+04541.068</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04544.446</ax4><ax5>+04544.446</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04547.401</ax4><ax5>+04547.401</ax5><sub>pos</sub></axes>
<blocks><act>N21060 G01 X4126.2 Y0 Z24 A229.592</act><auto>yes</auto><pas>N21065 G01 X4126.2 Y0 Z24 A231.429</pas><sub>basis</sub><temp>N21070 G01 X4126.2 Y0 Z24 A233.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04550.779</ax4><ax5>+04550.779</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04554.156</ax4><ax5>+04554.156</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04557.534</ax4><ax5>+04557.534</ax5><sub>pos</sub></axes>
<blocks><act>N21090 G01 X4126.2 Y0 Z24 A240.612</act><auto>yes</auto><pas>N21095 G01 X4126.2 Y0 Z24 A242.449</pas><sub>basis</sub><temp>N21100 G01 X4126.2 Y0 Z24 A244.286</temp></blocks>
<axes><auto>yes</auto><ax4>+04560.911</ax4><ax5>+04560.911</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04563.867</ax4><ax5>+04563.867</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04567.244</ax4><ax5>+04567.244</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04570.622</ax4><ax5>+04570.622</ax5><sub>pos</sub></axes>
<blocks><act>N21125 G01 X4126.2 Y0 Z24 A253.469</act><auto>yes</auto><pas>N21130 G01 X4126.2 Y0 Z24 A255.306</pas><sub>basis</sub><temp>N21135 G01 X4126.2 Y0 Z24 A257.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04573.999</ax4><ax5>+04573.999</ax5><sub>pos</sub></axes>
<blocks><act>N21130 G01 X4126.2 Y0 Z24 A255.306</act><auto>yes</auto><pas>N21135 G01 X4126.2 Y0 Z24 A257.143</pas><sub>basis</sub><temp>N21140 G01 X4126.2 Y0 Z24 A258.98</temp></blocks>
<axes><auto>yes</auto><ax4>+04577.377</ax4><ax5>+04577.377</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04580.332</ax4><ax5>+04580.332</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04583.710</ax4><ax5>+04583.710</ax5><sub>pos</sub></axes>
<blocks><act>N21160 G01 X4126.2 Y0 Z24 A266.327</act><auto>yes</auto><pas>N21165 G01 X4126.2 Y0 Z24 A268.163</pas><sub>basis</sub><temp>N21170 G01 X4126.2 Y0 Z24 A270</temp></blocks>
<axes><auto>yes</auto><ax4>+04587.087</ax4><ax5>+04587.087</ax5><sub>pos</sub></axes>
<blocks><act>N21170 G01 X4126.2 Y0 Z24 A270</act><auto>yes</auto><pas>N21175 G01 X4126.2 Y0 Z24 A271.837</pas><sub>basis</sub><temp>N21180 G01 X4126.2 Y0 Z24 A273.673</temp></blocks>
<axes><auto>yes</auto><ax4>+04590.465</ax4><ax5>+04590.465</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04593.842</ax4><ax5>+04593.842</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04596.798</ax4><ax5>+04596.798</ax5><sub>pos</sub></axes>
<blocks><act>N21195 G01 X4126.2 Y0 Z24 A279.184</act><auto>yes</auto><pas>N21200 G01 X4126.2 Y0 Z24 A281.02</pas><sub>basis</sub><temp>N21205 G01 X4126.2 Y0 Z24 A282.857</temp></blocks>
<axes><auto>yes</auto><ax4>+04600.175</ax4><ax5>+04600.175</ax5><sub>pos</sub></axes>
<blocks><act>N21205 G01 X4126.2 Y0 Z24 A282.857</act><auto>yes</auto><pas>N21210 G01 X4126.2 Y0 Z24 A284.694</pas><sub>basis</sub><temp>N21215 G01 X4126.2 Y0 Z24 A286.531</temp></blocks>
<axes><auto>yes</auto><ax4>+04603.553</ax4><ax5>+04603.553</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04606.930</ax4><ax5>+04606.930</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04610.308</ax4><ax5>+04610.308</ax5><sub>pos</sub></axes>
<blocks><act>N21230 G01 X4126.2 Y0 Z24 A292.041</act><auto>yes</auto><pas>N21235 G01 X4126.2 Y0 Z24 A293.878</pas><sub>basis</sub><temp>N21240 G01 X4126.2 Y0 Z24 A295.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04613.686</ax4><ax5>+04613.686</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04616.641</ax4><ax5>+04616.641</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04620.018</ax4><ax5>+04620.018</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04623.396</ax4><ax5>+04623.396</ax5><sub>pos</sub></axes>
<blocks><act>N21265 G01 X4126.2 Y0 Z24 A304.898</act><auto>yes</auto><pas>N21270 G01 X4126.2 Y0 Z24 A306.735</pas><sub>basis</sub><temp>N21275 G01 X4126.2 Y0 Z24 A308.571</temp></blocks>
<axes><auto>yes</auto><ax4>+04626.774</ax4><ax5>+04626.774</ax5><sub>pos</sub></axes>
<blocks><act>N21275 G01 X4126.2 Y0 Z24 A308.571</act><auto>yes</auto><pas>N21280 G01 X4126.2 Y0 Z24 A310.408</pas><sub>basis</sub><temp>N21285 G01 X4126.2 Y0 Z24 A312.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04629.729</ax4><ax5>+04629.729</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04633.106</ax4><ax5>+04633.106</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04636.484</ax4><ax5>+04636.484</ax5><sub>pos</sub></axes>
<blocks><act>N21300 G01 X4126.2 Y0 Z24 A317.755</act><auto>yes</auto><pas>N21305 G01 X4126.2 Y0 Z24 A319.592</pas><sub>basis</sub><temp>N21310 G01 X4126.2 Y0 Z24 A321.429</temp></blocks>
<axes><auto>yes</auto><ax4>+04639.862</ax4><ax5>+04639.862</ax5><sub>pos</sub></axes>
<blocks><act>N21310 G01 X4126.2 Y0 Z24 A321.429</act><auto>yes</auto><pas>N21315 G01 X4126.2 Y0 Z24 A323.265</pas><sub>basis</sub><temp>N21320 G01 X4126.2 Y0 Z24 A325.102</temp></blocks>
<axes><auto>yes</auto><ax4>+04643.239</ax4><ax5>+04643.239</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04646.617</ax4><ax5>+04646.617</ax5><sub>pos</sub></axes>
<blocks><act>N21330 G01 X4126.2 Y0 Z24 A328.775</act><auto>yes</auto><pas>N21335 G01 X4126.2 Y0 Z24 A330.612</pas><sub>basis</sub><temp>N21340 G01 X4126.2 Y0 Z24 A332.449</temp></blocks>
<axes><auto>yes</auto><ax4>+04649.572</ax4><ax5>+04649.572</ax5><sub>pos</sub></axes>
<blocks><act>N21340 G01 X4126.2 Y0 Z24 A332.449</act><auto>yes</auto><pas>N21345 G01 X4126.2 Y0 Z24 A334.286</pas><sub>basis</sub><temp>N21350 G01 X4126.2 Y0 Z24 A336.122</temp></blocks>
<axes><auto>yes</auto><ax4>+04652.950</ax4><ax5>+04652.950</ax5><sub>pos</sub></axes>
<blocks><act>N21345 G01 X4126.2 Y0 Z24 A334.286</act><auto>yes</auto><pas>N21350 G01 X4126.2 Y0 Z24 A336.122</pas><sub>basis</sub><temp>N21355 G01 X4126.2 Y0 Z24 A337.959</temp></blocks>
<axes><auto>yes</auto><ax4>+04656.327</ax4><ax5>+04656.327</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04659.705</ax4><ax5>+04659.705</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04663.082</ax4><ax5>+04663.082</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04666.226</ax4><ax5>+04666.226</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>10</ax4><ax5>10</ax5><sub>vel</sub></axes>
<laser><act1>9887</act1><auto>yes</auto><preset1>9887</preset1></laser>
<axes><auto>yes</auto><ax4>+04666.983</ax4><ax5>+04666.983</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04670.080</ax4><ax5>+04670.080</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04673.458</ax4><ax5>+04673.458</ax5><sub>pos</sub></axes>
<blocks><act>N21410 G01 X4126.2 Y0 Z24 A356.327</act><auto>yes</auto><pas>N21415 G01 X4126.2 Y0 Z24 A358.163</pas><sub>basis</sub><temp>N21420 G01 X4126.2 Y0 Z24 A360</temp></blocks>
<axes><auto>yes</auto><ax4>+04676.835</ax4><ax5>+04676.835</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04679.788</ax4><ax5>+04679.788</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04681.835</ax4><ax5>+04681.835</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax4>+04681.837</ax4><ax5>+04681.837</ax5><sub>pos</sub></axes>
<blocks><act>N20980 G01 X4126.2 Y0 Z24 A200.204</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>+00021.713</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<blocks><act>N1100 M02</act><auto>yes</auto><pas>N21440 G00 Z29</pas><sub>basis</sub><temp>N21445 (====CONTOUR 2 ====)</temp></blocks>
<axes><auto>yes</auto><ax3>+00021.905</ax3><sub>pos</sub></axes>
<blocks><act>N21440 G00 Z29</act><auto>yes</auto><pas>N21445 (====CONTOUR 2 ====)</pas><sub>basis</sub><temp>N21450 G00 X4418.891 Y0 A360</temp></blocks>
<axes><auto>yes</auto><ax3>+00025.233</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>20</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00028.691</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05075.804</ax1><ax3>+00029.000</ax3><ax4>+04681.832</ax4><ax5>+04681.832</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>3</ax1><ax3>0</ax3><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><act2>9999</act2><auto>yes</auto><preset1>9994</preset1><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+05080.510</ax1><ax4>+04681.799</ax4><ax5>+04681.799</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05090.235</ax1><ax4>+04681.734</ax4><ax5>+04681.734</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>22</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05102.863</ax1><ax4>+04681.651</ax4><ax5>+04681.651</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05122.000</ax1><ax4>+04681.527</ax4><ax5>+04681.527</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>40</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05146.157</ax1><ax4>+04681.371</ax4><ax5>+04681.371</ax5><sub>pos</sub></axes>
<blocks><act>N21450 G00 X4418.891 Y0 A360</act><auto>yes</auto><pas>N21455 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax1>+05175.333</ax1><ax4>+04681.184</ax4><ax5>+04681.184</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>59</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05204.980</ax1><ax4>+04680.995</ax4><ax5>+04680.995</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05241.687</ax1><ax4>+04680.766</ax4><ax5>+04680.766</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>66</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05274.426</ax1><ax4>+04680.564</ax4><ax5>+04680.564</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05302.263</ax1><ax4>+04680.393</ax4><ax5>+04680.393</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>47</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05325.201</ax1><ax4>+04680.253</ax4><ax5>+04680.253</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05341.251</ax1><ax4>+04680.156</ax4><ax5>+04680.156</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>30</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05355.000</ax1><ax4>+04680.073</ax4><ax5>+04680.073</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05363.848</ax1><ax4>+04680.021</ax4><ax5>+04680.021</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>12</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05367.796</ax1><ax4>+04680.001</ax4><ax5>+04680.001</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05368.024</ax1><ax4>+04679.956</ax4><ax5>+04679.956</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><sub>vel</sub></axes>
<blocks><act>N21460 G00 X4418.891 Y0 A270.506</act><auto>yes</auto><pas>N21465 G01 X4418.891 Y0 Z24 F10000</pas><sub>basis</sub><temp>N21470 Q= P10 (LASER ON)</temp></blocks>
<laser><act1>9267</act1><auto>yes</auto><preset1>9267</preset1></laser>
<axes><auto>yes</auto><ax4>+04678.405</ax4><ax5>+04678.405</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04673.221</ax4><ax5>+04673.221</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>26</ax4><ax5>26</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04666.708</ax4><ax5>+04666.708</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04656.606</ax4><ax5>+04656.606</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>49</ax4><ax5>49</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04643.668</ax4><ax5>+04643.668</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04628.589</ax4><ax5>+04628.589</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>62</ax4><ax5>62</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04617.024</ax4><ax5>+04617.024</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04606.362</ax4><ax5>+04606.362</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>40</ax4><ax5>40</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04598.427</ax4><ax5>+04598.427</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04593.218</ax4><ax5>+04593.218</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>17</ax4><ax5>17</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04590.736</ax4><ax5>+04590.736</ax5><sub>pos</sub></axes>
<laser><act1>8244</act1><auto>yes</auto><preset1>8244</preset1></laser>
<axes><auto>yes</auto><ax3>+00028.880</ax3><ax4>+04590.506</ax4><ax5>+04590.506</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>3</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00026.045</ax3><sub>pos</sub></axes>
<blocks><act>N21465 G01 X4418.891 Y0 Z24 F10000</act><auto>yes</auto><pas>N21470 Q= P10 (LASER ON)</pas><sub>basis</sub><temp>N1010 G10 (STA-PO-PI-LI-P6018DA_A1)</temp></blocks>
<axes><auto>yes</auto><ax3>+00024.015</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>1</ax3><sub>vel</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>*N1040 IF P174=0 GO1060-P6018DA_A1)</temp></blocks>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<blocks><auto>yes</auto><sub>basis</sub><temp>N2250 G10</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N1010 G10 (PIE1-NORM)</act><auto>yes</auto><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<blocks><act>N1030 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<blocks><act>N1010 G10 (LD0-NORM)</act><auto>yes</auto><sub>basis</sub></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub></blocks>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<blocks><act>N21475 G01 X4421.888 Y0 Z24 A270.881 F3021</act><auto>yes</auto><pas>N21480 G01 X4421.876 Y0 Z24 A271.762 F9040</pas><sub>basis</sub><temp>N21485 G01 X4421.802 Y0 Z24 A273.521 F8984</temp></blocks>
<axes><auto>yes</auto><ax1>+05368.708</ax1><ax4>+04590.609</ax4><ax5>+04590.609</ax5><sub>pos</sub></axes>
<laser><act1>9412</act1><act2>9998</act2><auto>yes</auto><preset1>9412</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05369.828</ax1><ax4>+04590.749</ax4><ax5>+04590.749</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05370.924</ax1><ax4>+04590.878</ax4><ax5>+04590.878</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05371.012</ax1><ax4>+04591.971</ax4><ax5>+04591.971</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N21485 G01 X4421.802 Y0 Z24 A273.521 F8984</act><auto>yes</auto><pas>N21490 G01 X4421.682 Y0 Z24 A275.232 F8873</pas><sub>basis</sub><temp>N21495 G01 X4421.517 Y0 Z24 A276.889 F8706</temp></blocks>
<axes><auto>yes</auto><ax1>+05370.840</ax1><ax4>+04595.291</ax4><ax5>+04595.291</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05370.494</ax1><ax4>+04598.467</ax4><ax5>+04598.467</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05369.977</ax1><ax4>+04601.402</ax4><ax5>+04601.402</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05369.301</ax1><ax4>+04603.892</ax4><ax5>+04603.892</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<laser><act1>9444</act1><act2>9996</act2><auto>yes</auto><preset1>9444</preset1><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05369.069</ax1><ax4>+04604.385</ax4><ax5>+04604.385</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9997</act2><auto>yes</auto><preset1>9994</preset1><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05368.389</ax1><ax4>+04606.213</ax4><ax5>+04606.213</ax5><sub>pos</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05367.411</ax1><ax4>+04607.596</ax4><ax5>+04607.596</ax5><sub>pos</sub></axes>
<blocks><act>N21560 G01 X4417.392 Y0 Z24 A288.118 F3381</act><auto>yes</auto><pas>N21565 G01 X4416.898 Y0 Z24 A288.21 F3046</pas><sub>basis</sub><temp>N21570 G01 X4416.407 Y0 Z24 A288.118</temp></blocks>
<laser><act1>9958</act1><auto>yes</auto><preset1>9958</preset1></laser>
<axes><auto>yes</auto><ax1>+05366.337</ax1><ax4>+04608.179</ax4><ax5>+04608.179</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9433</act1><auto>yes</auto><preset1>9433</preset1></laser>
<blocks><act>N21570 G01 X4416.407 Y0 Z24 A288.118</act><auto>yes</auto><pas>N21575 G01 X4415.92 Y0 Z24 A287.845 F3383</pas><sub>basis</sub><temp>N21580 G01 X4415.44 Y0 Z24 A287.39 F3941</temp></blocks>
<axes><auto>yes</auto><ax1>+05365.236</ax1><ax4>+04607.870</ax4><ax5>+04607.870</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N21580 G01 X4415.44 Y0 Z24 A287.39 F3941</act><auto>yes</auto><pas>N21585 G01 X4414.979 Y0 Z24 A286.769 F4596</pas><sub>basis</sub><temp>N21590 G01 X4414.539 Y0 Z24 A285.989 F5261</temp></blocks>
<axes><auto>yes</auto><ax1>+05364.315</ax1><ax4>+04606.870</ax4><ax5>+04606.870</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>5</ax4><ax5>5</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05363.366</ax1><ax4>+04605.055</ax4><ax5>+04605.055</ax5><sub>pos</sub></axes>
<blocks><act>N21600 G01 X4413.726 Y0 Z24 A283.972 F6483</act><auto>yes</auto><pas>N21605 G01 X4413.366 Y0 Z24 A282.771 F7007</pas><sub>basis</sub><temp>N21610 G01 X4413.038 Y0 Z24 A281.447 F7467</temp></blocks>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05362.563</ax1><ax4>+04602.667</ax4><ax5>+04602.667</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<blocks><act>N21610 G01 X4413.038 Y0 Z24 A281.447 F7467</act><auto>yes</auto><pas>N21615 G01 X4412.746 Y0 Z24 A280.014 F7864</pas><sub>basis</sub><temp>N21620 G01 X4412.494 Y0 Z24 A278.496 F8203</temp></blocks>
<axes><auto>yes</auto><ax1>+05361.923</ax1><ax4>+04599.867</ax4><ax5>+04599.867</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05361.453</ax1><ax4>+04596.781</ax4><ax5>+04596.781</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<blocks><act>N21630 G01 X4412.118 Y0 Z24 A275.229 F8706</act><auto>yes</auto><pas>N21635 G01 X4411.997 Y0 Z24 A273.512 F8872</pas><sub>basis</sub><temp>N21640 G01 X4411.924 Y0 Z24 A271.771 F8984</temp></blocks>
<axes><auto>yes</auto><ax1>+05361.188</ax1><ax4>+04593.924</ax4><ax5>+04593.924</ax5><sub>pos</sub></axes>
<blocks><act>N21640 G01 X4411.924 Y0 Z24 A271.771 F8984</act><auto>yes</auto><pas>N21645 G01 X4411.9 Y0 Z24 A270 F9040</pas><sub>basis</sub><temp>N21650 G01 X4411.924 Y0 Z24 A268.229</temp></blocks>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05361.047</ax1><ax4>+04590.572</ax4><ax5>+04590.572</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N21650 G01 X4411.924 Y0 Z24 A268.229</act><auto>yes</auto><pas>N21655 G01 X4411.997 Y0 Z24 A266.485 F8984</pas><sub>basis</sub><temp>N21660 G01 X4412.118 Y0 Z24 A264.766 F8872</temp></blocks>
<axes><auto>yes</auto><ax1>+05361.083</ax1><ax4>+04587.206</ax4><ax5>+04587.206</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05361.296</ax1><ax4>+04583.916</ax4><ax5>+04583.916</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05361.682</ax1><ax4>+04580.793</ax4><ax5>+04580.793</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05362.240</ax1><ax4>+04577.935</ax4><ax5>+04577.935</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05362.859</ax1><ax4>+04575.742</ax4><ax5>+04575.742</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05363.714</ax1><ax4>+04573.721</ax4><ax5>+04573.721</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05364.698</ax1><ax4>+04572.365</ax4><ax5>+04572.365</ax5><sub>pos</sub></axes>
<blocks><act>N21720 G01 X4416.408 Y0 Z24 A251.881 F3381</act><auto>yes</auto><pas>N21725 G01 X4416.9 Y0 Z24 A251.79 F3045</pas><sub>basis</sub><temp>N21730 G01 X4417.396 Y0 Z24 A251.883 F3046</temp></blocks>
<laser><act1>9812</act1><auto>yes</auto><preset1>9812</preset1></laser>
<axes><auto>yes</auto><ax1>+05365.770</ax1><ax4>+04571.813</ax4><ax5>+04571.813</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9433</act1><auto>yes</auto><preset1>9433</preset1></laser>
<blocks><act>N21730 G01 X4417.396 Y0 Z24 A251.883 F3046</act><auto>yes</auto><pas>N21735 G01 X4417.883 Y0 Z24 A252.157 F3385</pas><sub>basis</sub><temp>N21740 G01 X4418.358 Y0 Z24 A252.607 F3941</temp></blocks>
<axes><auto>yes</auto><ax1>+05366.733</ax1><ax4>+04572.075</ax4><ax5>+04572.075</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05367.789</ax1><ax4>+04573.180</ax4><ax5>+04573.180</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>5</ax4><ax5>5</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05368.733</ax1><ax4>+04575.022</ax4><ax5>+04575.022</ax5><sub>pos</sub></axes>
<blocks><act>N21760 G01 X4420.073 Y0 Z24 A256.025 F6484</act><auto>yes</auto><pas>N21765 G01 X4420.341 Y0 Z24 A256.917 F7007</pas><sub>basis</sub><temp>N21770 G01 X4420.341 Y0 Z24 A256.917 F3000</temp></blocks>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05369.441</ax1><ax4>+04576.894</ax4><ax5>+04576.894</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>2</ax4><ax5>2</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<blocks><act>N21775 G01 X4420.436 Y0 Z24 A257.234 F7007</act><auto>yes</auto><pas>N21780 G01 X4420.763 Y0 Z24 A258.556 F7467</pas><sub>basis</sub><temp>N21785 G01 X4421.053 Y0 Z24 A259.979 F7865</temp></blocks>
<axes><auto>yes</auto><ax1>+05369.747</ax1><ax4>+04578.291</ax4><ax5>+04578.291</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9996</act2><auto>yes</auto><preset1>9994</preset1><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05370.327</ax1><ax4>+04581.206</ax4><ax5>+04581.206</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<blocks><act>N21795 G01 X4421.517 Y0 Z24 A263.109 F8482</act><auto>yes</auto><pas>N21800 G01 X4421.682 Y0 Z24 A264.768 F8706</pas><sub>basis</sub><temp>N21805 G01 X4421.803 Y0 Z24 A266.484 F8873</temp></blocks>
<axes><auto>yes</auto><ax1>+05370.695</ax1><ax4>+04583.964</ax4><ax5>+04583.964</ax5><sub>pos</sub></axes>
<blocks><act>N21805 G01 X4421.803 Y0 Z24 A266.484 F8873</act><auto>yes</auto><pas>N21810 G01 X4421.876 Y0 Z24 A268.239 F8984</pas><sub>basis</sub><temp>N21815 G01 X4421.9 Y0 Z24 A270 F9040</temp></blocks>
<axes><auto>yes</auto><ax1>+05370.951</ax1><ax4>+04587.259</ax4><ax5>+04587.259</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<blocks><act>N21815 G01 X4421.9 Y0 Z24 A270 F9040</act><auto>yes</auto><pas>N21820 G01 X4421.888 Y0 Z24 A270.881</pas><sub>basis</sub><temp>N21830 Q= P11 (LASER OFF)</temp></blocks>
<axes><auto>yes</auto><ax1>+05371.033</ax1><ax4>+04590.325</ax4><ax5>+04590.325</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05371.021</ax1><ax4>+04590.881</ax4><ax5>+04590.881</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00022.236</ax3><sub>pos</sub></axes>
<blocks><act>N20980 G01 X4126.2 Y0 Z24 A200.204</act><auto>yes</auto><pas>N20980 G01 X4126.2 Y0 Z24 A200.204</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>+00024.540</ax3><sub>pos</sub></axes>
<blocks><act>N21835 G00 Z29</act><auto>yes</auto><pas>N21840 (====CONTOUR 3 ====)</pas><sub>basis</sub><temp>N21845 G00 X4691.393 Y0 A270.328</temp></blocks>
<axes><auto>yes</auto><ax3>+00028.377</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>8</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05371.051</ax1><ax3>+00029.000</ax3><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9999</act2><auto>yes</auto><preset1>9994</preset1><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+05374.188</ax1><ax4>+04590.873</ax4><ax5>+04590.873</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>10</ax1><ax3>0</ax3><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05382.345</ax1><ax4>+04590.855</ax4><ax5>+04590.855</ax5><sub>pos</sub></axes>
<blocks><act>N21845 G00 X4691.393 Y0 A270.328</act><auto>yes</auto><pas>N21850 G01 X4691.393 Y0 Z24 F10000</pas><sub>basis</sub><temp>N21855 Q= P10 (LASER ON)</temp></blocks>
<axes><auto>yes</auto><ax1>+05395.521</ax1><ax4>+04590.827</ax4><ax5>+04590.827</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>29</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05413.717</ax1><ax4>+04590.788</ax4><ax5>+04590.788</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05436.933</ax1><ax4>+04590.739</ax4><ax5>+04590.739</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>47</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05461.364</ax1><ax4>+04590.688</ax4><ax5>+04590.688</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05493.992</ax1><ax4>+04590.620</ax4><ax5>+04590.620</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>65</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05529.276</ax1><ax4>+04590.548</ax4><ax5>+04590.548</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05560.070</ax1><ax4>+04590.486</ax4><ax5>+04590.486</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>53</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05582.930</ax1><ax4>+04590.440</ax4><ax5>+04590.440</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05604.389</ax1><ax4>+04590.397</ax4><ax5>+04590.397</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>35</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05620.869</ax1><ax4>+04590.365</ax4><ax5>+04590.365</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05632.370</ax1><ax4>+04590.343</ax4><ax5>+04590.343</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>17</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05638.893</ax1><ax4>+04590.330</ax4><ax5>+04590.330</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05640.526</ax1><ax3>+00028.976</ax3><ax4>+04590.328</ax4><ax5>+04590.328</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax3>1</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax3>+00026.978</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00024.203</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>5</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<laser><act1>9412</act1><act2>9998</act2><auto>yes</auto><preset1>9412</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05641.350</ax1><ax4>+04590.449</ax4><ax5>+04590.449</ax5><sub>pos</sub></axes>
<blocks><act>N21860 G01 X4694.39 Y0 Z24 A270.705 F3021</act><auto>yes</auto><pas>N21865 G01 X4694.38 Y0 Z24 A271.411 F9040</pas><sub>basis</sub><temp>N21870 G01 X4694.322 Y0 Z24 A272.812 F8984</temp></blocks>
<axes><auto>yes</auto><ax1>+05642.469</ax1><ax4>+04590.590</ax4><ax5>+04590.590</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05643.497</ax1><ax4>+04590.705</ax4><ax5>+04590.705</ax5><sub>pos</sub></axes>
<laser><act1>7673</act1><act2>9995</act2><auto>yes</auto><preset1>7673</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05643.480</ax1><ax4>+04592.632</ax4><ax5>+04592.632</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N21875 G01 X4694.227 Y0 Z24 A274.176 F8873</act><auto>yes</auto><pas>N21880 G01 X4694.094 Y0 Z24 A275.499 F8705</pas><sub>basis</sub><temp>N21885 G01 X4693.926 Y0 Z24 A276.78 F8481</temp></blocks>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05643.269</ax1><ax4>+04595.488</ax4><ax5>+04595.488</ax5><sub>pos</sub></axes>
<blocks><act>N21885 G01 X4693.926 Y0 Z24 A276.78 F8481</act><auto>yes</auto><pas>N21890 G01 X4693.724 Y0 Z24 A277.995 F8198</pas><sub>basis</sub><temp>N21895 G01 X4693.492 Y0 Z24 A279.128 F7855</temp></blocks>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05642.824</ax1><ax4>+04598.517</ax4><ax5>+04598.517</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>12</ax4><ax5>12</ax5><sub>vel</sub></axes>
<blocks><act>N21900 G01 X4693.229 Y0 Z24 A280.18 F7450</act><auto>yes</auto><pas>N21905 G01 X4692.938 Y0 Z24 A281.14 F6981</pas><sub>basis</sub><temp>N21910 G01 X4692.625 Y0 Z24 A281.992 F6450</temp></blocks>
<axes><auto>yes</auto><ax1>+05642.162</ax1><ax4>+04601.141</ax4><ax5>+04601.141</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05641.319</ax1><ax4>+04603.068</ax4><ax5>+04603.068</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>4</ax4><ax5>4</ax5><sub>vel</sub></axes>
<laser><act1>7820</act1><act2>9995</act2><auto>yes</auto><preset1>7820</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05640.855</ax1><ax4>+04603.791</ax4><ax5>+04603.791</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9998</act2><auto>yes</auto><preset1>9994</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05639.807</ax1><ax4>+04604.453</ax4><ax5>+04604.453</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9431</act1><auto>yes</auto><preset1>9431</preset1></laser>
<axes><auto>yes</auto><ax1>+05638.844</ax1><ax4>+04604.154</ax4><ax5>+04604.154</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05637.817</ax1><ax4>+04602.821</ax4><ax5>+04602.821</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05636.946</ax1><ax4>+04600.666</ax4><ax5>+04600.666</ax5><sub>pos</sub></axes>
<blocks><act>N21995 G01 X4687.309 Y0 Z24 A279.131 F7450</act><auto>yes</auto><pas>N22000 G01 X4687.075 Y0 Z24 A277.99 F7854</pas><sub>basis</sub><temp>N22005 G01 X4686.874 Y0 Z24 A276.781 F8198</temp></blocks>
<axes><auto>yes</auto><ax1>+05636.272</ax1><ax4>+04597.941</ax4><ax5>+04597.941</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>12</ax4><ax5>12</ax5><sub>vel</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<blocks><act>N22005 G01 X4686.874 Y0 Z24 A276.781 F8198</act><auto>yes</auto><pas>N22010 G01 X4686.706 Y0 Z24 A275.507 F8481</pas><sub>basis</sub><temp>N22015 G01 X4686.573 Y0 Z24 A274.174 F8706</temp></blocks>
<axes><auto>yes</auto><ax1>+05635.856</ax1><ax4>+04595.237</ax4><ax5>+04595.237</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05635.592</ax1><ax4>+04591.946</ax4><ax5>+04591.946</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05635.548</ax1><ax4>+04588.577</ax4><ax5>+04588.577</ax5><sub>pos</sub></axes>
<blocks><act>N22040 G01 X4686.478 Y0 Z24 A267.189 F8984</act><auto>yes</auto><pas>N22045 G01 X4686.574 Y0 Z24 A265.815 F8872</pas><sub>basis</sub><temp>N22050 G01 X4686.707 Y0 Z24 A264.484 F8704</temp></blocks>
<axes><auto>yes</auto><ax1>+05635.721</ax1><ax4>+04585.270</ax4><ax5>+04585.270</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<blocks><act>N22055 G01 X4686.875 Y0 Z24 A263.215 F8479</act><auto>yes</auto><pas>N22060 G01 X4687.077 Y0 Z24 A262.003 F8196</pas><sub>basis</sub><temp>N22065 G01 X4687.311 Y0 Z24 A260.862 F7852</temp></blocks>
<axes><auto>yes</auto><ax1>+05636.121</ax1><ax4>+04582.165</ax4><ax5>+04582.165</ax5><sub>pos</sub></axes>
<blocks><act>N22065 G01 X4687.311 Y0 Z24 A260.862 F7852</act><auto>yes</auto><pas>N22070 G01 X4687.573 Y0 Z24 A259.812 F7447</pas><sub>basis</sub><temp>N22075 G01 X4687.863 Y0 Z24 A258.857 F6978</temp></blocks>
<axes><auto>yes</auto><ax1>+05636.725</ax1><ax4>+04579.440</ax4><ax5>+04579.440</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>10</ax4><ax5>10</ax5><sub>vel</sub></axes>
<blocks><act>N22080 G01 X4688.178 Y0 Z24 A258.002 F6447</act><auto>yes</auto><pas>N22085 G01 X4688.514 Y0 Z24 A257.264 F5855</pas><sub>basis</sub><temp>N22090 G01 X4688.867 Y0 Z24 A256.65 F5215</temp></blocks>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05637.427</ax1><ax4>+04577.500</ax4><ax5>+04577.500</ax5><sub>pos</sub></axes>
<blocks><act>N22090 G01 X4688.867 Y0 Z24 A256.65 F5215</act><auto>yes</auto><pas>N22095 G01 X4689.238 Y0 Z24 A256.16 F4550</pas><sub>basis</sub><temp>N22100 G01 X4689.62 Y0 Z24 A255.806 F3909</temp></blocks>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05638.375</ax1><ax4>+04576.033</ax4><ax5>+04576.033</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>4</ax4><ax5>4</ax5><sub>vel</sub></axes>
<laser><act1>9839</act1><auto>yes</auto><preset1>9839</preset1></laser>
<blocks><act>N22105 G01 X4690.008 Y0 Z24 A255.593 F3366</act><auto>yes</auto><pas>N22110 G01 X4690.401 Y0 Z24 A255.522 F3043</pas><sub>basis</sub><temp>N22115 G01 X4690.796 Y0 Z24 A255.595 F3045</temp></blocks>
<axes><auto>yes</auto><ax1>+05639.446</ax1><ax4>+04575.530</ax4><ax5>+04575.530</ax5><sub>pos</sub></axes>
<laser><act1>9432</act1><auto>yes</auto><preset1>9432</preset1></laser>
<axes><auto>yes</auto><ax1>+05640.534</ax1><ax4>+04576.135</ax4><ax5>+04576.135</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><sub>vel</sub></axes>
<blocks><act>N22130 G01 X4691.936 Y0 Z24 A256.654 F4557</act><auto>yes</auto><pas>N22135 G01 X4692.291 Y0 Z24 A257.272 F5222</pas><sub>basis</sub><temp>N22140 G01 X4692.407 Y0 Z24 A257.529 F5861</temp></blocks>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05641.390</ax1><ax4>+04577.354</ax4><ax5>+04577.354</ax5><sub>pos</sub></axes>
<blocks><act>N22140 G01 X4692.407 Y0 Z24 A257.529 F5861</act><auto>yes</auto><pas>N22145 G01 X4692.407 Y0 Z24 A257.529</pas><sub>basis</sub><temp>N22150 G01 X4692.625 Y0 Z24 A258.008 F5861</temp></blocks>
<laser><act1>8023</act1><act2>9995</act2><auto>yes</auto><preset1>8023</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05641.715</ax1><ax4>+04578.176</ax4><ax5>+04578.176</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>9</ax4><ax5>9</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><act2>9997</act2><auto>yes</auto><preset1>9994</preset1><preset2>9997</preset2></laser>
<blocks><act>N22160 G01 X4693.23 Y0 Z24 A259.822 F6981</act><auto>yes</auto><pas>N22165 G01 X4693.491 Y0 Z24 A260.871 F7450</pas><sub>basis</sub><temp>N22170 G01 X4693.724 Y0 Z24 A262.006 F7855</temp></blocks>
<axes><auto>yes</auto><ax1>+05642.486</ax1><ax4>+04580.652</ax4><ax5>+04580.652</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05643.056</ax1><ax4>+04583.594</ax4><ax5>+04583.594</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05643.401</ax1><ax4>+04586.819</ax4><ax5>+04586.819</ax5><sub>pos</sub></axes>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05643.529</ax1><ax4>+04589.934</ax4><ax5>+04589.934</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>9</ax4><ax5>9</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05643.523</ax1><ax4>+04590.705</ax4><ax5>+04590.705</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax1>0</ax1><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00022.101</ax3><sub>pos</sub></axes>
<blocks><act>N20980 G01 X4126.2 Y0 Z24 A200.204</act><auto>yes</auto><pas>N20980 G01 X4126.2 Y0 Z24 A200.204</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N22220 G00 Z29</act><auto>yes</auto><pas>N22225 Q990036</pas><sub>basis</sub><temp>N1010 (UNLOAD ONE ENTER)</temp></blocks>
<axes><auto>yes</auto><ax3>+00022.741</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>8</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00026.720</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00028.968</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>2</ax3><sub>vel</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>+00029.000</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N1060 G4 F200</act><auto>yes</auto><pas>N1070 M02</pas><sub>basis</sub><temp>N22230 (====CONTOUR 4 ====)</temp></blocks>
<axes><auto>yes</auto><ax4>+04590.749</ax4><ax5>+04590.749</ax5><sub>pos</sub></axes>
<laser><act1>9308</act1><auto>yes</auto><preset1>9308</preset1></laser>
<axes><auto>yes</auto><ax1>+05643.819</ax1><ax4>+04592.699</ax4><ax5>+04592.699</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><act2>9997</act2><auto>yes</auto><preset1>9994</preset1><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05644.640</ax1><ax4>+04597.484</ax4><ax5>+04597.484</ax5><sub>pos</sub></axes>
<blocks><act>N22235 G00 X4710.6 Y0 A358.163</act><auto>yes</auto><pas>N22240 G01 X4710.6 Y0 Z24 F10000</pas><sub>basis</sub><temp>N22245 Q= P10 (LASER ON)</temp></blocks>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05645.987</ax1><ax4>+04605.105</ax4><ax5>+04605.105</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>38</ax4><ax5>38</ax5><sub>vel</sub></axes>
<laser><act2>9999</act2><auto>yes</auto><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+05647.597</ax1><ax4>+04614.099</ax4><ax5>+04614.099</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05649.929</ax1><ax4>+04627.037</ax4><ax5>+04627.037</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>4</ax1><ax4>61</ax4><ax5>61</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05652.688</ax1><ax4>+04641.931</ax4><ax5>+04641.931</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05655.105</ax1><ax4>+04654.628</ax4><ax5>+04654.628</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>49</ax4><ax5>49</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05657.016</ax1><ax4>+04664.596</ax4><ax5>+04664.596</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05658.273</ax1><ax4>+04671.081</ax4><ax5>+04671.081</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>27</ax4><ax5>27</ax5><sub>vel</sub></axes>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05659.235</ax1><ax4>+04675.933</ax4><ax5>+04675.933</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05659.692</ax1><ax4>+04678.057</ax4><ax5>+04678.057</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>3</ax4><ax5>3</ax5><sub>vel</sub></axes>
<laser><act1>7443</act1><act2>9995</act2><auto>yes</auto><preset1>7443</preset1><preset2>9995</preset2></laser>
<blocks><act>N22240 G01 X4710.6 Y0 Z24 F10000</act><auto>yes</auto><pas>N22245 Q= P10 (LASER ON)</pas><sub>basis</sub><temp>N1010 G10 (STA-PO-PI-LI-P6018DA_A1)</temp></blocks>
<axes><auto>yes</auto><ax1>+05659.733</ax1><ax3>+00028.496</ax3><ax4>+04678.163</ax4><ax5>+04678.163</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00025.218</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax3>11</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>N1050 Q999992</temp></blocks>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><auto>yes</auto><pas>N2250 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<axes><auto>yes</auto><ax1>+05659.469</ax1><sub>pos</sub></axes>
<blocks><act>N22250 G01 X4707.6 Y0 Z24 A358.163 F3000</act><auto>yes</auto><pas>N22255 G01 X4707.6 Y0 Z24 A356.327 F9047</pas><sub>basis</sub><temp>N22260 G01 X4707.6 Y0 Z24 A354.49</temp></blocks>
<laser><act1>9395</act1><act2>9998</act2><auto>yes</auto><preset1>9395</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05658.349</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05657.369</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05656.733</ax1><ax4>+04677.997</ax4><ax5>+04677.997</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>5</ax4><ax5>5</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><act2>9995</act2><auto>yes</auto><preset1>9994</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax4>+04674.961</ax4><ax5>+04674.961</ax5><sub>pos</sub></axes>
<blocks><act>N22265 G01 X4707.6 Y0 Z24 A352.653</act><auto>yes</auto><pas>N22270 G01 X4707.6 Y0 Z24 A350.816</pas><sub>basis</sub><temp>N22275 G01 X4707.6 Y0 Z24 A348.98</temp></blocks>
<axes><auto>yes</auto><ax4>+04671.583</ax4><ax5>+04671.583</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N22275 G01 X4707.6 Y0 Z24 A348.98</act><auto>yes</auto><pas>N22280 G01 X4707.6 Y0 Z24 A347.143</pas><sub>basis</sub><temp>N22285 G01 X4707.6 Y0 Z24 A345.306</temp></blocks>
<axes><auto>yes</auto><ax4>+04668.206</ax4><ax5>+04668.206</ax5><sub>pos</sub></axes>
<blocks><act>N22285 G01 X4707.6 Y0 Z24 A345.306</act><auto>yes</auto><pas>N22290 G01 X4707.6 Y0 Z24 A343.085</pas><sub>basis</sub><temp>N22295 G01 X4707.6 Y0 Z24 A343.083</temp></blocks>
<axes><auto>yes</auto><ax4>+04665.250</ax4><ax5>+04665.250</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04663.087</ax4><ax5>+04663.087</ax5><sub>pos</sub></axes>
<blocks><act>N22300 G01 X4707.6 Y0 Z24 A341.633</act><auto>yes</auto><pas>N22305 G01 X4707.6 Y0 Z24 A339.796</pas><sub>basis</sub><temp>N22310 G01 X4707.6 Y0 Z24 A337.959</temp></blocks>
<laser><act1>7573</act1><auto>yes</auto><preset1>7573</preset1></laser>
<axes><auto>yes</auto><ax4>+04661.162</ax4><ax5>+04661.162</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N22305 G01 X4707.6 Y0 Z24 A339.796</act><auto>yes</auto><pas>N22310 G01 X4707.6 Y0 Z24 A337.959</pas><sub>basis</sub><temp>N22315 G01 X4707.6 Y0 Z24 A336.122</temp></blocks>
<axes><auto>yes</auto><ax4>+04657.785</ax4><ax5>+04657.785</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04654.407</ax4><ax5>+04654.407</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04651.452</ax4><ax5>+04651.452</ax5><sub>pos</sub></axes>
<blocks><act>N22335 G01 X4707.6 Y0 Z24 A328.775</act><auto>yes</auto><pas>N22340 G01 X4707.6 Y0 Z24 A326.939</pas><sub>basis</sub><temp>N22345 G01 X4707.6 Y0 Z24 A325.102</temp></blocks>
<axes><auto>yes</auto><ax4>+04648.074</ax4><ax5>+04648.074</ax5><sub>pos</sub></axes>
<blocks><act>N22345 G01 X4707.6 Y0 Z24 A325.102</act><auto>yes</auto><pas>N22350 G01 X4707.6 Y0 Z24 A323.265</pas><sub>basis</sub><temp>N22355 G01 X4707.6 Y0 Z24 A321.429</temp></blocks>
<axes><auto>yes</auto><ax4>+04644.697</ax4><ax5>+04644.697</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04641.319</ax4><ax5>+04641.319</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04637.942</ax4><ax5>+04637.942</ax5><sub>pos</sub></axes>
<blocks><act>N22370 G01 X4707.6 Y0 Z24 A315.918</act><auto>yes</auto><pas>N22375 G01 X4707.6 Y0 Z24 A314.082</pas><sub>basis</sub><temp>N22380 G01 X4707.6 Y0 Z24 A312.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04634.564</ax4><ax5>+04634.564</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04631.609</ax4><ax5>+04631.609</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04628.231</ax4><ax5>+04628.231</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04624.854</ax4><ax5>+04624.854</ax5><sub>pos</sub></axes>
<blocks><act>N22405 G01 X4707.6 Y0 Z24 A303.061</act><auto>yes</auto><pas>N22410 G01 X4707.6 Y0 Z24 A301.225</pas><sub>basis</sub><temp>N22415 G01 X4707.6 Y0 Z24 A299.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04621.476</ax4><ax5>+04621.476</ax5><sub>pos</sub></axes>
<blocks><act>N22415 G01 X4707.6 Y0 Z24 A299.388</act><auto>yes</auto><pas>N22420 G01 X4707.6 Y0 Z24 A297.551</pas><sub>basis</sub><temp>N22425 G01 X4707.6 Y0 Z24 A295.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04618.521</ax4><ax5>+04618.521</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04615.143</ax4><ax5>+04615.143</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04611.766</ax4><ax5>+04611.766</ax5><sub>pos</sub></axes>
<blocks><act>N22440 G01 X4707.6 Y0 Z24 A290.204</act><auto>yes</auto><pas>N22445 G01 X4707.6 Y0 Z24 A288.367</pas><sub>basis</sub><temp>N22450 G01 X4707.6 Y0 Z24 A286.531</temp></blocks>
<axes><auto>yes</auto><ax4>+04608.388</ax4><ax5>+04608.388</ax5><sub>pos</sub></axes>
<blocks><act>N22450 G01 X4707.6 Y0 Z24 A286.531</act><auto>yes</auto><pas>N22455 G01 X4707.6 Y0 Z24 A284.694</pas><sub>basis</sub><temp>N22460 G01 X4707.6 Y0 Z24 A282.857</temp></blocks>
<axes><auto>yes</auto><ax4>+04605.011</ax4><ax5>+04605.011</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04601.633</ax4><ax5>+04601.633</ax5><sub>pos</sub></axes>
<blocks><act>N22470 G01 X4707.6 Y0 Z24 A279.184</act><auto>yes</auto><pas>N22475 G01 X4707.6 Y0 Z24 A277.347</pas><sub>basis</sub><temp>N22480 G01 X4707.6 Y0 Z24 A275.51</temp></blocks>
<axes><auto>yes</auto><ax4>+04598.678</ax4><ax5>+04598.678</ax5><sub>pos</sub></axes>
<blocks><act>N22475 G01 X4707.6 Y0 Z24 A277.347</act><auto>yes</auto><pas>N22480 G01 X4707.6 Y0 Z24 A275.51</pas><sub>basis</sub><temp>N22485 G01 X4707.6 Y0 Z24 A273.673</temp></blocks>
<axes><auto>yes</auto><ax4>+04595.300</ax4><ax5>+04595.300</ax5><sub>pos</sub></axes>
<blocks><act>N22485 G01 X4707.6 Y0 Z24 A273.673</act><auto>yes</auto><pas>N22490 G01 X4707.6 Y0 Z24 A271.837</pas><sub>basis</sub><temp>N22495 G01 X4707.6 Y0 Z24 A270</temp></blocks>
<axes><auto>yes</auto><ax4>+04591.923</ax4><ax5>+04591.923</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04588.545</ax4><ax5>+04588.545</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04585.167</ax4><ax5>+04585.167</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04582.212</ax4><ax5>+04582.212</ax5><sub>pos</sub></axes>
<blocks><act>N22520 G01 X4707.6 Y0 Z24 A260.816</act><auto>yes</auto><pas>N22525 G01 X4707.6 Y0 Z24 A258.98</pas><sub>basis</sub><temp>N22530 G01 X4707.6 Y0 Z24 A257.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04578.835</ax4><ax5>+04578.835</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04575.457</ax4><ax5>+04575.457</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04572.080</ax4><ax5>+04572.080</ax5><sub>pos</sub></axes>
<blocks><act>N22550 G01 X4707.6 Y0 Z24 A249.796</act><auto>yes</auto><pas>N22555 G01 X4707.6 Y0 Z24 A247.959</pas><sub>basis</sub><temp>N22560 G01 X4707.6 Y0 Z24 A246.122</temp></blocks>
<axes><auto>yes</auto><ax4>+04568.702</ax4><ax5>+04568.702</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04565.747</ax4><ax5>+04565.747</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04562.369</ax4><ax5>+04562.369</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04558.992</ax4><ax5>+04558.992</ax5><sub>pos</sub></axes>
<blocks><act>N22585 G01 X4707.6 Y0 Z24 A236.939</act><auto>yes</auto><pas>N22590 G01 X4707.6 Y0 Z24 A235.102</pas><sub>basis</sub><temp>N22595 G01 X4707.6 Y0 Z24 A233.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04555.614</ax4><ax5>+04555.614</ax5><sub>pos</sub></axes>
<blocks><act>N22595 G01 X4707.6 Y0 Z24 A233.265</act><auto>yes</auto><pas>N22600 G01 X4707.6 Y0 Z24 A231.429</pas><sub>basis</sub><temp>N22605 G01 X4707.6 Y0 Z24 A229.592</temp></blocks>
<axes><auto>yes</auto><ax4>+04552.236</ax4><ax5>+04552.236</ax5><sub>pos</sub></axes>
<blocks><act>N22605 G01 X4707.6 Y0 Z24 A229.592</act><auto>yes</auto><pas>N22610 G01 X4707.6 Y0 Z24 A227.755</pas><sub>basis</sub><temp>N22615 G01 X4707.6 Y0 Z24 A225.918</temp></blocks>
<axes><auto>yes</auto><ax4>+04549.281</ax4><ax5>+04549.281</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04545.904</ax4><ax5>+04545.904</ax5><sub>pos</sub></axes>
<blocks><act>N22620 G01 X4707.6 Y0 Z24 A224.082</act><auto>yes</auto><pas>N22625 G01 X4707.6 Y0 Z24 A222.245</pas><sub>basis</sub><temp>N22630 G01 X4707.6 Y0 Z24 A220.408</temp></blocks>
<axes><auto>yes</auto><ax4>+04542.526</ax4><ax5>+04542.526</ax5><sub>pos</sub></axes>
<blocks><act>N22630 G01 X4707.6 Y0 Z24 A220.408</act><auto>yes</auto><pas>N22635 G01 X4707.6 Y0 Z24 A218.571</pas><sub>basis</sub><temp>N22640 G01 X4707.6 Y0 Z24 A216.735</temp></blocks>
<axes><auto>yes</auto><ax4>+04539.148</ax4><ax5>+04539.148</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04535.771</ax4><ax5>+04535.771</ax5><sub>pos</sub></axes>
<blocks><act>N22650 G01 X4707.6 Y0 Z24 A213.061</act><auto>yes</auto><pas>N22655 G01 X4707.6 Y0 Z24 A211.225</pas><sub>basis</sub><temp>N22660 G01 X4707.6 Y0 Z24 A209.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04532.816</ax4><ax5>+04532.816</ax5><sub>pos</sub></axes>
<blocks><act>N22655 G01 X4707.6 Y0 Z24 A211.225</act><auto>yes</auto><pas>N22660 G01 X4707.6 Y0 Z24 A209.388</pas><sub>basis</sub><temp>N22665 G01 X4707.6 Y0 Z24 A207.551</temp></blocks>
<axes><auto>yes</auto><ax4>+04529.438</ax4><ax5>+04529.438</ax5><sub>pos</sub></axes>
<blocks><act>N22665 G01 X4707.6 Y0 Z24 A207.551</act><auto>yes</auto><pas>N22670 G01 X4707.6 Y0 Z24 A205.714</pas><sub>basis</sub><temp>N22675 G01 X4707.6 Y0 Z24 A203.878</temp></blocks>
<axes><auto>yes</auto><ax4>+04526.060</ax4><ax5>+04526.060</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04522.683</ax4><ax5>+04522.683</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04519.305</ax4><ax5>+04519.305</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04516.350</ax4><ax5>+04516.350</ax5><sub>pos</sub></axes>
<blocks><act>N22700 G01 X4707.6 Y0 Z24 A194.694</act><auto>yes</auto><pas>N22705 G01 X4707.6 Y0 Z24 A192.857</pas><sub>basis</sub><temp>N22710 G01 X4707.6 Y0 Z24 A191.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04512.972</ax4><ax5>+04512.972</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04509.595</ax4><ax5>+04509.595</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04506.217</ax4><ax5>+04506.217</ax5><sub>pos</sub></axes>
<blocks><act>N22730 G01 X4707.6 Y0 Z24 A183.673</act><auto>yes</auto><pas>N22735 G01 X4707.6 Y0 Z24 A181.837</pas><sub>basis</sub><temp>N22740 G01 X4707.6 Y0 Z24 A180</temp></blocks>
<axes><auto>yes</auto><ax4>+04502.840</ax4><ax5>+04502.840</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04499.884</ax4><ax5>+04499.884</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04496.507</ax4><ax5>+04496.507</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04493.129</ax4><ax5>+04493.129</ax5><sub>pos</sub></axes>
<blocks><act>N22765 G01 X4707.6 Y0 Z24 A170.816</act><auto>yes</auto><pas>N22770 G01 X4707.6 Y0 Z24 A168.98</pas><sub>basis</sub><temp>N22775 G01 X4707.6 Y0 Z24 A167.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04489.752</ax4><ax5>+04489.752</ax5><sub>pos</sub></axes>
<blocks><act>N22775 G01 X4707.6 Y0 Z24 A167.143</act><auto>yes</auto><pas>N22780 G01 X4707.6 Y0 Z24 A165.306</pas><sub>basis</sub><temp>N22785 G01 X4707.6 Y0 Z24 A163.469</temp></blocks>
<axes><auto>yes</auto><ax4>+04486.374</ax4><ax5>+04486.374</ax5><sub>pos</sub></axes>
<blocks><act>N22785 G01 X4707.6 Y0 Z24 A163.469</act><auto>yes</auto><pas>N22790 G01 X4707.6 Y0 Z24 A161.633</pas><sub>basis</sub><temp>N22795 G01 X4707.6 Y0 Z24 A159.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04482.997</ax4><ax5>+04482.997</ax5><sub>pos</sub></axes>
<blocks><act>N22790 G01 X4707.6 Y0 Z24 A161.633</act><auto>yes</auto><pas>N22795 G01 X4707.6 Y0 Z24 A159.796</pas><sub>basis</sub><temp>N22800 G01 X4707.6 Y0 Z24 A157.959</temp></blocks>
<axes><auto>yes</auto><ax4>+04480.041</ax4><ax5>+04480.041</ax5><sub>pos</sub></axes>
<blocks><act>N22800 G01 X4707.6 Y0 Z24 A157.959</act><auto>yes</auto><pas>N22805 G01 X4707.6 Y0 Z24 A156.122</pas><sub>basis</sub><temp>N22810 G01 X4707.6 Y0 Z24 A154.286</temp></blocks>
<axes><auto>yes</auto><ax4>+04476.664</ax4><ax5>+04476.664</ax5><sub>pos</sub></axes>
<blocks><act>N22810 G01 X4707.6 Y0 Z24 A154.286</act><auto>yes</auto><pas>N22815 G01 X4707.6 Y0 Z24 A152.449</pas><sub>basis</sub><temp>N22820 G01 X4707.6 Y0 Z24 A150.612</temp></blocks>
<axes><auto>yes</auto><ax4>+04473.286</ax4><ax5>+04473.286</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04469.909</ax4><ax5>+04469.909</ax5><sub>pos</sub></axes>
<blocks><act>N22830 G01 X4707.6 Y0 Z24 A146.939</act><auto>yes</auto><pas>N22835 G01 X4707.6 Y0 Z24 A145.102</pas><sub>basis</sub><temp>N22840 G01 X4707.6 Y0 Z24 A143.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04466.531</ax4><ax5>+04466.531</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04463.576</ax4><ax5>+04463.576</ax5><sub>pos</sub></axes>
<blocks><act>N22845 G01 X4707.6 Y0 Z24 A141.429</act><auto>yes</auto><pas>N22850 G01 X4707.6 Y0 Z24 A139.592</pas><sub>basis</sub><temp>N22855 G01 X4707.6 Y0 Z24 A137.755</temp></blocks>
<axes><auto>yes</auto><ax4>+04460.198</ax4><ax5>+04460.198</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04456.821</ax4><ax5>+04456.821</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04453.443</ax4><ax5>+04453.443</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04450.066</ax4><ax5>+04450.066</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04447.110</ax4><ax5>+04447.110</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04443.733</ax4><ax5>+04443.733</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04440.355</ax4><ax5>+04440.355</ax5><sub>pos</sub></axes>
<blocks><act>N22910 G01 X4707.6 Y0 Z24 A117.551</act><auto>yes</auto><pas>N22915 G01 X4707.6 Y0 Z24 A115.714</pas><sub>basis</sub><temp>N22920 G01 X4707.6 Y0 Z24 A113.878</temp></blocks>
<axes><auto>yes</auto><ax4>+04436.978</ax4><ax5>+04436.978</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04433.600</ax4><ax5>+04433.600</ax5><sub>pos</sub></axes>
<blocks><act>N22925 G01 X4707.6 Y0 Z24 A112.041</act><auto>yes</auto><pas>N22930 G01 X4707.6 Y0 Z24 A110.204</pas><sub>basis</sub><temp>N22935 G01 X4707.6 Y0 Z24 A108.367</temp></blocks>
<axes><auto>yes</auto><ax4>+04430.645</ax4><ax5>+04430.645</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04427.267</ax4><ax5>+04427.267</ax5><sub>pos</sub></axes>
<blocks><act>N22945 G01 X4707.6 Y0 Z24 A104.694</act><auto>yes</auto><pas>N22950 G01 X4707.6 Y0 Z24 A102.857</pas><sub>basis</sub><temp>N22955 G01 X4707.6 Y0 Z24 A101.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04423.890</ax4><ax5>+04423.890</ax5><sub>pos</sub></axes>
<blocks><act>N22955 G01 X4707.6 Y0 Z24 A101.02</act><auto>yes</auto><pas>N22960 G01 X4707.6 Y0 Z24 A99.184</pas><sub>basis</sub><temp>N22965 G01 X4707.6 Y0 Z24 A97.347</temp></blocks>
<axes><auto>yes</auto><ax4>+04420.512</ax4><ax5>+04420.512</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04417.135</ax4><ax5>+04417.135</ax5><sub>pos</sub></axes>
<blocks><act>N22970 G01 X4707.6 Y0 Z24 A95.51</act><auto>yes</auto><pas>N22975 G01 X4707.6 Y0 Z24 A93.673</pas><sub>basis</sub><temp>N22980 G01 X4707.6 Y0 Z24 A91.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04414.179</ax4><ax5>+04414.179</ax5><sub>pos</sub></axes>
<blocks><act>N22980 G01 X4707.6 Y0 Z24 A91.837</act><auto>yes</auto><pas>N22985 G01 X4707.6 Y0 Z24 A90</pas><sub>basis</sub><temp>N22990 G01 X4707.6 Y0 Z24 A88.163</temp></blocks>
<axes><auto>yes</auto><ax4>+04410.802</ax4><ax5>+04410.802</ax5><sub>pos</sub></axes>
<blocks><act>N22990 G01 X4707.6 Y0 Z24 A88.163</act><auto>yes</auto><pas>N22995 G01 X4707.6 Y0 Z24 A86.327</pas><sub>basis</sub><temp>N23000 G01 X4707.6 Y0 Z24 A84.49</temp></blocks>
<axes><auto>yes</auto><ax4>+04407.424</ax4><ax5>+04407.424</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04404.047</ax4><ax5>+04404.047</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04400.669</ax4><ax5>+04400.669</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04397.714</ax4><ax5>+04397.714</ax5><sub>pos</sub></axes>
<blocks><act>N23025 G01 X4707.6 Y0 Z24 A75.306</act><auto>yes</auto><pas>N23030 G01 X4707.6 Y0 Z24 A73.469</pas><sub>basis</sub><temp>N23035 G01 X4707.6 Y0 Z24 A71.633</temp></blocks>
<axes><auto>yes</auto><ax4>+04394.336</ax4><ax5>+04394.336</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04390.959</ax4><ax5>+04390.959</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04387.581</ax4><ax5>+04387.581</ax5><sub>pos</sub></axes>
<blocks><act>N23050 G01 X4707.6 Y0 Z24 A66.122</act><auto>yes</auto><pas>N23055 G01 X4707.6 Y0 Z24 A64.286</pas><sub>basis</sub><temp>N23060 G01 X4707.6 Y0 Z24 A62.449</temp></blocks>
<axes><auto>yes</auto><ax4>+04384.203</ax4><ax5>+04384.203</ax5><sub>pos</sub></axes>
<blocks><act>N23060 G01 X4707.6 Y0 Z24 A62.449</act><auto>yes</auto><pas>N23065 G01 X4707.6 Y0 Z24 A60.612</pas><sub>basis</sub><temp>N23070 G01 X4707.6 Y0 Z24 A58.775</temp></blocks>
<axes><auto>yes</auto><ax4>+04381.248</ax4><ax5>+04381.248</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04377.871</ax4><ax5>+04377.871</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04374.493</ax4><ax5>+04374.493</ax5><sub>pos</sub></axes>
<blocks><act>N23090 G01 X4707.6 Y0 Z24 A51.429</act><auto>yes</auto><pas>N23095 G01 X4707.6 Y0 Z24 A49.592</pas><sub>basis</sub><temp>N23100 G01 X4707.6 Y0 Z24 A47.755</temp></blocks>
<axes><auto>yes</auto><ax4>+04371.115</ax4><ax5>+04371.115</ax5><sub>pos</sub></axes>
<blocks><act>N23095 G01 X4707.6 Y0 Z24 A49.592</act><auto>yes</auto><pas>N23100 G01 X4707.6 Y0 Z24 A47.755</pas><sub>basis</sub><temp>N23105 G01 X4707.6 Y0 Z24 A45.918</temp></blocks>
<axes><auto>yes</auto><ax4>+04367.738</ax4><ax5>+04367.738</ax5><sub>pos</sub></axes>
<blocks><act>N23105 G01 X4707.6 Y0 Z24 A45.918</act><auto>yes</auto><pas>N23110 G01 X4707.6 Y0 Z24 A44.082</pas><sub>basis</sub><temp>N23115 G01 X4707.6 Y0 Z24 A42.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04364.783</ax4><ax5>+04364.783</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04361.405</ax4><ax5>+04361.405</ax5><sub>pos</sub></axes>
<blocks><act>N23125 G01 X4707.6 Y0 Z24 A38.571</act><auto>yes</auto><pas>N23130 G01 X4707.6 Y0 Z24 A36.735</pas><sub>basis</sub><temp>N23135 G01 X4707.6 Y0 Z24 A34.898</temp></blocks>
<axes><auto>yes</auto><ax4>+04358.027</ax4><ax5>+04358.027</ax5><sub>pos</sub></axes>
<blocks><act>N23130 G01 X4707.6 Y0 Z24 A36.735</act><auto>yes</auto><pas>N23135 G01 X4707.6 Y0 Z24 A34.898</pas><sub>basis</sub><temp>N23140 G01 X4707.6 Y0 Z24 A33.061</temp></blocks>
<axes><auto>yes</auto><ax4>+04354.650</ax4><ax5>+04354.650</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04351.272</ax4><ax5>+04351.272</ax5><sub>pos</sub></axes>
<blocks><act>N23150 G01 X4707.6 Y0 Z24 A29.388</act><auto>yes</auto><pas>N23155 G01 X4707.6 Y0 Z24 A27.551</pas><sub>basis</sub><temp>N23160 G01 X4707.6 Y0 Z24 A25.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04348.317</ax4><ax5>+04348.317</ax5><sub>pos</sub></axes>
<blocks><act>N23160 G01 X4707.6 Y0 Z24 A25.714</act><auto>yes</auto><pas>N23165 G01 X4707.6 Y0 Z24 A23.878</pas><sub>basis</sub><temp>N23170 G01 X4707.6 Y0 Z24 A22.041</temp></blocks>
<axes><auto>yes</auto><ax4>+04344.939</ax4><ax5>+04344.939</ax5><sub>pos</sub></axes>
<blocks><act>N23170 G01 X4707.6 Y0 Z24 A22.041</act><auto>yes</auto><pas>N23175 G01 X4707.6 Y0 Z24 A20.204</pas><sub>basis</sub><temp>N23180 G01 X4707.6 Y0 Z24 A18.367</temp></blocks>
<axes><auto>yes</auto><ax4>+04341.562</ax4><ax5>+04341.562</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04338.184</ax4><ax5>+04338.184</ax5><sub>pos</sub></axes>
<blocks><act>N23185 G01 X4707.6 Y0 Z24 A16.531</act><auto>yes</auto><pas>N23190 G01 X4707.6 Y0 Z24 A14.694</pas><sub>basis</sub><temp>N23195 G01 X4707.6 Y0 Z24 A13.243</temp></blocks>
<axes><auto>yes</auto><ax4>+04334.831</ax4><ax5>+04334.831</ax5><sub>pos</sub></axes>
<blocks><act>N23195 G01 X4707.6 Y0 Z24 A13.243</act><auto>yes</auto><pas>N23200 G01 X4707.6 Y0 Z24 A13.242</pas><sub>basis</sub><temp>N23205 G01 X4707.6 Y0 Z24 A11.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04333.243</ax4><ax5>+04333.243</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>8074</act1><auto>yes</auto><preset1>8074</preset1></laser>
<blocks><act>N23205 G01 X4707.6 Y0 Z24 A11.02</act><auto>yes</auto><pas>N23210 G01 X4707.6 Y0 Z24 A9.184</pas><sub>basis</sub><temp>N23215 G01 X4707.6 Y0 Z24 A7.347</temp></blocks>
<axes><auto>yes</auto><ax4>+04331.186</ax4><ax5>+04331.186</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04327.809</ax4><ax5>+04327.809</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04324.431</ax4><ax5>+04324.431</ax5><sub>pos</sub></axes>
<blocks><act>N23230 G01 X4707.6 Y0 Z24 A1.837</act><auto>yes</auto><pas>N23235 G01 X4707.6 Y0 Z24 A0</pas><sub>basis</sub><temp>N23240 G01 X4707.6 Y0 Z24 A-1.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04321.054</ax4><ax5>+04321.054</ax5><sub>pos</sub></axes>
<blocks><act>N23240 G01 X4707.6 Y0 Z24 A-1.837</act><auto>yes</auto><pas>N23250 Q= P11 (LASER OFF)</pas><sub>basis</sub><temp>N1010 G10 (END-LI-PO_W)</temp></blocks>
<axes><auto>yes</auto><ax4>+04318.497</ax4><ax5>+04318.497</ax5><sub>pos</sub></axes>
<laser><act1>8866</act1><auto>yes</auto><preset1>8866</preset1></laser>
<axes><auto>yes</auto><ax4>+04318.163</ax4><ax5>+04318.163</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00022.025</ax3><sub>pos</sub></axes>
<blocks><act>N23140 G01 X4707.6 Y0 Z24 A33.061</act><auto>yes</auto><pas>N23140 G01 X4707.6 Y0 Z24 A33.061</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1070 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<axes><auto>yes</auto><ax3>+00024.329</ax3><sub>pos</sub></axes>
<blocks><act>N23255 G00 Z29</act><auto>yes</auto><pas>N23260 Q990039</pas><sub>basis</sub><temp>N1010 G10 (LCUT-END)</temp></blocks>
<axes><auto>yes</auto><ax3>+00028.233</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>9</ax3><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00029.000</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N1050 M223 (UD SP BOARD DOWN)</act><auto>yes</auto><pas>N1060 G4 F200</pas><sub>basis</sub><temp>N1070 G10</temp></blocks>
<blocks><act>N1060 G4 F200</act><auto>yes</auto><pas>N1070 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N1110 (M255 CUT OFF CHECK)</pas><sub>basis</sub><temp>N1120 G10</temp></blocks>
<blocks><act>N1160 (M254 LASER ON  CHECK)</act><auto>yes</auto><pas>N1170 M02</pas><sub>basis</sub><temp>N23270 (====PART 9 ====)</temp></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<laser><act1>9363</act1><act2>9997</act2><auto>yes</auto><preset1>9363</preset1><preset2>9997</preset2></laser>
<blocks><act>N23295 G00 X4709.8 Y0 A0</act><auto>yes</auto><pas>N23300 G10</pas><sub>basis</sub></blocks>
<axes><auto>yes</auto><ax1>+05658.114</ax1><ax4>+04319.514</ax4><ax5>+04319.514</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9998</act2><auto>yes</auto><preset1>9994</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05658.933</ax1><ax4>+04320.000</ax4><ax5>+04320.000</ax5><sub>pos</sub></axes>
<blocks><act>N23305 G00 X4709.8 Y0 A1.837</act><auto>yes</auto><pas>N23310 G01 X4709.8 Y0 Z24 F10000</pas><sub>basis</sub><temp>N23315 Q990051</temp></blocks>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax4>+04320.443</ax4><ax5>+04320.443</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04321.806</ax4><ax5>+04321.806</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<blocks><act>N23310 G01 X4709.8 Y0 Z24 F10000</act><auto>yes</auto><pas>N23315 Q990051</pas><sub>basis</sub><temp>N1010 G10 (LCUT-1)</temp></blocks>
<axes><auto>yes</auto><ax3>+00028.208</ax3><ax4>+04321.837</ax4><ax5>+04321.837</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00024.884</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>10</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N1210 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1070 Q= P005</act><auto>yes</auto><pas>N1010 G10 (PIE1-NORM)</pas><sub>basis</sub></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N1040 M02</pas><sub>basis</sub><temp>N1080 Q= P006</temp></blocks>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<ncda><auto>yes</auto><status>4</status></ncda>
<alarm><auto>yes</auto><no>1046</no><prio>4</prio><st>plc</st><v1> </v1></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<alarm><auto>yes</auto><no>1046</no><prio>4</prio><st>plc</st><v1> </v1></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<alarm><auto>yes</auto><no>0</no><prio>255</prio><st>plc</st></alarm>
<ncda><auto>yes</auto><status>6</status></ncda>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub></blocks>
<laser><act1>9395</act1><act2>9998</act2><auto>yes</auto><preset1>9395</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05659.758</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><sub>vel</sub></axes>
<blocks><act>N23325 G01 X4712.8 Y0 Z24 A1.837 F3000</act><auto>yes</auto><pas>N23330 G01 X4712.8 Y0 Z24 A3.673 F9047</pas><sub>basis</sub><temp>N23335 G01 X4712.8 Y0 Z24 A5.51</temp></blocks>
<axes><auto>yes</auto><ax1>+05660.878</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05661.907</ax1><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><sub>vel</sub></axes>
<laser><act1>7673</act1><act2>9995</act2><auto>yes</auto><preset1>7673</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05661.933</ax1><ax4>+04323.351</ax4><ax5>+04323.351</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04327.150</ax4><ax5>+04327.150</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04330.106</ax4><ax5>+04330.106</ax5><sub>pos</sub></axes>
<blocks><act>N23355 G01 X4712.8 Y0 Z24 A12.857</act><auto>yes</auto><pas>N23360 G01 X4712.8 Y0 Z24 A14.694</pas><sub>basis</sub><temp>N23365 G01 X4712.8 Y0 Z24 A16.915</temp></blocks>
<axes><auto>yes</auto><ax4>+04333.483</ax4><ax5>+04333.483</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04336.557</ax4><ax5>+04336.557</ax5><sub>pos</sub></axes>
<laser><act1>9009</act1><auto>yes</auto><preset1>9009</preset1></laser>
<axes><auto>yes</auto><ax4>+04337.604</ax4><ax5>+04337.604</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax4>+04340.526</ax4><ax5>+04340.526</ax5><sub>pos</sub></axes>
<blocks><act>N23385 G01 X4712.8 Y0 Z24 A22.041</act><auto>yes</auto><pas>N23390 G01 X4712.8 Y0 Z24 A23.878</pas><sub>basis</sub><temp>N23395 G01 X4712.8 Y0 Z24 A25.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04343.904</ax4><ax5>+04343.904</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04347.282</ax4><ax5>+04347.282</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04350.659</ax4><ax5>+04350.659</ax5><sub>pos</sub></axes>
<blocks><act>N23415 G01 X4712.8 Y0 Z24 A33.061</act><auto>yes</auto><pas>N23420 G01 X4712.8 Y0 Z24 A34.898</pas><sub>basis</sub><temp>N23425 G01 X4712.8 Y0 Z24 A36.735</temp></blocks>
<axes><auto>yes</auto><ax4>+04353.614</ax4><ax5>+04353.614</ax5><sub>pos</sub></axes>
<blocks><act>N23420 G01 X4712.8 Y0 Z24 A34.898</act><auto>yes</auto><pas>N23425 G01 X4712.8 Y0 Z24 A36.735</pas><sub>basis</sub><temp>N23430 G01 X4712.8 Y0 Z24 A38.571</temp></blocks>
<axes><auto>yes</auto><ax4>+04357.414</ax4><ax5>+04357.414</ax5><sub>pos</sub></axes>
<blocks><act>N23435 G01 X4712.8 Y0 Z24 A40.408</act><auto>yes</auto><pas>N23440 G01 X4712.8 Y0 Z24 A42.245</pas><sub>basis</sub><temp>N23445 G01 X4712.8 Y0 Z24 A44.082</temp></blocks>
<axes><auto>yes</auto><ax4>+04360.370</ax4><ax5>+04360.370</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04363.747</ax4><ax5>+04363.747</ax5><sub>pos</sub></axes>
<blocks><act>N23450 G01 X4712.8 Y0 Z24 A45.918</act><auto>yes</auto><pas>N23455 G01 X4712.8 Y0 Z24 A47.755</pas><sub>basis</sub><temp>N23460 G01 X4712.8 Y0 Z24 A49.592</temp></blocks>
<axes><auto>yes</auto><ax4>+04367.125</ax4><ax5>+04367.125</ax5><sub>pos</sub></axes>
<blocks><act>N23460 G01 X4712.8 Y0 Z24 A49.592</act><auto>yes</auto><pas>N23465 G01 X4712.8 Y0 Z24 A51.429</pas><sub>basis</sub><temp>N23470 G01 X4712.8 Y0 Z24 A53.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04370.502</ax4><ax5>+04370.502</ax5><sub>pos</sub></axes>
<blocks><act>N23470 G01 X4712.8 Y0 Z24 A53.265</act><auto>yes</auto><pas>N23475 G01 X4712.8 Y0 Z24 A55.102</pas><sub>basis</sub><temp>N23480 G01 X4712.8 Y0 Z24 A56.939</temp></blocks>
<axes><auto>yes</auto><ax4>+04373.458</ax4><ax5>+04373.458</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04376.835</ax4><ax5>+04376.835</ax5><sub>pos</sub></axes>
<blocks><act>N23485 G01 X4712.8 Y0 Z24 A58.775</act><auto>yes</auto><pas>N23490 G01 X4712.8 Y0 Z24 A60.612</pas><sub>basis</sub><temp>N23495 G01 X4712.8 Y0 Z24 A62.449</temp></blocks>
<axes><auto>yes</auto><ax4>+04380.213</ax4><ax5>+04380.213</ax5><sub>pos</sub></axes>
<blocks><act>N23495 G01 X4712.8 Y0 Z24 A62.449</act><auto>yes</auto><pas>N23500 G01 X4712.8 Y0 Z24 A64.286</pas><sub>basis</sub><temp>N23505 G01 X4712.8 Y0 Z24 A66.122</temp></blocks>
<axes><auto>yes</auto><ax4>+04383.590</ax4><ax5>+04383.590</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04386.546</ax4><ax5>+04386.546</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04390.345</ax4><ax5>+04390.345</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04393.301</ax4><ax5>+04393.301</ax5><sub>pos</sub></axes>
<blocks><act>N23530 G01 X4712.8 Y0 Z24 A75.306</act><auto>yes</auto><pas>N23535 G01 X4712.8 Y0 Z24 A77.143</pas><sub>basis</sub><temp>N23540 G01 X4712.8 Y0 Z24 A78.98</temp></blocks>
<axes><auto>yes</auto><ax4>+04396.678</ax4><ax5>+04396.678</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04400.056</ax4><ax5>+04400.056</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04403.433</ax4><ax5>+04403.433</ax5><sub>pos</sub></axes>
<blocks><act>N23560 G01 X4712.8 Y0 Z24 A86.327</act><auto>yes</auto><pas>N23565 G01 X4712.8 Y0 Z24 A88.163</pas><sub>basis</sub><temp>N23570 G01 X4712.8 Y0 Z24 A90</temp></blocks>
<axes><auto>yes</auto><ax4>+04406.389</ax4><ax5>+04406.389</ax5><sub>pos</sub></axes>
<blocks><act>N23565 G01 X4712.8 Y0 Z24 A88.163</act><auto>yes</auto><pas>N23570 G01 X4712.8 Y0 Z24 A90</pas><sub>basis</sub><temp>N23575 G01 X4712.8 Y0 Z24 A91.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04410.188</ax4><ax5>+04410.188</ax5><sub>pos</sub></axes>
<blocks><act>N23575 G01 X4712.8 Y0 Z24 A91.837</act><auto>yes</auto><pas>N23580 G01 X4712.8 Y0 Z24 A93.673</pas><sub>basis</sub><temp>N23585 G01 X4712.8 Y0 Z24 A95.51</temp></blocks>
<axes><auto>yes</auto><ax4>+04413.144</ax4><ax5>+04413.144</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04416.521</ax4><ax5>+04416.521</ax5><sub>pos</sub></axes>
<blocks><act>N23595 G01 X4712.8 Y0 Z24 A99.184</act><auto>yes</auto><pas>N23600 G01 X4712.8 Y0 Z24 A101.02</pas><sub>basis</sub><temp>N23605 G01 X4712.8 Y0 Z24 A102.857</temp></blocks>
<axes><auto>yes</auto><ax4>+04419.899</ax4><ax5>+04419.899</ax5><sub>pos</sub></axes>
<blocks><act>N23605 G01 X4712.8 Y0 Z24 A102.857</act><auto>yes</auto><pas>N23610 G01 X4712.8 Y0 Z24 A104.694</pas><sub>basis</sub><temp>N23615 G01 X4712.8 Y0 Z24 A106.531</temp></blocks>
<axes><auto>yes</auto><ax4>+04423.276</ax4><ax5>+04423.276</ax5><sub>pos</sub></axes>
<blocks><act>N23610 G01 X4712.8 Y0 Z24 A104.694</act><auto>yes</auto><pas>N23615 G01 X4712.8 Y0 Z24 A106.531</pas><sub>basis</sub><temp>N23620 G01 X4712.8 Y0 Z24 A108.367</temp></blocks>
<axes><auto>yes</auto><ax4>+04426.232</ax4><ax5>+04426.232</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04429.609</ax4><ax5>+04429.609</ax5><sub>pos</sub></axes>
<blocks><act>N23630 G01 X4712.8 Y0 Z24 A112.041</act><auto>yes</auto><pas>N23635 G01 X4712.8 Y0 Z24 A113.878</pas><sub>basis</sub><temp>N23640 G01 X4712.8 Y0 Z24 A115.714</temp></blocks>
<axes><auto>yes</auto><ax4>+04432.987</ax4><ax5>+04432.987</ax5><sub>pos</sub></axes>
<blocks><act>N23640 G01 X4712.8 Y0 Z24 A115.714</act><auto>yes</auto><pas>N23645 G01 X4712.8 Y0 Z24 A117.551</pas><sub>basis</sub><temp>N23650 G01 X4712.8 Y0 Z24 A119.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04436.364</ax4><ax5>+04436.364</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04439.742</ax4><ax5>+04439.742</ax5><sub>pos</sub></axes>
<blocks><act>N23655 G01 X4712.8 Y0 Z24 A121.225</act><auto>yes</auto><pas>N23660 G01 X4712.8 Y0 Z24 A123.061</pas><sub>basis</sub><temp>N23665 G01 X4712.8 Y0 Z24 A124.898</temp></blocks>
<axes><auto>yes</auto><ax4>+04443.119</ax4><ax5>+04443.119</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04446.075</ax4><ax5>+04446.075</ax5><sub>pos</sub></axes>
<blocks><act>N23675 G01 X4712.8 Y0 Z24 A128.571</act><auto>yes</auto><pas>N23680 G01 X4712.8 Y0 Z24 A130.408</pas><sub>basis</sub><temp>N23685 G01 X4712.8 Y0 Z24 A132.245</temp></blocks>
<axes><auto>yes</auto><ax4>+04449.452</ax4><ax5>+04449.452</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04452.830</ax4><ax5>+04452.830</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04456.207</ax4><ax5>+04456.207</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04459.163</ax4><ax5>+04459.163</ax5><sub>pos</sub></axes>
<blocks><act>N23710 G01 X4712.8 Y0 Z24 A141.429</act><auto>yes</auto><pas>N23715 G01 X4712.8 Y0 Z24 A143.265</pas><sub>basis</sub><temp>N23720 G01 X4712.8 Y0 Z24 A145.102</temp></blocks>
<axes><auto>yes</auto><ax4>+04462.540</ax4><ax5>+04462.540</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04465.918</ax4><ax5>+04465.918</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04469.295</ax4><ax5>+04469.295</ax5><sub>pos</sub></axes>
<blocks><act>N23735 G01 X4712.8 Y0 Z24 A150.612</act><auto>yes</auto><pas>N23740 G01 X4712.8 Y0 Z24 A152.449</pas><sub>basis</sub><temp>N23745 G01 X4712.8 Y0 Z24 A154.286</temp></blocks>
<axes><auto>yes</auto><ax4>+04472.673</ax4><ax5>+04472.673</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04476.051</ax4><ax5>+04476.051</ax5><sub>pos</sub></axes>
<blocks><act>N23755 G01 X4712.8 Y0 Z24 A157.959</act><auto>yes</auto><pas>N23760 G01 X4712.8 Y0 Z24 A159.796</pas><sub>basis</sub><temp>N23765 G01 X4712.8 Y0 Z24 A161.633</temp></blocks>
<axes><auto>yes</auto><ax4>+04479.006</ax4><ax5>+04479.006</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04482.383</ax4><ax5>+04482.383</ax5><sub>pos</sub></axes>
<blocks><act>N23775 G01 X4712.8 Y0 Z24 A165.306</act><auto>yes</auto><pas>N23780 G01 X4712.8 Y0 Z24 A167.143</pas><sub>basis</sub><temp>N23785 G01 X4712.8 Y0 Z24 A168.98</temp></blocks>
<axes><auto>yes</auto><ax4>+04485.761</ax4><ax5>+04485.761</ax5><sub>pos</sub></axes>
<blocks><act>N23780 G01 X4712.8 Y0 Z24 A167.143</act><auto>yes</auto><pas>N23785 G01 X4712.8 Y0 Z24 A168.98</pas><sub>basis</sub><temp>N23790 G01 X4712.8 Y0 Z24 A170.816</temp></blocks>
<axes><auto>yes</auto><ax4>+04489.139</ax4><ax5>+04489.139</ax5><sub>pos</sub></axes>
<blocks><act>N23790 G01 X4712.8 Y0 Z24 A170.816</act><auto>yes</auto><pas>N23795 G01 X4712.8 Y0 Z24 A172.653</pas><sub>basis</sub><temp>N23800 G01 X4712.8 Y0 Z24 A174.49</temp></blocks>
<axes><auto>yes</auto><ax4>+04492.094</ax4><ax5>+04492.094</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04495.471</ax4><ax5>+04495.471</ax5><sub>pos</sub></axes>
<blocks><act>N23810 G01 X4712.8 Y0 Z24 A178.163</act><auto>yes</auto><pas>N23815 G01 X4712.8 Y0 Z24 A180</pas><sub>basis</sub><temp>N23820 G01 X4712.8 Y0 Z24 A181.837</temp></blocks>
<axes><auto>yes</auto><ax4>+04498.849</ax4><ax5>+04498.849</ax5><sub>pos</sub></axes>
<blocks><act>N23820 G01 X4712.8 Y0 Z24 A181.837</act><auto>yes</auto><pas>N23825 G01 X4712.8 Y0 Z24 A183.673</pas><sub>basis</sub><temp>N23830 G01 X4712.8 Y0 Z24 A185.51</temp></blocks>
<axes><auto>yes</auto><ax4>+04502.227</ax4><ax5>+04502.227</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04505.604</ax4><ax5>+04505.604</ax5><sub>pos</sub></axes>
<blocks><act>N23835 G01 X4712.8 Y0 Z24 A187.347</act><auto>yes</auto><pas>N23840 G01 X4712.8 Y0 Z24 A189.184</pas><sub>basis</sub><temp>N23845 G01 X4712.8 Y0 Z24 A191.02</temp></blocks>
<axes><auto>yes</auto><ax4>+04508.982</ax4><ax5>+04508.982</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04511.937</ax4><ax5>+04511.937</ax5><sub>pos</sub></axes>
<blocks><act>N23855 G01 X4712.8 Y0 Z24 A194.694</act><auto>yes</auto><pas>N23860 G01 X4712.8 Y0 Z24 A196.531</pas><sub>basis</sub><temp>N23865 G01 X4712.8 Y0 Z24 A198.367</temp></blocks>
<axes><auto>yes</auto><ax4>+04515.315</ax4><ax5>+04515.315</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04518.692</ax4><ax5>+04518.692</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04522.070</ax4><ax5>+04522.070</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04525.025</ax4><ax5>+04525.025</ax5><sub>pos</sub></axes>
<blocks><act>N23890 G01 X4712.8 Y0 Z24 A207.551</act><auto>yes</auto><pas>N23895 G01 X4712.8 Y0 Z24 A209.388</pas><sub>basis</sub><temp>N23900 G01 X4712.8 Y0 Z24 A211.225</temp></blocks>
<axes><auto>yes</auto><ax4>+04528.403</ax4><ax5>+04528.403</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04531.780</ax4><ax5>+04531.780</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04535.158</ax4><ax5>+04535.158</ax5><sub>pos</sub></axes>
<blocks><act>N23915 G01 X4712.8 Y0 Z24 A216.735</act><auto>yes</auto><pas>N23920 G01 X4712.8 Y0 Z24 A218.571</pas><sub>basis</sub><temp>N23925 G01 X4712.8 Y0 Z24 A220.408</temp></blocks>
<axes><auto>yes</auto><ax4>+04538.535</ax4><ax5>+04538.535</ax5><sub>pos</sub></axes>
<blocks><act>N23925 G01 X4712.8 Y0 Z24 A220.408</act><auto>yes</auto><pas>N23930 G01 X4712.8 Y0 Z24 A222.245</pas><sub>basis</sub><temp>N23935 G01 X4712.8 Y0 Z24 A224.082</temp></blocks>
<axes><auto>yes</auto><ax4>+04541.913</ax4><ax5>+04541.913</ax5><sub>pos</sub></axes>
<blocks><act>N23935 G01 X4712.8 Y0 Z24 A224.082</act><auto>yes</auto><pas>N23940 G01 X4712.8 Y0 Z24 A225.918</pas><sub>basis</sub><temp>N23945 G01 X4712.8 Y0 Z24 A227.755</temp></blocks>
<axes><auto>yes</auto><ax4>+04544.868</ax4><ax5>+04544.868</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04548.246</ax4><ax5>+04548.246</ax5><sub>pos</sub></axes>
<blocks><act>N23950 G01 X4712.8 Y0 Z24 A229.592</act><auto>yes</auto><pas>N23955 G01 X4712.8 Y0 Z24 A231.429</pas><sub>basis</sub><temp>N23960 G01 X4712.8 Y0 Z24 A233.265</temp></blocks>
<axes><auto>yes</auto><ax4>+04551.623</ax4><ax5>+04551.623</ax5><sub>pos</sub></axes>
<blocks><act>N23960 G01 X4712.8 Y0 Z24 A233.265</act><auto>yes</auto><pas>N23965 G01 X4712.8 Y0 Z24 A235.102</pas><sub>basis</sub><temp>N23970 G01 X4712.8 Y0 Z24 A236.939</temp></blocks>
<axes><auto>yes</auto><ax4>+04555.001</ax4><ax5>+04555.001</ax5><sub>pos</sub></axes>
<blocks><act>N23970 G01 X4712.8 Y0 Z24 A236.939</act><auto>yes</auto><pas>N23975 G01 X4712.8 Y0 Z24 A238.775</pas><sub>basis</sub><temp>N23980 G01 X4712.8 Y0 Z24 A240.612</temp></blocks>
<axes><auto>yes</auto><ax4>+04557.956</ax4><ax5>+04557.956</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04561.334</ax4><ax5>+04561.334</ax5><sub>pos</sub></axes>
<blocks><act>N23990 G01 X4712.8 Y0 Z24 A244.286</act><auto>yes</auto><pas>N23995 G01 X4712.8 Y0 Z24 A246.122</pas><sub>basis</sub><temp>N24000 G01 X4712.8 Y0 Z24 A247.959</temp></blocks>
<axes><auto>yes</auto><ax4>+04564.711</ax4><ax5>+04564.711</ax5><sub>pos</sub></axes>
<blocks><act>N23995 G01 X4712.8 Y0 Z24 A246.122</act><auto>yes</auto><pas>N24000 G01 X4712.8 Y0 Z24 A247.959</pas><sub>basis</sub><temp>N24005 G01 X4712.8 Y0 Z24 A249.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04568.089</ax4><ax5>+04568.089</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04571.466</ax4><ax5>+04571.466</ax5><sub>pos</sub></axes>
<blocks><act>N24015 G01 X4712.8 Y0 Z24 A253.469</act><auto>yes</auto><pas>N24020 G01 X4712.8 Y0 Z24 A255.306</pas><sub>basis</sub><temp>N24025 G01 X4712.8 Y0 Z24 A257.143</temp></blocks>
<axes><auto>yes</auto><ax4>+04574.844</ax4><ax5>+04574.844</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04577.799</ax4><ax5>+04577.799</ax5><sub>pos</sub></axes>
<blocks><act>N24035 G01 X4712.8 Y0 Z24 A260.816</act><auto>yes</auto><pas>N24040 G01 X4712.8 Y0 Z24 A262.653</pas><sub>basis</sub><temp>N24045 G01 X4712.8 Y0 Z24 A264.49</temp></blocks>
<axes><auto>yes</auto><ax4>+04581.177</ax4><ax5>+04581.177</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04584.554</ax4><ax5>+04584.554</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04587.932</ax4><ax5>+04587.932</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04590.887</ax4><ax5>+04590.887</ax5><sub>pos</sub></axes>
<blocks><act>N24070 G01 X4712.8 Y0 Z24 A273.673</act><auto>yes</auto><pas>N24075 G01 X4712.8 Y0 Z24 A275.51</pas><sub>basis</sub><temp>N24080 G01 X4712.8 Y0 Z24 A277.347</temp></blocks>
<axes><auto>yes</auto><ax4>+04594.265</ax4><ax5>+04594.265</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04597.642</ax4><ax5>+04597.642</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04601.020</ax4><ax5>+04601.020</ax5><sub>pos</sub></axes>
<blocks><act>N24095 G01 X4712.8 Y0 Z24 A282.857</act><auto>yes</auto><pas>N24100 G01 X4712.8 Y0 Z24 A284.694</pas><sub>basis</sub><temp>N24105 G01 X4712.8 Y0 Z24 A286.531</temp></blocks>
<axes><auto>yes</auto><ax4>+04604.397</ax4><ax5>+04604.397</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04607.775</ax4><ax5>+04607.775</ax5><sub>pos</sub></axes>
<blocks><act>N24115 G01 X4712.8 Y0 Z24 A290.204</act><auto>yes</auto><pas>N24120 G01 X4712.8 Y0 Z24 A292.041</pas><sub>basis</sub><temp>N24125 G01 X4712.8 Y0 Z24 A293.878</temp></blocks>
<axes><auto>yes</auto><ax4>+04610.730</ax4><ax5>+04610.730</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04614.108</ax4><ax5>+04614.108</ax5><sub>pos</sub></axes>
<blocks><act>N24130 G01 X4712.8 Y0 Z24 A295.714</act><auto>yes</auto><pas>N24135 G01 X4712.8 Y0 Z24 A297.551</pas><sub>basis</sub><temp>N24140 G01 X4712.8 Y0 Z24 A299.388</temp></blocks>
<axes><auto>yes</auto><ax4>+04617.485</ax4><ax5>+04617.485</ax5><sub>pos</sub></axes>
<blocks><act>N24140 G01 X4712.8 Y0 Z24 A299.388</act><auto>yes</auto><pas>N24145 G01 X4712.8 Y0 Z24 A301.225</pas><sub>basis</sub><temp>N24150 G01 X4712.8 Y0 Z24 A303.061</temp></blocks>
<axes><auto>yes</auto><ax4>+04620.863</ax4><ax5>+04620.863</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04624.240</ax4><ax5>+04624.240</ax5><sub>pos</sub></axes>
<blocks><act>N24160 G01 X4712.8 Y0 Z24 A306.735</act><auto>yes</auto><pas>N24165 G01 X4712.8 Y0 Z24 A308.571</pas><sub>basis</sub><temp>N24170 G01 X4712.8 Y0 Z24 A310.408</temp></blocks>
<axes><auto>yes</auto><ax4>+04627.618</ax4><ax5>+04627.618</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04630.573</ax4><ax5>+04630.573</ax5><sub>pos</sub></axes>
<blocks><act>N24175 G01 X4712.8 Y0 Z24 A312.245</act><auto>yes</auto><pas>N24180 G01 X4712.8 Y0 Z24 A314.082</pas><sub>basis</sub><temp>N24185 G01 X4712.8 Y0 Z24 A315.918</temp></blocks>
<axes><auto>yes</auto><ax4>+04633.951</ax4><ax5>+04633.951</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04637.328</ax4><ax5>+04637.328</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04640.706</ax4><ax5>+04640.706</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04643.661</ax4><ax5>+04643.661</ax5><sub>pos</sub></axes>
<blocks><act>N24210 G01 X4712.8 Y0 Z24 A325.102</act><auto>yes</auto><pas>N24215 G01 X4712.8 Y0 Z24 A326.939</pas><sub>basis</sub><temp>N24220 G01 X4712.8 Y0 Z24 A328.775</temp></blocks>
<axes><auto>yes</auto><ax4>+04647.039</ax4><ax5>+04647.039</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04650.416</ax4><ax5>+04650.416</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04653.794</ax4><ax5>+04653.794</ax5><sub>pos</sub></axes>
<blocks><act>N24240 G01 X4712.8 Y0 Z24 A336.122</act><auto>yes</auto><pas>N24245 G01 X4712.8 Y0 Z24 A337.959</pas><sub>basis</sub><temp>N24250 G01 X4712.8 Y0 Z24 A339.796</temp></blocks>
<axes><auto>yes</auto><ax4>+04657.171</ax4><ax5>+04657.171</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04660.549</ax4><ax5>+04660.549</ax5><sub>pos</sub></axes>
<blocks><act>N24260 G01 X4712.8 Y0 Z24 A343.469</act><auto>yes</auto><pas>N24265 G01 X4712.8 Y0 Z24 A345.306</pas><sub>basis</sub><temp>N24270 G01 X4712.8 Y0 Z24 A346.757</temp></blocks>
<axes><auto>yes</auto><ax4>+04663.504</ax4><ax5>+04663.504</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04666.463</ax4><ax5>+04666.463</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<blocks><act>N24275 G01 X4712.8 Y0 Z24 A346.758</act><auto>yes</auto><pas>N24280 G01 X4712.8 Y0 Z24 A348.98</pas><sub>basis</sub><temp>N24285 G01 X4712.8 Y0 Z24 A350.816</temp></blocks>
<laser><act1>8639</act1><auto>yes</auto><preset1>8639</preset1></laser>
<axes><auto>yes</auto><ax4>+04667.559</ax4><ax5>+04667.559</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N24280 G01 X4712.8 Y0 Z24 A348.98</act><auto>yes</auto><pas>N24285 G01 X4712.8 Y0 Z24 A350.816</pas><sub>basis</sub><temp>N24290 G01 X4712.8 Y0 Z24 A352.653</temp></blocks>
<axes><auto>yes</auto><ax4>+04670.925</ax4><ax5>+04670.925</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04673.880</ax4><ax5>+04673.880</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04677.680</ax4><ax5>+04677.680</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04680.610</ax4><ax5>+04680.610</ax5><sub>pos</sub></axes>
<blocks><act>N24315 G01 X4712.8 Y0 Z24 A361.837</act><auto>yes</auto><pas>N24325 Q= P11 (LASER OFF)</pas><sub>basis</sub><temp>N1010 G10 (END-LI-PO_W)</temp></blocks>
<axes><auto>yes</auto><ax4>+04681.837</ax4><ax5>+04681.837</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<axes><auto>yes</auto><ax3>+00021.853</ax3><sub>pos</sub></axes>
<blocks><act>N1090 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax3>+00023.197</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00027.563</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>13</ax3><sub>vel</sub></axes>
<blocks><act>N24330 G00 Z29</act><auto>yes</auto><pas>N24335 (====CONTOUR 2 ====)</pas><sub>basis</sub><temp>N24340 G00 X5005.491 Y0 A360</temp></blocks>
<axes><auto>yes</auto><ax3>+00029.000</ax3><sub>pos</sub></axes>
<laser><act1>8993</act1><act2>9998</act2><auto>yes</auto><preset1>8993</preset1><preset2>9998</preset2></laser>
<blocks><act>N24340 G00 X5005.491 Y0 A360</act><auto>yes</auto><pas>N24345 G10</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<axes><auto>yes</auto><ax1>+05664.433</ax1><ax4>+04681.817</ax4><ax5>+04681.817</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>9</ax1><ax3>0</ax3><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><act2>9999</act2><auto>yes</auto><preset1>9994</preset1><preset2>9999</preset2></laser>
<axes><auto>yes</auto><ax1>+05670.747</ax1><ax4>+04681.774</ax4><ax5>+04681.774</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05682.669</ax1><ax4>+04681.695</ax4><ax5>+04681.695</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>26</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05699.610</ax1><ax4>+04681.585</ax4><ax5>+04681.585</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05721.571</ax1><ax4>+04681.443</ax4><ax5>+04681.443</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>45</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05744.904</ax1><ax4>+04681.293</ax4><ax5>+04681.293</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05776.276</ax1><ax4>+04681.093</ax4><ax5>+04681.093</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>63</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05812.423</ax1><ax4>+04680.863</ax4><ax5>+04680.863</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05847.306</ax1><ax4>+04680.648</ax4><ax5>+04680.648</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>61</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05873.808</ax1><ax4>+04680.485</ax4><ax5>+04680.485</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05902.369</ax1><ax4>+04680.310</ax4><ax5>+04680.310</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>42</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05920.295</ax1><ax4>+04680.201</ax4><ax5>+04680.201</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05936.188</ax1><ax4>+04680.105</ax4><ax5>+04680.105</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>25</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05947.180</ax1><ax4>+04680.040</ax4><ax5>+04680.040</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05953.271</ax1><ax4>+04680.006</ax4><ax5>+04680.006</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>7</ax1><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05954.624</ax1><ax4>+04680.000</ax4><ax5>+04680.000</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<blocks><act>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    ))</act><auto>yes</auto><pas>N24350 G00 X5005.491 Y0 A270.506</pas><sub>basis</sub><temp>N24355 G01 X5005.491 Y0 Z24 F10000</temp></blocks>
<axes><auto>yes</auto><ax4>+04679.335</ax4><ax5>+04679.335</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>0</ax1><ax4>8</ax4><ax5>8</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N24350 G00 X5005.491 Y0 A270.506</act><auto>yes</auto><pas>N24355 G01 X5005.491 Y0 Z24 F10000</pas><sub>basis</sub><temp>N24360 Q= P10 (LASER ON)</temp></blocks>
<axes><auto>yes</auto><ax4>+04676.544</ax4><ax5>+04676.544</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04670.695</ax4><ax5>+04670.695</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>31</ax4><ax5>31</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04662.011</ax4><ax5>+04662.011</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04650.491</ax4><ax5>+04650.491</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>55</ax4><ax5>55</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04638.085</ax4><ax5>+04638.085</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04623.377</ax4><ax5>+04623.377</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>58</ax4><ax5>58</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04611.352</ax4><ax5>+04611.352</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04602.054</ax4><ax5>+04602.054</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>34</ax4><ax5>34</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04596.154</ax4><ax5>+04596.154</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>+04591.968</ax4><ax5>+04591.968</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>12</ax4><ax5>12</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax4>+04590.508</ax4><ax5>+04590.508</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><auto>yes</auto><preset1>6996</preset1></laser>
<blocks><act>N24355 G01 X5005.491 Y0 Z24 F10000</act><auto>yes</auto><pas>N24360 Q= P10 (LASER ON)</pas><sub>basis</sub><temp>N1010 G10 (STA-PO-PI-LI-P6018DA_A1)</temp></blocks>
<axes><auto>yes</auto><ax3>+00027.856</ax3><ax4>+04590.506</ax4><ax5>+04590.506</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>11</ax3><ax4>0</ax4><ax5>0</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax3>+00024.884</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>+00024.000</ax3><sub>pos</sub></axes>
<axes><auto>yes</auto><ax3>0</ax3><sub>vel</sub></axes>
<blocks><act>N2250 G10</act><auto>yes</auto><pas>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</pas><sub>basis</sub><temp>(AUTOMATICALLY GENERATED INTERMEDIATE BLOCK    )</temp></blocks>
<blocks><act>N1020 U3 M117</act><auto>yes</auto><pas>N1030 G10</pas><sub>basis</sub></blocks>
<laser><act1>8995</act1><act2>0</act2><auto>yes</auto><preset1>8995</preset1><preset2>0</preset2></laser>
<laser><act1>0</act1><auto>yes</auto><preset1>0</preset1></laser>
<blocks><act>N1020 U0 M118</act><auto>yes</auto><pas>N1022 G10</pas><sub>basis</sub></blocks>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05955.308</ax1><ax4>+04590.609</ax4><ax5>+04590.609</ax5><sub>pos</sub></axes>
<laser><act1>9412</act1><act2>9998</act2><auto>yes</auto><preset1>9412</preset1><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05956.428</ax1><ax4>+04590.749</ax4><ax5>+04590.749</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<blocks><act>N24365 G01 X5008.488 Y0 Z24 A270.881 F3021</act><auto>yes</auto><pas>N24370 G01 X5008.476 Y0 Z24 A271.762 F9040</pas><sub>basis</sub><temp>N24375 G01 X5008.402 Y0 Z24 A273.521 F8984</temp></blocks>
<axes><auto>yes</auto><ax1>+05957.524</ax1><ax4>+04590.878</ax4><ax5>+04590.878</ax5><sub>pos</sub></axes>
<laser><act1>6996</act1><act2>9995</act2><auto>yes</auto><preset1>6996</preset1><preset2>9995</preset2></laser>
<blocks><act>N24370 G01 X5008.476 Y0 Z24 A271.762 F9040</act><auto>yes</auto><pas>N24375 G01 X5008.402 Y0 Z24 A273.521 F8984</pas><sub>basis</sub><temp>N24380 G01 X5008.282 Y0 Z24 A275.232 F8873</temp></blocks>
<axes><auto>yes</auto><ax1>+05957.601</ax1><ax4>+04592.390</ax4><ax5>+04592.390</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<blocks><act>N24380 G01 X5008.282 Y0 Z24 A275.232 F8873</act><auto>yes</auto><pas>N24385 G01 X5008.117 Y0 Z24 A276.889 F8706</pas><sub>basis</sub><temp>N24390 G01 X5007.906 Y0 Z24 A278.492 F8483</temp></blocks>
<axes><auto>yes</auto><ax1>+05957.440</ax1><ax4>+04595.291</ax4><ax5>+04595.291</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05957.094</ax1><ax4>+04598.467</ax4><ax5>+04598.467</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<blocks><act>N24395 G01 X5007.654 Y0 Z24 A280.018 F8204</act><auto>yes</auto><pas>N24400 G01 X5007.364 Y0 Z24 A281.441 F7866</pas><sub>basis</sub><temp>N24405 G01 X5007.037 Y0 Z24 A282.762 F7468</temp></blocks>
<axes><auto>yes</auto><ax1>+05956.577</ax1><ax4>+04601.402</ax4><ax5>+04601.402</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<blocks><act>N24405 G01 X5007.037 Y0 Z24 A282.762 F7468</act><auto>yes</auto><pas>N24410 G01 X5006.674 Y0 Z24 A283.972 F7008</pas><sub>basis</sub><temp>N24415 G01 X5006.536 Y0 Z24 A284.352 F6485</temp></blocks>
<axes><auto>yes</auto><ax1>+05955.901</ax1><ax4>+04603.892</ax4><ax5>+04603.892</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<laser><act1>9444</act1><act2>9996</act2><auto>yes</auto><preset1>9444</preset1><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05955.657</ax1><ax4>+04604.518</ax4><ax5>+04604.518</ax5><sub>pos</sub></axes>
<blocks><act>N24430 G01 X5005.864 Y0 Z24 A285.984 F5900</act><auto>yes</auto><pas>N24435 G01 X5005.42 Y0 Z24 A286.771 F5263</pas><sub>basis</sub><temp>N24440 G01 X5004.957 Y0 Z24 A287.394 F4594</temp></blocks>
<laser><act1>9994</act1><act2>9997</act2><auto>yes</auto><preset1>9994</preset1><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05954.868</ax1><ax4>+04606.427</ax4><ax5>+04606.427</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05954.011</ax1><ax4>+04607.596</ax4><ax5>+04607.596</ax5><sub>pos</sub></axes>
<laser><act1>9958</act1><act2>9998</act2><auto>yes</auto><preset1>9958</preset1><preset2>9998</preset2></laser>
<blocks><act>N24450 G01 X5003.992 Y0 Z24 A288.118 F3381</act><auto>yes</auto><pas>N24455 G01 X5003.498 Y0 Z24 A288.21 F3046</pas><sub>basis</sub><temp>N24460 G01 X5003.007 Y0 Z24 A288.118</temp></blocks>
<axes><auto>yes</auto><ax1>+05952.937</ax1><ax4>+04608.179</ax4><ax5>+04608.179</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9433</act1><auto>yes</auto><preset1>9433</preset1></laser>
<axes><auto>yes</auto><ax1>+05951.836</ax1><ax4>+04607.870</ax4><ax5>+04607.870</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05950.787</ax1><ax4>+04606.686</ax4><ax5>+04606.686</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>6</ax4><ax5>6</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05949.853</ax1><ax4>+04604.777</ax4><ax5>+04604.777</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<axes><auto>yes</auto><ax1>+05949.163</ax1><ax4>+04602.667</ax4><ax5>+04602.667</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05948.523</ax1><ax4>+04599.867</ax4><ax5>+04599.867</ax5><sub>pos</sub></axes>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05948.053</ax1><ax4>+04596.781</ax4><ax5>+04596.781</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>
<blocks><act>N24520 G01 X4998.718 Y0 Z24 A275.229 F8706</act><auto>yes</auto><pas>N24525 G01 X4998.597 Y0 Z24 A273.512 F8872</pas><sub>basis</sub><temp>N24530 G01 X4998.524 Y0 Z24 A271.771 F8984</temp></blocks>
<axes><auto>yes</auto><ax1>+05947.759</ax1><ax4>+04593.511</ax4><ax5>+04593.511</ax5><sub>pos</sub></axes>
<laser><act2>9995</act2><auto>yes</auto><preset2>9995</preset2></laser>
<axes><auto>yes</auto><ax1>+05947.641</ax1><ax4>+04590.150</ax4><ax5>+04590.150</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>14</ax4><ax5>14</ax5><sub>vel</sub></axes>
<blocks><act>N24540 G01 X4998.524 Y0 Z24 A268.229</act><auto>yes</auto><pas>N24545 G01 X4998.597 Y0 Z24 A266.485 F8984</pas><sub>basis</sub><temp>N24550 G01 X4998.718 Y0 Z24 A264.766 F8872</temp></blocks>
<axes><auto>yes</auto><ax1>+05947.683</ax1><ax4>+04587.206</ax4><ax5>+04587.206</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05947.896</ax1><ax4>+04583.916</ax4><ax5>+04583.916</ax5><sub>pos</sub></axes>
<blocks><act>N24560 G01 X4999.094 Y0 Z24 A261.503 F8482</act><auto>yes</auto><pas>N24565 G01 X4999.345 Y0 Z24 A259.988 F8203</pas><sub>basis</sub><temp>N24570 G01 X4999.637 Y0 Z24 A258.555 F7865</temp></blocks>
<laser><act2>9996</act2><auto>yes</auto><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05948.282</ax1><ax4>+04580.793</ax4><ax5>+04580.793</ax5><sub>pos</sub></axes>
<blocks><act>N24570 G01 X4999.637 Y0 Z24 A258.555 F7865</act><auto>yes</auto><pas>N24575 G01 X4999.966 Y0 Z24 A257.23 F7468</pas><sub>basis</sub><temp>N24580 G01 X5000.325 Y0 Z24 A256.028 F7008</temp></blocks>
<axes><auto>yes</auto><ax1>+05948.840</ax1><ax4>+04577.935</ax4><ax5>+04577.935</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>11</ax4><ax5>11</ax5><sub>vel</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<blocks><act>N24580 G01 X5000.325 Y0 Z24 A256.028 F7008</act><auto>yes</auto><pas>N24585 G01 X5000.718 Y0 Z24 A254.95 F6484</pas><sub>basis</sub><temp>N24590 G01 X5001.138 Y0 Z24 A254.011 F5899</temp></blocks>
<axes><auto>yes</auto><ax1>+05949.459</ax1><ax4>+04575.742</ax4><ax5>+04575.742</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>+05950.314</ax1><ax4>+04573.721</ax4><ax5>+04573.721</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<blocks><act>N24600 G01 X5002.043 Y0 Z24 A252.607 F4592</act><auto>yes</auto><pas>N24605 G01 X5002.522 Y0 Z24 A252.154 F3940</pas><sub>basis</sub><temp>N24610 G01 X5003.008 Y0 Z24 A251.881 F3381</temp></blocks>
<laser><act2>9998</act2><auto>yes</auto><preset2>9998</preset2></laser>
<axes><auto>yes</auto><ax1>+05951.298</ax1><ax4>+04572.365</ax4><ax5>+04572.365</ax5><sub>pos</sub></axes>
<laser><act1>9812</act1><auto>yes</auto><preset1>9812</preset1></laser>
<blocks><act>N24610 G01 X5003.008 Y0 Z24 A251.881 F3381</act><auto>yes</auto><pas>N24615 G01 X5003.5 Y0 Z24 A251.79 F3045</pas><sub>basis</sub><temp>N24620 G01 X5003.996 Y0 Z24 A251.883 F3046</temp></blocks>
<axes><auto>yes</auto><ax1>+05952.370</ax1><ax4>+04571.813</ax4><ax5>+04571.813</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>2</ax1><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>9433</act1><auto>yes</auto><preset1>9433</preset1></laser>
<axes><auto>yes</auto><ax1>+05953.471</ax1><ax4>+04572.153</ax4><ax5>+04572.153</ax5><sub>pos</sub></axes>
<blocks><act>N24630 G01 X5004.958 Y0 Z24 A252.607 F3941</act><auto>yes</auto><pas>N24635 G01 X5005.42 Y0 Z24 A253.23 F4595</pas><sub>basis</sub><temp>N24640 G01 X5005.864 Y0 Z24 A254.016 F5263</temp></blocks>
<laser><act1>9994</act1><auto>yes</auto><preset1>9994</preset1></laser>
<axes><auto>yes</auto><ax1>+05954.517</ax1><ax4>+04573.376</ax4><ax5>+04573.376</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax1>1</ax1><ax4>7</ax4><ax5>7</ax5><sub>vel</sub></axes>
<axes><auto>yes</auto><ax1>+05955.333</ax1><ax4>+04575.022</ax4><ax5>+04575.022</ax5><sub>pos</sub></axes>
<laser><act2>9997</act2><auto>yes</auto><preset2>9997</preset2></laser>
<blocks><act>N24650 G01 X5006.673 Y0 Z24 A256.025 F6484</act><auto>yes</auto><pas>N24655 G01 X5006.941 Y0 Z24 A256.917 F7007</pas><sub>basis</sub><temp>N24660 G01 X5006.941 Y0 Z24 A256.917 F3000</temp></blocks>
<axes><auto>yes</auto><ax1>+05956.067</ax1><ax4>+04576.917</ax4><ax5>+04576.917</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>1</ax4><ax5>1</ax5><sub>vel</sub></axes>
<laser><act1>7708</act1><act2>9995</act2><auto>yes</auto><preset1>7708</preset1><preset2>9995</preset2></laser>
<blocks><act>N24665 G01 X5007.036 Y0 Z24 A257.234 F7007</act><auto>yes</auto><pas>N24670 G01 X5007.363 Y0 Z24 A258.556 F7467</pas><sub>basis</sub><temp>N24675 G01 X5007.653 Y0 Z24 A259.979 F7865</temp></blocks>
<axes><auto>yes</auto><ax1>+05956.347</ax1><ax4>+04578.291</ax4><ax5>+04578.291</ax5><sub>pos</sub></axes>
<laser><act1>9994</act1><act2>9996</act2><auto>yes</auto><preset1>9994</preset1><preset2>9996</preset2></laser>
<axes><auto>yes</auto><ax1>+05956.927</ax1><ax4>+04581.206</ax4><ax5>+04581.206</ax5><sub>pos</sub></axes>
<axes><auto>yes</auto><ax4>13</ax4><ax5>13</ax5><sub>vel</sub></axes>`
