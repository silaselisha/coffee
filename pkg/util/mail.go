package util

import (
	"context"
	"fmt"
	"net/smtp"
)

type Transporter interface {
	MailSender(ctx context.Context, reciver string, message []byte) error
}

func NewSMTPTransporter(envs *Config) Transporter {
	return &SMTPTransport{
		Host: envs.SMTP_HOST,
		Username: envs.SMTP_USERNAME,
		Password: envs.SMTP_PASSWORD,
		Port: envs.SMTP_PORT,
		Sender: envs.SMTP_SENDER,
	}
}

func (stp *SMTPTransport) MailSender(ctx context.Context, receiver string, message []byte) error {
	auth := smtp.PlainAuth("", stp.Sender, stp.Password, stp.Host)
	SMTP_URL := fmt.Sprintf("%s:%s", stp.Host, stp.Port)
	fmt.Println(SMTP_URL)
	to := []string{receiver}
	err := smtp.SendMail(SMTP_URL, auth, stp.Sender, to, message)
	
	if err != nil {
		return err
	}
	return nil
}