package support

import (
	"bytes"
	"encoding/json"

	"github.com/cloudfoundry-incubator/notifications/web/params"
)

type TemplatesService struct {
	client *Client
}

type Template struct {
	Name     string                 `json:"name"`
	Subject  string                 `json:"subject"`
	Text     string                 `json:"text"`
	HTML     string                 `json:"html"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (t TemplatesService) Create(token string, template params.Template) (int, string, error) {
	var status int
	var templateID string

	body, err := json.Marshal(template)
	if err != nil {
		return status, templateID, err
	}

	request, err := t.client.makeRequest("POST", t.client.server.TemplatesBasePath(), bytes.NewBuffer(body), token)
	if err != nil {
		return status, templateID, err
	}

	status, body, err = t.client.do(request)
	if err != nil {
		return status, templateID, err
	}

	var JSON struct {
		TemplateID string `json:"template_id"`
	}

	err = json.Unmarshal(body, &JSON)
	if err != nil {
		return status, templateID, err
	}

	return status, JSON.TemplateID, nil
}

func (t TemplatesService) AssignToClient(token, clientID, templateID string) (int, error) {
	var status int

	body, err := json.Marshal(map[string]string{
		"template": templateID,
	})
	if err != nil {
		return status, err
	}

	request, err := t.client.makeRequest("PUT", t.client.server.ClientsTemplatePath(clientID), bytes.NewBuffer(body), token)
	if err != nil {
		return status, err
	}

	status, _, err = t.client.do(request)
	if err != nil {
		return status, err
	}

	return status, nil
}

func (t TemplatesService) Get(token, templateID string) (int, Template, error) {
	var status int
	var template Template

	request, err := t.client.makeRequest("GET", t.client.server.TemplatePath(templateID), nil, token)
	if err != nil {
		return status, template, err
	}

	status, body, err := t.client.do(request)
	if err != nil {
		return status, template, err
	}

	err = json.Unmarshal(body, &template)
	if err != nil {
		return status, template, err
	}

	return status, template, nil
}