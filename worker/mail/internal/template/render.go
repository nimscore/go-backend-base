package template

import (
	"bytes"
	"fmt"
	"os"
	"text/template"
)

func Render(path string, values any) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	template, err := template.New("test").Parse(string(content))
	if err != nil {
		return "", fmt.Errorf("can't create template: %w", err)
	}

	err = template.Execute(&buffer, values)
	if err != nil {
		return "", fmt.Errorf("can't render template: %w", err)
	}

	return buffer.String(), err
}
