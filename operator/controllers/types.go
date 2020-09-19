package controllers

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	ResyncPeriod = 20 * time.Second

	CU SplitPiece = "cu"
	DU SplitPiece = "du"
	RU SplitPiece = "ru"

	CUTemplateConfigMapName = "operator-cu-template"
	DUTemplateConfigMapName = "operator-du-template"
	RUTemplateConfigMapName = "operator-ru-template"

	operatorNamespace          = "operator-system"
	cuConfigMapContentTemplate = "upfaddress: %s\nlocaladdress: %s\nsouthaddress: %s\n"
	duConfigMapContentTemplate = "northaddress: %s\nlocaladdress: %s\nsouthaddress: %s\n"
	ruConfigMapContentTemplate = "northaddress: %s\nlocaladdress: %s\n"
)

type SplitPiece string
type StringSet map[string]struct{}

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

type port struct {
	number   int32
	protocol v1.Protocol
}

var SplitPorts = map[SplitPiece][]port{
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

var Empty struct{}

var Splits = NewStringSet(
	string(CU),
	string(RU),
	string(DU),
)

func NewStringSet(values ...string) StringSet {
	stringSet := make(StringSet)
	for _, v := range values {
		stringSet[v] = Empty
	}
	return stringSet
}

// Add adds new values to the set.
func (s *StringSet) Add(items ...string) {
	for _, item := range items {
		(*s)[item] = Empty
	}
}

// Has returns true if item is in the Set
func (s StringSet) Has(item string) bool {
	_, contained := s[item]
	return contained
}
