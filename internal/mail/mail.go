package mail

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/silaselisha/coffee-api/types"
)

type Transporter interface {
	MailSender(ctx context.Context, reciver string, message []byte) error
}

type SMTPTransport struct {
	Username string
	Password string
	Port     string
	Host     string
	Sender   string
}

func NewSMTPTransporter(envs *types.Config) Transporter {
	return &SMTPTransport{
		Host:     envs.SMTP_HOST,
		Username: envs.SMTP_USERNAME,
		Password: envs.SMTP_PASSWORD,
		Port:     envs.SMTP_PORT,
		Sender:   envs.SMTP_SENDER,
	}
}

// TODO: implement mail-sender exclusively to use SEND GRID
func (stp *SMTPTransport) MailSender(ctx context.Context, receiver string, message []byte) error {
	auth := smtp.PlainAuth("", stp.Username, stp.Password, stp.Host)
	SMTP_URL := fmt.Sprintf("%s:%s", stp.Host, stp.Port)
	to := []string{receiver}

	err := smtp.SendMail(SMTP_URL, auth, stp.Sender, to, message)
	if err != nil {
		return err
	}
	return nil
}
