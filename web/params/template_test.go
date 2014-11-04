package params_test

import (
    "bytes"
    "encoding/json"

    "github.com/cloudfoundry-incubator/notifications/models"
    "github.com/cloudfoundry-incubator/notifications/web/params"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Template", func() {
    Describe("NewTemplate", func() {
        It("contructs parameters from a reader", func() {
            templateName := "user_body"
            body, err := json.Marshal(map[string]interface{}{
                "text": `its foobar of course`,
                "html": `<p>its foobar</p>`,
            })
            if err != nil {
                panic(err)
            }

            parameters, err := params.NewTemplate(templateName, bytes.NewBuffer(body))

            Expect(parameters).To(BeAssignableToTypeOf(params.Template{}))
            Expect(parameters.Name).To(Equal("user_body"))
            Expect(parameters.Text).To(Equal("its foobar of course"))
            Expect(parameters.HTML).To(Equal("<p>its foobar</p>"))
        })
    })

    Describe("Validate", func() {
        Context("when the name is valid", func() {
            It("returns no error", func() {
                bad_endings := []string{"user_body", "my.silly.space_body", "this.special.email_body", "emergency.email.subject.missing",
                    "subject.provided", "my.client.user_body", "client.space_body"}

                for _, ending := range bad_endings {
                    theTemplate := params.Template{
                        Name: ending,
                        Text: "its foobar of course",
                        HTML: "<p>its foobar</p>",
                    }
                    err := theTemplate.Validate()
                    Expect(err).ToNot(HaveOccurred())
                }
            })
        })

        Context("when the name is invalid", func() {
            It("returns an invalid name error", func() {
                bad_endings := []string{"user.body", "something_body", "subject.something", "still.missing.something",
                    "client.kind.otherkind.user_body", "stupid.stuff.subject.uh.oh.damn.email_body"}

                for _, ending := range bad_endings {
                    theTemplate := params.Template{
                        Name: ending,
                        Text: "its foobar of course",
                        HTML: "<p>its foobar</p>",
                    }
                    err := theTemplate.Validate()
                    Expect(err).To(HaveOccurred())
                }
            })
        })
    })

    Describe("ToModel", func() {
        It("turns a params.Template into a models.Template", func() {
            theTemplate := params.Template{
                Name: "user_body",
                Text: "its foobar of course",
                HTML: "<p>its foobar</p>",
            }
            theModel := theTemplate.ToModel()

            Expect(theModel).To(BeAssignableToTypeOf(models.Template{}))
            Expect(theModel.Name).To(Equal("user_body"))
            Expect(theModel.Text).To(Equal("its foobar of course"))
            Expect(theModel.HTML).To(Equal("<p>its foobar</p>"))
            Expect(theModel.Overridden).To(BeTrue())
            Expect(theModel.CreatedAt).To(BeZero())
        })
    })
})