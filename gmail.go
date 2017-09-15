package gmail

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

/* ================================================================================
 * 短信接口
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
	}

	MailConfig struct {
		Host     string `form:"host" json:"host"`
		Port     int32  `form:"port" json:"port"`
		Username string `form:"username" json:"username"`
		Password string `form:"password" json:"password"`
	}

	Email struct {
		Subject string   `form:"subject" json:"subject"`
		To      []string `form:"to" json:"to"`
		Content string   `form:"content" json:"content"`
		IsHtml  bool     `form:"is_html" json:"is_html"` //是否HTML内容
	}
)

/* ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++
 * 创建邮件发送客户端
 * ++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++ */
func NewMailClient(config *MailConfig) MailClient {
	mService := &mailService{
		Config: config,
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
	smtpAddr := s.Config.Host + ":" + fmt.Sprintf("%d", s.Config.Port)
	plainAuth := smtp.PlainAuth("", s.Config.Username, s.Config.Password, s.Config.Host)

	header := s.getMailHeader(email)
	content := header + "\r\n" + email.Content

	//发送邮件
	err := smtp.SendMail(
		smtpAddr,
		plainAuth,
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
