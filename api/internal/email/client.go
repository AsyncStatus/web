package email

import (
	"api/config"
	"api/internal/log"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

type (
	LocalClient struct {
		cfg    *config.Config
		logger *log.Logger
	}
	ProdClient struct {
		*resend.Client
	}
	Client struct {
		cfg *config.Config
		LocalClient
		ProdClient
	}
)

func NewLocalClient(cfg *config.Config, logger *log.Logger) *LocalClient {
	return &LocalClient{cfg, logger}
}

func NewProdClient(cfg *config.Config) *ProdClient {
	return &ProdClient{resend.NewClient(cfg.ResendAPIKey)}
}

func NewClient(cfg *config.Config, logger *log.Logger) *Client {
	return &Client{cfg: cfg, LocalClient: *NewLocalClient(cfg, logger), ProdClient: *NewProdClient(cfg)}
}

func (c *LocalClient) Send(email *resend.SendEmailRequest) error {
	c.logger.Info("Sending email",
		zap.String("From", email.From),
		zap.String("To", strings.Join(email.To, ", ")),
	)

	subject := fmt.Sprintf("Subject: %s\n", email.Subject)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(subject + mime + email.Html)

	err := smtp.SendMail(c.cfg.SMTPAddr, nil, email.From, email.To, msg)
	if err != nil {
		return err
	}

	return nil
}

func (c *ProdClient) Send(email *resend.SendEmailRequest) error {
	if email.Headers == nil {
		email.Headers = map[string]string{"X-Entity-Ref-ID": uuid.NewString()}
	} else {
		email.Headers["X-Entity-Ref-ID"] = uuid.NewString()
	}
	_, err := c.Client.Emails.Send(email)
	return err
}

func (c *Client) Send(email *resend.SendEmailRequest) error {
	// if c.cfg.AppEnv == env.Local {
	// 	return c.LocalClient.Send(email)
	// }

	return c.ProdClient.Send(email)
}
