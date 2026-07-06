package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/rajchodisetti/restaurant-platform/backend/internal/platform/config"
)

type smtpProvider struct {
	cfg      config.EmailConfig
	smtpCfg  config.SMTPConfig
}

func NewSMTP(emailCfg config.EmailConfig, smtpCfg config.SMTPConfig) (Provider, error) {
	return &smtpProvider{cfg: emailCfg, smtpCfg: smtpCfg}, nil
}

func (provider *smtpProvider) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if err := ctx.Err(); err != nil {
		return SendResult{}, err
	}

	to := strings.TrimSpace(req.To)
	originalTo := to
	if redirect := strings.TrimSpace(provider.cfg.RedirectTo); redirect != "" {
		to = redirect
	}

	from := strings.TrimSpace(provider.cfg.FromAddress)
	fromName := strings.TrimSpace(provider.cfg.FromName)
	fromHeader := from
	if fromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", encodeHeaderWord(fromName), from)
	}

	message := buildMIMEMessage(fromHeader, to, req.Subject, req.TextBody, req.HTMLBody, req.ReplyTo)

	addr := fmt.Sprintf("%s:%d", provider.smtpCfg.Host, provider.smtpCfg.Port)
	auth := smtp.PlainAuth("", provider.smtpCfg.Username, provider.smtpCfg.Password, provider.smtpCfg.Host)

	if provider.smtpCfg.UseTLS {
		if err := sendWithSTARTTLS(addr, auth, from, []string{to}, message, provider.smtpCfg.Host); err != nil {
			return SendResult{}, err
		}
	} else {
		if err := smtp.SendMail(addr, auth, from, []string{to}, message); err != nil {
			return SendResult{}, fmt.Errorf("smtp send: %w", err)
		}
	}

	result := SendResult{ProviderMessageID: "smtp:" + originalTo}
	if to != originalTo {
		result.RedirectedTo = to
	}
	return result, nil
}

func sendWithSTARTTLS(addr string, auth smtp.Auth, from string, to []string, message []byte, host string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp close data: %w", err)
	}
	return client.Quit()
}

func buildMIMEMessage(from, to, subject, textBody, htmlBody, replyTo string) []byte {
	boundary := "tuvi-email-boundary"
	var builder strings.Builder
	builder.WriteString("From: " + from + "\r\n")
	builder.WriteString("To: " + to + "\r\n")
	if replyTo != "" {
		builder.WriteString("Reply-To: " + replyTo + "\r\n")
	}
	builder.WriteString("Subject: " + encodeHeaderWord(subject) + "\r\n")
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n")
	builder.WriteString("\r\n")

	builder.WriteString("--" + boundary + "\r\n")
	builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(textBody)
	builder.WriteString("\r\n")

	builder.WriteString("--" + boundary + "\r\n")
	builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(htmlBody)
	builder.WriteString("\r\n")

	builder.WriteString("--" + boundary + "--\r\n")
	return []byte(builder.String())
}

func encodeHeaderWord(value string) string {
	if value == "" {
		return value
	}
	for _, r := range value {
		if r > 127 {
			return fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(value)))
		}
	}
	return value
}
