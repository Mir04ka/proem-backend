package service

import (
	"fmt"
	"log"
	"net/smtp"

	"proem-backend/internal/config"
)

type MailerService interface {
	SendVerificationEmail(toEmail, token string) error
}

type mailer struct {
	cfg *config.Config
}

func NewMailerService(cfg *config.Config) MailerService {
	return &mailer{cfg: cfg}
}

func (m *mailer) SendVerificationEmail(toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", m.cfg.AppBaseUrl, token)

	// SMTP is not configured
	if m.cfg.SMTPHost == "" {
		log.Printf("\n=======================================================\n[DEV EMAIL] Кому: %s\nСсылка активации: %s\n=======================================================\n", toEmail, verifyURL)
		return nil
	}

	auth := smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPass, m.cfg.SMTPHost)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: Подтверждение регистрации\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n"+
			"<h3>Добро пожаловать!</h3><p>Для завершения регистрации перейдите по ссылке:</p><p><a href=\"%s\">Подтвердить Email</a></p>",
		m.cfg.SMTPFrom, toEmail, verifyURL,
	))

	addr := fmt.Sprintf("%s:%s", m.cfg.SMTPHost, m.cfg.SMTPPort)
	return smtp.SendMail(addr, auth, m.cfg.SMTPFrom, []string{toEmail}, msg)
}
