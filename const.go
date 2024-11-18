package tcpxml

import "embed"

//go:embed *.yaml
var yamlFiles embed.FS

// @EnumConfig(noCamel, noCase)
//go:generate ag --package-mode=true
