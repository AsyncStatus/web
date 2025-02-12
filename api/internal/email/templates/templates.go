package email

import (
	"api/config"
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"os"
)

type Template[T any] struct {
	Name   string
	Render func(data *T) string
}

type CreateAccountTemplateData struct {
	Name             string
	CreateAccountURL string
}

func NewCreateAccountTemplateData(cfg *config.Config, name string, email string, token string) *CreateAccountTemplateData {
	return &CreateAccountTemplateData{
		Name:             name,
		CreateAccountURL: fmt.Sprintf("%s://%s/sign-up?token=%s&email=%s", cfg.AppProtocol, cfg.AppHost, url.QueryEscape(token), url.QueryEscape(email)),
	}
}

const (
	CreateAccountTemplateName = "create-account"
)

var CreateAccountTemplate = Template[CreateAccountTemplateData]{
	Name:   CreateAccountTemplateName,
	Render: makeRenderEmailFunc[CreateAccountTemplateData](CreateAccountTemplateName),
}

var TemplateNames = []string{
	CreateAccountTemplate.Name,
}

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
