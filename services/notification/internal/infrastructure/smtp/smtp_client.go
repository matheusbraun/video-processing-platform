package smtp

import (
	"fmt"
	"net/smtp"
)

type SMTPClient interface {
	SendEmail(to, subject, body string) error
}

type smtpClient struct {
	host     string
	port     string
	user     string
	password string
	from     string
}

func NewSMTPClient(host, port, user, password, from string) SMTPClient {
	return &smtpClient{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

func (c *smtpClient) SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", c.user, c.password, c.host)

	msg := fmt.Appendf(nil, "From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", c.from, to, subject, body)

	addr := fmt.Sprintf("%s:%s", c.host, c.port)

	return smtp.SendMail(addr, auth, c.from, []string{to}, msg)
}
