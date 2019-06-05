package gmail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

/* ================================================================================
 * 邮件发送
 * qq group: 582452342
 * email   : 2091938785@qq.com
 * author  : 美丽的地球啊
 * ================================================================================ */
type (
	MailClient interface {
		Send(email *Email) error
	}

	mailService struct {
		Config *MailConfig
		Auth   smtp.Auth
	}

	MailConfig struct {
		Host     string `form:"host" json:"host"`
		Port     int32  `form:"port" json:"port"`
		Username string `form:"username" json:"username"`
		Password string `form:"password" json:"password"`
		IsSsl    bool   `form:"is_ssl" json:"is_ssl"`
	}

	Email struct {
		Subject string   `form:"subject" json:"subject"`
		To      []string `form:"to" json:"to"`
		Content string   `form:"content" json:"content"`
		IsHtml  bool     `form:"is_html" json:"is_html"` //是否HTML内容
	}
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 初始化邮件发送客户端
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func NewMailClient(config *MailConfig) MailClient {
	mService := &mailService{
		Config: config,
		Auth:   smtp.PlainAuth("", config.Username, config.Password, config.Host),
	}

	if mService.Config.Port == 0 {
		mService.Config.Port = 25
	}

	return mService
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 发送邮件
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *mailService) Send(email *Email) error {
	var err error
	if s.Config.IsSsl {
		err = s.sendSslMail(email)
	} else {
		err = s.sendMail(email)
	}

	return err
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 发送邮件
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *mailService) sendMail(email *Email) error {
	header := s.getMailHeader(email)
	content := header + "\r\n" + email.Content
	smtpAddr := s.Config.Host + ":" + fmt.Sprintf("%d", s.Config.Port)

	//发送邮件
	err := smtp.SendMail(
		smtpAddr,
		s.Auth,
		s.Config.Username,
		email.To,
		[]byte(content),
	)

	if err != nil {
		log.Printf("gmail send error: %v", err)
	}

	return err
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 发送Ssl邮件
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *mailService) sendSslMail(email *Email) error {
	header := s.getMailHeader(email)
	content := header + "\r\n" + email.Content
	smtpAddr := s.Config.Host + ":" + fmt.Sprintf("%d", s.Config.Port)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.Config.Host,
	}

	conn, err := tls.Dial("tcp", smtpAddr, tlsconfig)
	if err != nil {
		return fmt.Errorf("tls dial error:%v", err)
	}

	smtpClient, err := smtp.NewClient(conn, s.Config.Host)
	if err != nil {
		return fmt.Errorf("smtp client error: %v", err)
	}
	defer smtpClient.Close()

	if s.Auth != nil {
		if ok, _ := smtpClient.Extension("AUTH"); ok {
			if err = smtpClient.Auth(s.Auth); err != nil {
				return fmt.Errorf("smtp client auth error: %v", err)
			}
		}
	}

	if err := smtpClient.Mail(s.Config.Username); err != nil {
		return fmt.Errorf("smtp client mail error: %v", err)
	}

	for _, toUser := range email.To {
		if err = smtpClient.Rcpt(toUser); err != nil {
			return fmt.Errorf("smtp client rcpt error: %v", err)
		}
	}

	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("smtp client data error: %v", err)
	}

	if _, err := w.Write([]byte(content)); err != nil {
		return fmt.Errorf("smtp client write body error: %v", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp client close body error:%v", err)
	}

	return smtpClient.Quit()
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 获取邮件头部
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func (s *mailService) getMailHeader(email *Email) string {
	headerContent := ""
	header := make(map[string]string, 0)

	contentType := "text/plain; charset=UTF-8"
	if email.IsHtml {
		contentType = "text/html; charset=UTF-8"
	}

	header["To"] = strings.Join(email.To, ",")
	header["From"] = s.Config.Username
	header["Subject"] = email.Subject
	header["Content-Type"] = contentType

	for k, v := range header {
		headerContent += k + ": " + v + "\r\n"
	}

	return headerContent
}
