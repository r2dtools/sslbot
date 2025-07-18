package contract

import (
	"github.com/r2dtools/agentintegration"
	"github.com/r2dtools/sslbot/internal/certificates/request"
)

func ConvertIssueRequest(r agentintegration.CertificateIssueRequestData) request.IssueRequest {
	return request.IssueRequest{
		Email:         r.Email,
		ServerName:    r.ServerName,
		WebServer:     r.WebServer,
		ChallengeType: r.ChallengeType,
		Subjects:      r.Subjects,
		Assign:        r.Assign,
		PreventReload: r.PreventReload,
	}
}

func ConvertAssignRequest(r agentintegration.CertificateAssignRequestData) request.AssignRequest {
	return request.AssignRequest{
		ServerName: r.ServerName,
		WebServer:  r.WebServer,
		CertName:   r.CertName,
	}
}

func ConvertUploadRequest(r agentintegration.CertificateUploadRequestData) request.UploadRequest {
	return request.UploadRequest{
		ServerName:     r.ServerName,
		WebServer:      r.WebServer,
		CertName:       r.CertName,
		PemCertificate: r.PemCertificate,
	}
}
