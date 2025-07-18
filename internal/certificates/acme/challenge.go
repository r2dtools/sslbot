package acme

const (
	HttpChallengeTypeCode = "http"
	DnsChallengeTypeCode  = "dns"
)

type ChallengeType interface {
	GetParams() []string
}
