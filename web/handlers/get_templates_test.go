package handlers_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/notifications/fakes"
	"github.com/cloudfoundry-incubator/notifications/models"
	"github.com/cloudfoundry-incubator/notifications/web/handlers"
	"github.com/ryanmoran/stack"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetTemplates", func() {
	var handler handlers.GetTemplates
	var request *http.Request
	var writer *httptest.ResponseRecorder
	var context stack.Context
	var finder *fakes.TemplateFinder
	var errorWriter *fakes.ErrorWriter

	Describe("ServeHTTP", func() {
		var templateID string

		BeforeEach(func() {
			finder = fakes.NewTemplateFinder()
			templateID = "theTemplateID"

			finder.Templates[templateID] = models.Template{
				Name:     "The Name of The Template",
				Subject:  "All about the {{.Subject}}",
				Text:     "the template {{variable}}",
				HTML:     "<p> the template {{variable}} </p>",
				Metadata: `{"hello": "world"}`,
			}
			writer = httptest.NewRecorder()
			errorWriter = fakes.NewErrorWriter()
			handler = handlers.NewGetTemplates(finder, errorWriter)
		})

		Context("When the finder does not error", func() {
			BeforeEach(func() {

				var err error
				request, err = http.NewRequest("GET", "/templates/"+templateID, bytes.NewBuffer([]byte{}))
				if err != nil {
					panic(err)
				}
			})

			It("Calls find on its finder with appropriate arguments", func() {
				handler.ServeHTTP(writer, request, context)
				Expect(finder.TemplateID).To(Equal(templateID))
			})

			It("writes out the finder's response", func() {
				handler.ServeHTTP(writer, request, context)
				Expect(writer.Code).To(Equal(http.StatusOK))

				var template map[string]interface{}

				err := json.Unmarshal(writer.Body.Bytes(), &template)
				if err != nil {
					panic(err)
				}

				Expect(template).To(HaveLen(5))
				Expect(template["name"]).To(Equal("The Name of The Template"))
				Expect(template["subject"]).To(Equal("All about the {{.Subject}}"))
				Expect(template["text"]).To(Equal("the template {{variable}}"))
				Expect(template["html"]).To(Equal("<p> the template {{variable}} </p>"))
				Expect(template["metadata"]).To(Equal(map[string]interface{}{"hello": "world"}))
			})
		})

		Context("When the finder errors", func() {
			BeforeEach(func() {
				finder.FindError = errors.New("BOOM!")

				var err error
				request, err = http.NewRequest("GET", "/templates/someTemplateID", bytes.NewBuffer([]byte{}))
				if err != nil {
					panic(err)
				}
			})

			It("writes the error to the errorWriter", func() {
				handler.ServeHTTP(writer, request, context)
				Expect(errorWriter.Error).To(Equal(errors.New("BOOM!")))
			})
		})
	})
})
