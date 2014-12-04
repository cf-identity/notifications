package acceptance

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/cloudfoundry-incubator/notifications/acceptance/servers"
	"github.com/cloudfoundry-incubator/notifications/config"
	"github.com/cloudfoundry-incubator/notifications/models"
	"github.com/pivotal-cf/uaa-sso-golang/uaa"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Templates GET Endpoint", func() {

	BeforeEach(func() {
		TruncateTables()

		env := config.NewEnvironment()
		migrationsPath := path.Join(env.RootPath, env.ModelMigrationsDir)
		database := models.NewDatabase(env.DatabaseURL, migrationsPath)

		templateData := models.Template{
			Name: "overridden-client." + models.UserBodyTemplateName,
			Text: "Text Template",
			HTML: "<p>HTML Template</p>",
		}
		database.Connection().Insert(&templateData)
	})

	It("allows a user to get body templates", func() {
		// Boot Fake SMTP Server
		smtpServer := servers.NewSMTP()
		smtpServer.Boot()

		// Boot Fake UAA Server
		uaaServer := servers.NewUAA()
		uaaServer.Boot()
		defer uaaServer.Close()

		// Boot Real Notifications Server
		notificationsServer := servers.NewNotifications()
		notificationsServer.Boot()
		defer notificationsServer.Close()

		// Retrieve Client UAA token
		overriddenClientID := "overridden-client"
		clientID := "notifications-sender"
		env := config.NewEnvironment()
		uaaClient := uaa.NewUAA("", env.UAAHost, clientID, "secret", "")
		clientToken, err := uaaClient.GetClientToken()
		if err != nil {
			panic(err)
		}

		kindID := "spam-email"
		test := GetTemplates{}
		test.GetUserTemplates(notificationsServer, clientToken)
		test.GetSpaceTemplates(notificationsServer, clientToken)
		test.GetEmailTemplates(notificationsServer, clientToken)
		test.GetUserTemplatesForClient(notificationsServer, clientToken, clientID)
		test.GetUserTemplatesForClientAndKind(notificationsServer, clientToken, clientID, kindID)
		test.GetUserTemplatesForOverriddenClient(notificationsServer, clientToken, overriddenClientID)
	})
})

type GetTemplates struct{}

func (t GetTemplates) GetSpaceTemplates(notificationsServer servers.Notifications, clientToken uaa.Token) {
	request, err := http.NewRequest("GET", notificationsServer.TemplatePath(models.SpaceBodyTemplateName), bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+clientToken.Access)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	// Confirm response status code looks ok
	Expect(response.StatusCode).To(Equal(http.StatusOK))

	// Confirm we got the correct template info
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	responseJSON := models.Template{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		panic(err)
	}

	Expect(responseJSON.Text).To(Equal(`Hello {{.To}},

The following "{{.KindDescription}}" notification was sent to you by the "{{.SourceDescription}}"
component of Cloud Foundry because you are a member of the "{{.Space}}" space
in the "{{.Organization}}" organization:

{{.Text}}

This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
notification can be identified with the {{.MessageID}} identifier and was sent
with the {{.ClientID}} UAA client. The notification can be unsubscribed from
using the "{{.UnsubscribeID}}" unsubscribe token.
`))

	Expect(responseJSON.HTML).To(Equal(`<p>Hello {{.To}},</p>

<p>The following "{{.KindDescription}}" notification was sent to you by the "{{.SourceDescription}}"
    component of Cloud Foundry because you are a member of the "{{.Space}}" space
    in the "{{.Organization}}" organization:</p>

{{.HTML}}

<p>This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
    notification can be identified with the {{.MessageID}} identifier and was sent
    with the {{.ClientID}} UAA client. The notification can be unsubscribed from
    using the "{{.UnsubscribeID}}" unsubscribe token.</p>
`))

}

func (t GetTemplates) GetUserTemplates(notificationsServer servers.Notifications, clientToken uaa.Token) {
	request, err := http.NewRequest("GET", notificationsServer.TemplatePath(models.UserBodyTemplateName), bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+clientToken.Access)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// Confirm response status code looks ok
	Expect(response.StatusCode).To(Equal(http.StatusOK))

	// Confirm we got the correct template info
	responseJSON := models.Template{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		panic(err)
	}

	Expect(responseJSON.Text).To(Equal(`Hello {{.To}},

The following "{{.KindDescription}}" notification was sent to you directly by the
"{{.SourceDescription}}" component of Cloud Foundry:

{{.Text}}

This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
notification can be identified with the {{.MessageID}} identifier and was sent
with the {{.ClientID}} UAA client. The notification can be unsubscribed from
using the "{{.UnsubscribeID}}" unsubscribe token.
`))

	Expect(responseJSON.HTML).To(Equal(`<p>Hello {{.To}},</p>

<p>The following "{{.KindDescription}}" notification was sent to you directly by the
    "{{.SourceDescription}}" component of Cloud Foundry:</p>

{{.HTML}}

<p>This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
    notification can be identified with the {{.MessageID}} identifier and was sent
    with the {{.ClientID}} UAA client. The notification can be unsubscribed from
    using the "{{.UnsubscribeID}}" unsubscribe token.</p>
`))

}

func (t GetTemplates) GetEmailTemplates(notificationsServer servers.Notifications, clientToken uaa.Token) {
	request, err := http.NewRequest("GET", notificationsServer.TemplatePath(models.EmailBodyTemplateName), bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+clientToken.Access)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// Confirm response status code looks ok
	Expect(response.StatusCode).To(Equal(http.StatusOK))

	// Confirm we got the correct template info
	responseJSON := models.Template{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		panic(err)
	}

	Expect(responseJSON.Text).To(Equal(`Hello {{.To}},

The following "{{.Subject}}" notification was sent to you directly by the "{{.SourceDescription}}"
component of Cloud Foundry:

{{.Text}}

This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
notification can be identified with the {{.MessageID}} identifier and was sent
with the {{.ClientID}} UAA client. The notification can be unsubscribed from
using the "{{.UnsubscribeID}}" unsubscribe token.
`))

	Expect(responseJSON.HTML).To(Equal(`<p>Hello {{.To}},</p>

<p>The following "{{.Subject}}" notification was sent to you directly by the "{{.SourceDescription}}"
    component of Cloud Foundry:</p>

{{.HTML}}

<p>This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
    notification can be identified with the {{.MessageID}} identifier and was sent
    with the {{.ClientID}} UAA client. The notification can be unsubscribed from
    using the "{{.UnsubscribeID}}" unsubscribe token.</p>
`))

}

func (t GetTemplates) GetUserTemplatesForOverriddenClient(notificationsServer servers.Notifications, clientToken uaa.Token, clientID string) {
	request, err := http.NewRequest("GET", notificationsServer.TemplatePath(clientID+"."+models.UserBodyTemplateName), bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+clientToken.Access)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// Confirm response status code looks ok
	Expect(response.StatusCode).To(Equal(http.StatusOK))

	// Confirm we got the correct template info
	responseJSON := models.Template{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		panic(err)
	}

	Expect(responseJSON.Text).To(Equal("Text Template"))

	Expect(responseJSON.HTML).To(Equal("<p>HTML Template</p>"))

	request, err = http.NewRequest("GET", notificationsServer.TemplatePath(models.UserBodyTemplateName), bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+clientToken.Access)

	response, err = http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	responseJSON = models.Template{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		panic(err)
	}
}

func (t GetTemplates) GetUserTemplatesForClient(notificationsServer servers.Notifications, clientToken uaa.Token, clientID string) {
	request, err := http.NewRequest("GET", notificationsServer.TemplatePath(clientID+"."+models.UserBodyTemplateName), bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+clientToken.Access)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// Confirm response status code looks ok
	Expect(response.StatusCode).To(Equal(http.StatusOK))

	// Confirm we got the correct template info
	responseJSON := models.Template{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		panic(err)
	}

	Expect(responseJSON.Text).To(Equal(`Hello {{.To}},

The following "{{.KindDescription}}" notification was sent to you directly by the
"{{.SourceDescription}}" component of Cloud Foundry:

{{.Text}}

This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
notification can be identified with the {{.MessageID}} identifier and was sent
with the {{.ClientID}} UAA client. The notification can be unsubscribed from
using the "{{.UnsubscribeID}}" unsubscribe token.
`))

	Expect(responseJSON.HTML).To(Equal(`<p>Hello {{.To}},</p>

<p>The following "{{.KindDescription}}" notification was sent to you directly by the
    "{{.SourceDescription}}" component of Cloud Foundry:</p>

{{.HTML}}

<p>This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
    notification can be identified with the {{.MessageID}} identifier and was sent
    with the {{.ClientID}} UAA client. The notification can be unsubscribed from
    using the "{{.UnsubscribeID}}" unsubscribe token.</p>
`))

}

func (t GetTemplates) GetUserTemplatesForClientAndKind(notificationsServer servers.Notifications, clientToken uaa.Token, clientID, kindID string) {
	request, err := http.NewRequest("GET", notificationsServer.TemplatePath(clientID+"."+kindID+"."+models.UserBodyTemplateName), bytes.NewBuffer([]byte{}))
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+clientToken.Access)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	// Confirm response status code looks ok
	Expect(response.StatusCode).To(Equal(http.StatusOK))

	// Confirm we got the correct template info
	responseJSON := models.Template{}
	err = json.Unmarshal(body, &responseJSON)
	if err != nil {
		panic(err)
	}

	Expect(responseJSON.Text).To(Equal(`Hello {{.To}},

The following "{{.KindDescription}}" notification was sent to you directly by the
"{{.SourceDescription}}" component of Cloud Foundry:

{{.Text}}

This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
notification can be identified with the {{.MessageID}} identifier and was sent
with the {{.ClientID}} UAA client. The notification can be unsubscribed from
using the "{{.UnsubscribeID}}" unsubscribe token.
`))

	Expect(responseJSON.HTML).To(Equal(`<p>Hello {{.To}},</p>

<p>The following "{{.KindDescription}}" notification was sent to you directly by the
    "{{.SourceDescription}}" component of Cloud Foundry:</p>

{{.HTML}}

<p>This message was sent from {{.From}} and can be replied to at {{.ReplyTo}}. The
    notification can be identified with the {{.MessageID}} identifier and was sent
    with the {{.ClientID}} UAA client. The notification can be unsubscribed from
    using the "{{.UnsubscribeID}}" unsubscribe token.</p>
`))

}
