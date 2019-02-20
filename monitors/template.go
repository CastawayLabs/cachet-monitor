package monitors

import (
	"bytes"
	"text/template"
)

type MonitorTemplates struct {
	Investigating MessageTemplate
	Fixed         MessageTemplate
}

type MessageTemplate struct {
	Subject string `json:"subject"`
	Message string `json:"message"`

	subjectTpl *template.Template
	messageTpl *template.Template
}

func (t *MessageTemplate) SetDefault(d MessageTemplate) {
	if len(t.Subject) == 0 {
		t.Subject = d.Subject
	}
	if len(t.Message) == 0 {
		t.Message = d.Message
	}
}

// TODO: test
func (t *MessageTemplate) Compile() error {
	var err error

	if len(t.Subject) > 0 {
		t.subjectTpl, err = compileTemplate(t.Subject)
	}

	if err == nil && len(t.Message) > 0 {
		t.messageTpl, err = compileTemplate(t.Message)
	}

	return err
}

func (t *MessageTemplate) Exec(data interface{}) (string, string) {
	return t.exec(t.subjectTpl, data), t.exec(t.messageTpl, data)
}

func (t *MessageTemplate) exec(tpl *template.Template, data interface{}) string {
	buf := new(bytes.Buffer)

	tpl.Execute(buf, data)
	return buf.String()
}

func compileTemplate(text string) (*template.Template, error) {
	return template.New("").Parse(text)
}
