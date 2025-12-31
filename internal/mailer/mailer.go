package mailer

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// Sender 邮件发送器
type Sender struct {
	Host string
	Port int
	User string
	Pass string
	To   string
}

// NewSender 创建发送器
func NewSender(host string, port int, user, pass, to string) *Sender {
	return &Sender{
		Host: host,
		Port: port,
		User: user,
		Pass: pass,
		To:   to,
	}
}

// Send 发送邮件
func (s *Sender) Send(subject, body string) error {
	// 构建邮件内容
	message := []byte("Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"\r\n" + body)

	// 分割收件人
	recipients := strings.Split(s.To, ",")

	// 逐个发送
	for _, recipient := range recipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}

		if err := s.sendOne(recipient, message); err != nil {
			return fmt.Errorf("发送给 %s 失败: %w", recipient, err)
		}
	}

	return nil
}

// sendOne 发送给单个收件人
func (s *Sender) sendOne(recipient string, message []byte) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)

	// 建立 TLS 连接
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return fmt.Errorf("TLS 连接失败: %w", err)
	}
	defer conn.Close()

	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		return fmt.Errorf("创建 SMTP 客户端失败: %w", err)
	}
	defer client.Close()

	// 认证
	auth := smtp.PlainAuth("", s.User, s.Pass, s.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("认证失败: %w", err)
	}

	// 设置发件人
	if err := client.Mail(s.User); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("设置收件人失败: %w", err)
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("获取写入器失败: %w", err)
	}

	if _, err := w.Write(message); err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	return w.Close()
}
