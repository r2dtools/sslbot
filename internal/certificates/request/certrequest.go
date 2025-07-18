package request

type IssueRequest struct {
	Email         string
	ServerName    string
	WebServer     string
	ChallengeType string
	Subjects      []string
	Assign        bool
	PreventReload bool
}

type UploadRequest struct {
	ServerName     string
	WebServer      string
	CertName       string
	PemCertificate string
}

type AssignRequest struct {
	ServerName  string
	WebServer   string
	CertName    string
	StorageType string
}
