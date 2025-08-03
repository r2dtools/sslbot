package certbot

type HTTPChallengeType struct {
	WebRoot string
}

type DNSChallengeType struct {
	Provider string
}

func (ct HTTPChallengeType) GetParams() []string {
	return []string{"-w", ct.WebRoot}
}
