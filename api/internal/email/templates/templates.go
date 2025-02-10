//nolint:gochecknoglobals // it's okay
package email

import (
	"api/config"
	"bytes"
	"fmt"
	"html/template"
	"os"
)

type Template[T any] struct {
	Name   string
	Render func(data *T) string
}

type ConfirmEmailTemplateData struct {
	ConfirmURL string
	Code       string
}

func NewConfirmEmailTemplateData(cfg *config.Config, email string, code string) *ConfirmEmailTemplateData {
	return &ConfirmEmailTemplateData{
		ConfirmURL: fmt.Sprintf("%s://%s/?hash=%s&email=%s", cfg.AppProtocol, cfg.AppHost, code, email),
		Code:       code,
	}
}

type ResetPasswordEmailTemplateData struct {
	ResetPasswordURL string
}

type ProjectInviteTemplateData struct {
	JoinProjectURL string
	ProjectName    string
	UserName       string
}

const (
	ConfirmEmailTemplateName       = "confirm-email"
	ResetPasswordEmailTemplateName = "reset-password-email"
	ProjectInviteTemplateName      = "project-invite"
)

var ConfirmEmailTemplate = Template[ConfirmEmailTemplateData]{
	Name:   ConfirmEmailTemplateName,
	Render: makeRenderEmailFunc[ConfirmEmailTemplateData](ConfirmEmailTemplateName),
}

var ResetPasswordEmailTemplate = Template[ResetPasswordEmailTemplateData]{
	Name:   ResetPasswordEmailTemplateName,
	Render: makeRenderEmailFunc[ResetPasswordEmailTemplateData](ResetPasswordEmailTemplateName),
}

var ProjectInviteTemplate = Template[ProjectInviteTemplateData]{
	Name:   ProjectInviteTemplateName,
	Render: makeRenderEmailFunc[ProjectInviteTemplateData](ProjectInviteTemplateName),
}

var TemplateNames = []string{ConfirmEmailTemplate.Name, ResetPasswordEmailTemplate.Name, ProjectInviteTemplate.Name}

func renderEmailTemplate(templateName string, data interface{}) (string, error) {
	dat, err := os.ReadFile("./internal/email/templates/" + templateName + ".html")
	if err != nil {
		return "", err
	}
	email := string(dat)
	tmpl, err := template.New(templateName).Parse(email)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func mustRenderEmailTemplate(templateName string, data interface{}) string {
	tmpl, err := renderEmailTemplate(templateName, data)
	if err != nil {
		panic(err)
	}
	return tmpl
}

func makeRenderEmailFunc[T any](templateName string) func(data *T) string {
	return func(data *T) string {
		return mustRenderEmailTemplate(templateName, data)
	}
}
