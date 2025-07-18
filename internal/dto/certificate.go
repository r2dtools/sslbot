package dto

type Certificate struct {
	CN             string
	ValidFrom      string
	ValidTo        string
	DNSNames       []string
	EmailAddresses []string
	Organization   []string
	Province       []string
	Country        []string
	Locality       []string
	IsCA, IsValid  bool
	Issuer         Issuer
}

type Issuer struct {
	CN           string
	Organization []string
}
