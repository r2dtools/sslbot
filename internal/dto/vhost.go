package dto

type VirtualHost struct {
	FilePath    string
	ServerName  string
	DocRoot     string
	WebServer   string
	Aliases     []string
	Ssl         bool
	Addresses   []VirtualHostAddress
	Certificate *Certificate
}

type VirtualHostAddress struct {
	IsIpv6 bool
	Host   string
	Port   string
}
