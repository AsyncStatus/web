package main

import (
	email "api/internal/email/templates"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	for _, templateName := range email.TemplateNames {
		if _, goFileErr := os.Stat("internal/email/templates/" + templateName + ".mjml"); errors.Is(goFileErr, os.ErrNotExist) {
			panic(
				errors.New("email template must have a corresponding .mjml file, but " + templateName + ".mjml does not exist"),
			)
		}

		fmt.Printf("generating html email template for %s.go\n", templateName) //nolint:forbidigo //it's alright

		_, execErr := exec.Command( //nolint:gosec //it's alright here
			"bash",
			"-c",
			fmt.Sprintf("mjml %s --output %s", "internal/email/templates/"+templateName+".mjml", "internal/email/templates/"+templateName+".html"),
		).Output()
		if execErr != nil {
			panic(execErr)
		}
	}
}
