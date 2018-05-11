package gotmpl

type Lookup interface {
	Resolve(variable string) (string, bool)
}

type MapLookup map[string]string

func (m MapLookup) Resolve(s string) (string, bool) {
	res, ok := m[s]
	return res, ok
}
