package controllers

import (
	"github.com/juliorenner/oai-k8s/operator/controllers/utils"
	v1 "k8s.io/api/core/v1"
)

const (
	CU SplitPiece = "cu"
	DU SplitPiece = "du"
	RU SplitPiece = "ru"

	CUTemplateConfigMapName     = "operator-cu-template"
	DUTemplateConfigMapName     = "operator-du-template"
	RUTemplateConfigMapName     = "operator-ru-template"
	DisaggregationConfigMapName = "operator-disaggregations"

	DisaggregationKey = "disaggregation"

	SplitMemoryLimitValue   = 1024
	SplitMemoryRequestValue = 512
	SplitCPULimitValue      = 1000
	SplitCPURequestValue    = 500

	operatorNamespace          = "operator-system"
	cuConfigMapContentTemplate = "upfaddress: %s\nlocaladdress: %s\nsouthaddress: %s\n"
	duConfigMapContentTemplate = "northaddress: %s\nlocaladdress: %s\nsouthaddress: %s\n"
	ruConfigMapContentTemplate = "northaddress: %s\nlocaladdress: %s\n"
)

type SplitPiece string

type cuContent struct {
	UPF          string
	LocalAddress string
	SouthAddress string
}

type duContent struct {
	LocalAddress string
	NorthAddress string
	SouthAddress string
}

type ruContent struct {
	LocalAddress string
	NorthAddress string
}

type Port struct {
	number   int32
	protocol v1.Protocol
}

var SplitPorts = map[SplitPiece][]Port{
	CU: {{501, v1.ProtocolUDP}, {601, v1.ProtocolUDP}, {2152, v1.ProtocolUDP},
		{36412, v1.ProtocolUDP}, {36422, v1.ProtocolUDP}, {30923, v1.ProtocolUDP},
		{37659, v1.ProtocolTCP}},
	DU: {{500, v1.ProtocolUDP}, {600, v1.ProtocolUDP}, {30923, v1.ProtocolUDP},
		{34878, v1.ProtocolUDP}, {45501, v1.ProtocolTCP}, {50001, v1.ProtocolUDP},
		{50011, v1.ProtocolUDP}},
	RU: {{8888, v1.ProtocolUDP}, {9999, v1.ProtocolUDP}, {10000, v1.ProtocolUDP},
		{32123, v1.ProtocolUDP}, {38927, v1.ProtocolTCP}, {50000, v1.ProtocolUDP},
		{50010, v1.ProtocolUDP}, {58363, v1.ProtocolUDP}},
}

var TemplateConfigMaps = map[SplitPiece]string{
	CU: CUTemplateConfigMapName,
	DU: DUTemplateConfigMapName,
	RU: RUTemplateConfigMapName,
}

var Splits = utils.NewStringSet(
	string(CU),
	string(RU),
	string(DU),
)
