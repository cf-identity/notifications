package mail_test

import (
	"bytes"
	"errors"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/notifications/mail"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mail", func() {
	var mailServer *SMTPServer
	var client *mail.Client
	var logger *log.Logger
	var buffer *bytes.Buffer
	var config mail.Config

	BeforeEach(func() {
		var err error

		buffer = bytes.NewBuffer([]byte{})
		logger = log.New(buffer, "", 0)
		mailServer = NewSMTPServer("user", "pass")

		config = mail.Config{
			User:          "user",
			Pass:          "pass",
			TestMode:      false,
			SkipVerifySSL: true,
			DisableTLS:    false,
		}

		config.Host, config.Port, err = net.SplitHostPort(mailServer.URL.String())
		if err != nil {
			panic(err)
		}

		client = mail.NewClient(config, logger)
	})

	AfterEach(func() {
		mailServer.Close()
	})

	Describe("NewClient", func() {
		It("defaults the ConnectTimeout to 15 seconds", func() {
			config.ConnectTimeout = 0

			client = mail.NewClient(config, logger)

			Expect(client.ConnectTimeout()).To(Equal(15 * time.Second))
		})
	})

	Describe("Send", func() {
		Context("when in Testmode", func() {
			var msg mail.Message

			BeforeEach(func() {
				var err error

				mailServer.SupportsTLS = true
				config.Host, config.Port, err = net.SplitHostPort(mailServer.URL.String())
				if err != nil {
					panic(err)
				}

				config.TestMode = true
				client = mail.NewClient(config, logger)

				msg = mail.Message{
					From:    "me@example.com",
					To:      "you@example.com",
					Subject: "Urgent! Read now!",
					Body: []mail.Part{
						{
							ContentType: "text/plain",
							Content:     "This email is the most important thing you will read all day!",
						},
					},
				}
			})

			It("does not connect to the smtp server", func() {
				err := client.Send(msg)
				if err != nil {
					panic(err)
				}

				Eventually(func() int {
					return len(mailServer.Deliveries)
				}).Should(Equal(0))
			})

			It("logs that it is in test mode", func() {
				err := client.Send(msg)
				if err != nil {
					panic(err)
				}

				line, err := buffer.ReadString('\n')
				if err != nil {
					panic(err)
				}

				Expect(line).To(Equal("TEST_MODE is enabled, emails not being sent\n"))
			})
		})
	})

	Context("when the SMTP server is properly configured", func() {
		BeforeEach(func() {
			var err error

			mailServer.SupportsTLS = true
			config.Host, config.Port, err = net.SplitHostPort(mailServer.URL.String())
			if err != nil {
				panic(err)
			}
			client = mail.NewClient(config, logger)
		})

		It("can send mail", func() {
			msg := mail.Message{
				From:    "me@example.com",
				To:      "you@example.com",
				Subject: "Urgent! Read now!",
				Body: []mail.Part{
					{
						ContentType: "text/plain",
						Content:     "This email is the most important thing you will read all day%40!",
					},
				},
			}

			err := client.Send(msg)
			if err != nil {
				panic(err)
			}

			Eventually(func() int {
				return len(mailServer.Deliveries)
			}).Should(Equal(1))
			delivery := mailServer.Deliveries[0]

			Expect(delivery.Sender).To(Equal("me@example.com"))
			Expect(delivery.Recipient).To(Equal("you@example.com"))
			Expect(delivery.Data).To(Equal(strings.Split(msg.Data(), "\n")))
			Expect(delivery.UsedTLS).To(BeTrue())
		})

		It("can make multiple requests", func() {
			firstMsg := mail.Message{
				From:    "me@example.com",
				To:      "you@example.com",
				Subject: "Urgent! Read now!",
				Body: []mail.Part{
					{
						ContentType: "text/plain",
						Content:     "This email is the most important thing you will read all day!",
					},
				},
			}

			err := client.Send(firstMsg)
			if err != nil {
				panic(err)
			}

			Eventually(func() int {
				return len(mailServer.Deliveries)
			}).Should(Equal(1))
			delivery := mailServer.Deliveries[0]

			Expect(delivery.Sender).To(Equal("me@example.com"))
			Expect(delivery.Recipient).To(Equal("you@example.com"))
			Expect(delivery.Data).To(Equal(strings.Split(firstMsg.Data(), "\n")))

			secondMsg := mail.Message{
				From:    "first@example.com",
				To:      "second@example.com",
				Subject: "Boring. Do not read.",
				Body: []mail.Part{
					{
						ContentType: "text/plain",
						Content:     "This email is the least important thing you will read all day. Sorry.",
					},
				},
			}

			err = client.Send(secondMsg)
			if err != nil {
				panic(err)
			}

			Eventually(func() int {
				return len(mailServer.Deliveries)
			}).Should(Equal(2))
			delivery = mailServer.Deliveries[1]

			Expect(delivery.Sender).To(Equal("first@example.com"))
			Expect(delivery.Recipient).To(Equal("second@example.com"))
			Expect(delivery.Data).To(Equal(strings.Split(secondMsg.Data(), "\n")))
		})

		Context("when configured to use TLS", func() {
			BeforeEach(func() {
				config.SkipVerifySSL = true
				client = mail.NewClient(config, logger)
			})

			It("communicates over TLS", func() {
				msg := mail.Message{
					From:    "me@example.com",
					To:      "you@example.com",
					Subject: "Urgent! Read now!",
					Body: []mail.Part{
						{
							ContentType: "text/plain",
							Content:     "This email is the most important thing you will read all day!",
						},
					},
				}

				err := client.Send(msg)
				if err != nil {
					panic(err)
				}

				Eventually(func() int {
					return len(mailServer.Deliveries)
				}).Should(Equal(1))
				delivery := mailServer.Deliveries[0]

				Expect(delivery.Sender).To(Equal("me@example.com"))
				Expect(delivery.Recipient).To(Equal("you@example.com"))
				Expect(delivery.Data).To(Equal(strings.Split(msg.Data(), "\n")))
				Expect(delivery.UsedTLS).To(BeTrue())
			})
		})

		Context("when configured to not use TLS", func() {
			BeforeEach(func() {
				mailServer.SupportsTLS = false
				config.DisableTLS = true
				client = mail.NewClient(config, logger)
			})

			It("does not authenticate", func() {
				msg := mail.Message{
					From:    "me@example.com",
					To:      "you@example.com",
					Subject: "Urgent! Read now!",
					Body: []mail.Part{
						{
							ContentType: "text/plain",
							Content:     "This email is the most important thing you will read all day!",
						},
					},
				}

				err := client.Send(msg)
				if err != nil {
					panic(err)
				}

				Eventually(func() int {
					return len(mailServer.Deliveries)
				}).Should(Equal(1))
				delivery := mailServer.Deliveries[0]

				Expect(delivery.Sender).To(Equal("me@example.com"))
				Expect(delivery.Recipient).To(Equal("you@example.com"))
				Expect(delivery.Data).To(Equal(strings.Split(msg.Data(), "\n")))
				Expect(delivery.UsedTLS).To(BeFalse())
			})
		})

		Context("when a message fails to send", func() {
			It("quits the SMTP connection", func() {
				msg := mail.Message{
					From:    "me@example.com",
					To:      "you@example.com",
					Subject: "Urgent! Read now!",
					Body: []mail.Part{
						{
							ContentType: "text/plain",
							Content:     "This email is the most important thing you will read all day!",
						},
					},
				}
				mailServer.FailsHello = true
				err := client.Send(msg)
				Expect(err).To(HaveOccurred())

				mailServer.FailsHello = false
				err = client.Send(msg)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() int {
					return len(mailServer.Deliveries)
				}).Should(Equal(1))

				delivery := mailServer.Deliveries[0]

				Expect(delivery.Sender).To(Equal("me@example.com"))
				Expect(delivery.Recipient).To(Equal("you@example.com"))
				Expect(delivery.Data).To(Equal(strings.Split(msg.Data(), "\n")))
			})
		})
	})

	Describe("Connect", func() {
		Context("when in test mode", func() {
			It("does not connect to the smtp server", func() {
				var err error

				config.Host, config.Port, err = net.SplitHostPort("fakewebsiteoninternet.com:587")
				if err != nil {
					panic(err)
				}
				config.TestMode = true
				client = mail.NewClient(config, logger)

				err = client.Connect()
				Expect(err).To(BeNil())
			})
		})

		It("returns an error if it cannot connect within the given timeout duration", func() {
			var err error

			mailServer.ConnectWait = 5 * time.Second
			config.ConnectTimeout = 100 * time.Millisecond

			config.Host, config.Port, err = net.SplitHostPort(mailServer.URL.String())
			if err != nil {
				panic(err)
			}

			client = mail.NewClient(config, logger)

			err = client.Connect()
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("server timeout"))
		})
	})

	Describe("Extension", func() {
		BeforeEach(func() {
			var err error

			mailServer.SupportsTLS = true
			config.Host, config.Port, err = net.SplitHostPort(mailServer.URL.String())
			if err != nil {
				panic(err)
			}

			client = mail.NewClient(config, logger)
		})

		It("returns a bool, representing presence of, and parameters for a given SMTP extension", func() {
			err := client.Connect()
			if err != nil {
				panic(err)
			}

			err = client.Hello()
			if err != nil {
				panic(err)
			}

			ok, params := client.Extension("AUTH")
			Expect(ok).To(BeTrue())
			Expect(params).To(Equal("PLAIN LOGIN"))

			ok, params = client.Extension("STARTTLS")
			Expect(ok).To(BeTrue())
			Expect(params).To(Equal(""))
		})
	})

	Describe("AuthMechanism", func() {
		Context("when configured to use PLAIN auth", func() {
			BeforeEach(func() {
				config.AuthMechanism = mail.AuthPlain
				client = mail.NewClient(config, logger)
			})

			It("creates a PlainAuth strategy", func() {
				auth := smtp.PlainAuth("", config.User, config.Pass, config.Host)
				mechanism := client.AuthMechanism()

				Expect(mechanism).To(BeAssignableToTypeOf(auth))
			})
		})

		Context("when configured to use CRAMMD5 auth", func() {
			BeforeEach(func() {
				config.AuthMechanism = mail.AuthCRAMMD5
				client = mail.NewClient(config, logger)
			})

			It("creates a CRAMMD5Auth strategy", func() {
				auth := smtp.CRAMMD5Auth(config.User, config.Secret)
				mechanism := client.AuthMechanism()

				Expect(mechanism).To(BeAssignableToTypeOf(auth))
			})
		})

		Context("when configured to use no auth", func() {
			BeforeEach(func() {
				config.AuthMechanism = mail.AuthNone
				client = mail.NewClient(config, logger)
			})

			It("creates a CRAMMD5Auth strategy", func() {
				mechanism := client.AuthMechanism()

				Expect(mechanism).To(BeNil())
			})
		})
	})

	Describe("Error", func() {
		It("logs the error and returns it", func() {
			err := errors.New("BOOM!")

			otherErr := client.Error(err)

			Expect(otherErr).To(Equal(err))

			Expect(buffer.String()).To(ContainSubstring("SMTP Error: BOOM!"))
		})

		It("quits the current connection when an error occurs", func() {
			Expect(mailServer.ConnectionState).To(Equal(StateUnknown))

			client.Connect()
			Expect(mailServer.ConnectionState).To(Equal(StateConnected))

			client.Error(errors.New("BOOM!!"))
			Expect(mailServer.ConnectionState).To(Equal(StateClosed))
		})
	})

	Describe("PrintLog", func() {
		Context("when the client is configured to log", func() {
			BeforeEach(func() {
				config.LoggingEnabled = true
				client = mail.NewClient(config, logger)
			})

			It("writes to the logger", func() {
				client.PrintLog("banana %s", "panic")

				Expect(buffer.String()).To(ContainSubstring("banana panic"))
			})
		})

		Context("when the client is not configured to log", func() {
			BeforeEach(func() {
				config.LoggingEnabled = false
				client = mail.NewClient(config, logger)
			})

			It("does not write to the logger", func() {
				client.PrintLog("banana %s", "panic")

				Expect(buffer.String()).NotTo(ContainSubstring("banana panic"))
			})
		})
	})
})
