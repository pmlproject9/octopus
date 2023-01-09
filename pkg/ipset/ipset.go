package ipset

type Params struct {
	HashFamily string
	HashSize   int
	MaxElem    int
	Timeout    int
}

type IPSet struct {
	Name       string
	HashType   string
	HashFamily string
	HashSize   int
	MaxElem    int
	Timeout    int
}

func New(name string, hashtype string, param *Params) *IPSet {
	if param == nil {
		param = &Params{}
	}

	if param.HashFamily == "" {
		param.HashFamily = ProtocolFamilyIPV4
	}
	if param.HashSize == 0 {
		param.HashSize = DefaultHashSize
	}
	if param.MaxElem == 0 {
		param.MaxElem = DefaultMaxElem
	}

	ipset := &IPSet{
		Name:       name,
		HashType:   hashtype,
		HashFamily: param.HashFamily,
		HashSize:   param.HashSize,
		MaxElem:    param.MaxElem,
	}
	return ipset
}
