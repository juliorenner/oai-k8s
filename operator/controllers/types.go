package controllers

import "time"

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

var Empty struct{}
var TemplateConfigMaps = map[SplitPiece]string{
	CU: CUTemplateConfigMapName,
	DU: DUTemplateConfigMapName,
	RU: RUTemplateConfigMapName,
}

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
