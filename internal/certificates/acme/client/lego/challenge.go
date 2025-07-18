package lego

import "fmt"

type HTTPChallengeType struct {
	HTTPPort,
	TLSPort int
	WebRoot string
}

type DNSChallengeType struct {
	Provider string
}

func (ct *HTTPChallengeType) GetParams() []string {
	return []string{"--http", fmt.Sprintf("--http.port=%d", ct.HTTPPort), fmt.Sprintf("--tls.port=%d", ct.TLSPort), "--http.webroot=" + ct.WebRoot}
}

func (ct *DNSChallengeType) GetParams() []string {
	return []string{"--dns=" + ct.Provider}
}
