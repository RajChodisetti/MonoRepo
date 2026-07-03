package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type Config struct {
	FromAddress string
	FromName    string
	Host        string
	Port        int
	Username    string
	Password    string
	UseTLS      bool
	Disabled    bool
}

type Message struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
	ReplyTo  string
}

type Sender struct {
	cfg Config
}

func NewSender(cfg Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) Send(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.cfg.Disabled {
		return nil
	}
	if strings.TrimSpace(s.cfg.Host) == "" {
		return fmt.Errorf("smtp host not configured")
	}

	to := strings.TrimSpace(msg.To)
	from := strings.TrimSpace(s.cfg.FromAddress)
	fromHeader := from
	if name := strings.TrimSpace(s.cfg.FromName); name != "" {
		fromHeader = fmt.Sprintf("%s <%s>", encodeHeaderWord(name), from)
	}

	mime := buildMIMEMessage(fromHeader, to, msg.Subject, msg.TextBody, msg.HTMLBody, msg.ReplyTo)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	var auth smtp.Auth
	if strings.TrimSpace(s.cfg.Username) != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	if s.cfg.UseTLS {
		return sendWithSTARTTLS(addr, auth, from, []string{to}, mime, s.cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, mime)
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
