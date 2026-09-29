package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"mizu/internal/application/mailer"
	"mizu/internal/domain/iam"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testMailpit struct {
	container testcontainers.Container
	config    SMTPConfig
	httpPort  int
}

var sharedMailpit testMailpit

func TestMain(m *testing.M) {
	sharedMailpit = newTestMailpit()

	code := m.Run()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := sharedMailpit.container.Terminate(ctx); err != nil {
		fmt.Fprintf(
			os.Stderr,
			"failed to terminate Mailpit container: %v\n",
			err,
		)
	}

	os.Exit(code)
}

func newTestMailpit() testMailpit {
	ctx := context.Background()

	certPEM, keyPEM := newTestTLSCertificate()

	container, err := testcontainers.Run(
		ctx,
		"axllent/mailpit:v1.27.8",
		testcontainers.WithExposedPorts(
			"1025/tcp",
			"8025/tcp",
		),
		testcontainers.WithFiles(
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(certPEM),
				ContainerFilePath: "/tmp/mailpit-cert.pem",
				FileMode:          0644,
			},
			testcontainers.ContainerFile{
				Reader:            bytes.NewReader(keyPEM),
				ContainerFilePath: "/tmp/mailpit-key.pem",
				FileMode:          0600,
			},
		),
		testcontainers.WithCmd(
			"--smtp", "0.0.0.0:1025",
			"--listen", "0.0.0.0:8025",
			"--smtp-tls-cert", "/tmp/mailpit-cert.pem",
			"--smtp-tls-key", "/tmp/mailpit-key.pem",
			"--smtp-require-tls",
		),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("1025/tcp").
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to start Mailpit container: %w",
			err,
		))
	}

	smtpPort, err := container.MappedPort(
		ctx,
		"1025/tcp",
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to get Mailpit SMTP port: %w",
			err,
		))
	}

	smtpPortNumber, err := strconv.Atoi(
		smtpPort.Port(),
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse Mailpit SMTP port: %w",
			err,
		))
	}

	httpPort, err := container.MappedPort(
		ctx,
		"8025/tcp",
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to get Mailpit HTTP port: %w",
			err,
		))
	}

	httpPortNumber, err := strconv.Atoi(
		httpPort.Port(),
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to parse Mailpit HTTP port: %w",
			err,
		))
	}

	rootCAs := x509.NewCertPool()

	if !rootCAs.AppendCertsFromPEM(certPEM) {
		panic("failed to add Mailpit certificate to root CA pool")
	}

	return testMailpit{
		container: container,
		httpPort:  httpPortNumber,
		config: SMTPConfig{
			Host:     "localhost",
			Port:     smtpPortNumber,
			Username: "",
			Password: "",
			From:     "noreply@example.com",
			BaseURL:  "https://example.com",
			TLSConfig: &tls.Config{
				ServerName: "localhost",
				RootCAs:    rootCAs,
				MinVersion: tls.VersionTLS12,
			},
		},
	}
}

func newTestTLSCertificate() ([]byte, []byte) {
	privateKey, err := rsa.GenerateKey(
		rand.Reader,
		2048,
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to generate TLS private key: %w",
			err,
		))
	}

	serialNumber, err := rand.Int(
		rand.Reader,
		new(big.Int).Lsh(
			big.NewInt(1),
			128,
		),
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to generate certificate serial number: %w",
			err,
		))
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		DNSNames: []string{
			"localhost",
		},
		NotBefore: time.Now().Add(-time.Minute),
		NotAfter:  time.Now().Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}

	certificateDER, err := x509.CreateCertificate(
		rand.Reader,
		template,
		template,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		panic(fmt.Errorf(
			"failed to create TLS certificate: %w",
			err,
		))
	}

	certPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certificateDER,
		},
	)

	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)

	return certPEM, keyPEM
}

func clearMailpit(t *testing.T) {
	t.Helper()

	url := fmt.Sprintf(
		"http://localhost:%d/api/v1/messages",
		sharedMailpit.httpPort,
	)

	req, err := http.NewRequest(
		http.MethodDelete,
		url,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"failed to create Mailpit delete request: %v",
			err,
		)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf(
			"failed to clear Mailpit messages: %v",
			err,
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent &&
		resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		t.Fatalf(
			"failed to clear Mailpit messages: status=%d body=%s",
			resp.StatusCode,
			body,
		)
	}
}

func TestSMTPMailer_SendVerifiedEmail(t *testing.T) {
	t.Run("sends verified email", func(t *testing.T) {
		clearMailpit(t)

		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			sharedMailpit.config,
			templates,
			10,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)

		m.Start(ctx)

		defer func() {
			cancel()
			m.Stop()
		}()

		email, err := iam.NewEmail(
			"user@example.com",
		)
		if err != nil {
			t.Fatal(err)
		}

		err = m.SendVerifiedEmail(
			ctx,
			email,
		)
		if err != nil {
			t.Fatalf(
				"SendVerifiedEmail() error = %v",
				err,
			)
		}

		message := waitForLatestMessage(
			t,
			sharedMailpit,
		)

		if message.Subject !=
			"Your Account Has Been Successfully Verified" {
			t.Errorf(
				"subject = %q, want %q",
				message.Subject,
				"Your Account Has Been Successfully Verified",
			)
		}

		if len(message.To) != 1 {
			t.Fatalf(
				"recipient count = %d, want 1",
				len(message.To),
			)
		}

		if message.To[0].Address != "user@example.com" {
			t.Errorf(
				"recipient = %q, want %q",
				message.To[0].Address,
				"user@example.com",
			)
		}

		body := getLatestHTML(
			t,
			sharedMailpit,
		)

		if !strings.Contains(
			body,
			"Email verified successfully",
		) {
			t.Error(
				"email body does not contain expected verified message",
			)
		}
	})
}

func TestSMTPMailer_SendVerificationEmail(t *testing.T) {
	t.Run("sends verification email", func(t *testing.T) {
		clearMailpit(t)

		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			sharedMailpit.config,
			templates,
			10,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)

		m.Start(ctx)

		defer func() {
			cancel()
			m.Stop()
		}()

		email, err := iam.NewEmail(
			"user@example.com",
		)
		if err != nil {
			t.Fatal(err)
		}

		verificationID, err := iam.NewVerificationID(
			"verification-token-123",
		)
		if err != nil {
			t.Fatal(err)
		}

		err = m.SendVerificationEmail(
			ctx,
			email,
			verificationID,
		)
		if err != nil {
			t.Fatalf(
				"SendVerificationEmail() error = %v",
				err,
			)
		}

		message := waitForLatestMessage(
			t,
			sharedMailpit,
		)

		if message.Subject != "Verify Your Account" {
			t.Errorf(
				"subject = %q, want %q",
				message.Subject,
				"Verify Your Account",
			)
		}

		if len(message.To) != 1 {
			t.Fatalf(
				"recipient count = %d, want 1",
				len(message.To),
			)
		}

		if message.To[0].Address != "user@example.com" {
			t.Errorf(
				"recipient = %q, want %q",
				message.To[0].Address,
				"user@example.com",
			)
		}

		body := getLatestHTML(
			t,
			sharedMailpit,
		)

		for _, want := range []string{
			"verification-token-123",
			"https://example.com",
		} {
			if !strings.Contains(body, want) {
				t.Errorf(
					"email body does not contain %q",
					want,
				)
			}
		}
	})
}

func TestSMTPMailer_SendRecoveryEmail(t *testing.T) {
	t.Run("sends recovery email", func(t *testing.T) {
		clearMailpit(t)

		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			sharedMailpit.config,
			templates,
			10,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)

		m.Start(ctx)

		defer func() {
			cancel()
			m.Stop()
		}()

		email, err := iam.NewEmail(
			"user@example.com",
		)
		if err != nil {
			t.Fatal(err)
		}

		token, err := iam.NewRecoveryToken(
			"recovery-token-123",
		)
		if err != nil {
			t.Fatal(err)
		}

		err = m.SendRecoveryEmail(
			ctx,
			email,
			token,
		)
		if err != nil {
			t.Fatalf(
				"SendRecoveryEmail() error = %v",
				err,
			)
		}

		message := waitForLatestMessage(
			t,
			sharedMailpit,
		)

		if message.Subject != "Recover Your Account" {
			t.Errorf(
				"subject = %q, want %q",
				message.Subject,
				"Recover Your Account",
			)
		}

		if len(message.To) != 1 {
			t.Fatalf(
				"recipient count = %d, want 1",
				len(message.To),
			)
		}

		if message.To[0].Address != "user@example.com" {
			t.Errorf(
				"recipient = %q, want %q",
				message.To[0].Address,
				"user@example.com",
			)
		}

		body := getLatestHTML(
			t,
			sharedMailpit,
		)

		for _, want := range []string{
			"recovery-token-123",
			"https://example.com",
		} {
			if !strings.Contains(body, want) {
				t.Errorf(
					"email body does not contain %q",
					want,
				)
			}
		}
	})
}

func TestSMTPMailer_SendContractEmail(t *testing.T) {
	t.Run("sends contract email", func(t *testing.T) {
		clearMailpit(t)

		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			sharedMailpit.config,
			templates,
			10,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)

		m.Start(ctx)

		defer func() {
			cancel()
			m.Stop()
		}()

		now := time.Now().UTC()

		email := mailer.ContractEmail{
			RecipientEmail: "client@example.com",
			Subject:        "Contract Ready for Signature",
			ContractID:     "contract-123",
			ContractName:   "Website Development Contract",
			ContractTerms:  "Payment is due within 30 days.",
			ProjectID:      "project-123",
			ProjectName:    "Website Project",
			Signatories: []mailer.ContractSignatory{
				{
					ID:        "signatory-1",
					FirstName: "Client",
					LastName:  "User",
					Email:     "client@example.com",
					Role:      "client",
					Status:    "signed",
					UpdatedAt: now,
				},
				{
					ID:        "signatory-2",
					FirstName: "Freelancer",
					LastName:  "User",
					Email:     "freelancer@example.com",
					Role:      "freelancer",
					Status:    "signed",
					UpdatedAt: now,
				},
			},
		}

		err = m.SendContractEmail(
			ctx,
			email,
		)
		if err != nil {
			t.Fatalf(
				"SendContractEmail() error = %v",
				err,
			)
		}

		message := waitForLatestMessage(
			t,
			sharedMailpit,
		)

		if message.Subject != email.Subject {
			t.Errorf(
				"subject = %q, want %q",
				message.Subject,
				email.Subject,
			)
		}

		if len(message.To) != 1 {
			t.Fatalf(
				"recipient count = %d, want 1",
				len(message.To),
			)
		}

		if message.To[0].Address != email.RecipientEmail {
			t.Errorf(
				"recipient = %q, want %q",
				message.To[0].Address,
				email.RecipientEmail,
			)
		}

		body := getLatestHTML(
			t,
			sharedMailpit,
		)

		for _, want := range []string{
			email.ContractName,
			email.ContractTerms,
			email.ProjectName,
			"Client",
			"User",
			"Freelancer",
		} {
			if !strings.Contains(body, want) {
				t.Errorf(
					"email body does not contain %q",
					want,
				)
			}
		}
	})
}

func TestSMTPMailer_SendInvoiceEmail(t *testing.T) {
	t.Run("sends invoice email", func(t *testing.T) {
		clearMailpit(t)

		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			sharedMailpit.config,
			templates,
			10,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)

		m.Start(ctx)

		defer func() {
			cancel()
			m.Stop()
		}()

		dueAt := time.Date(
			2026,
			time.October,
			15,
			0,
			0,
			0,
			0,
			time.UTC,
		)

		email := mailer.InvoiceEmail{
			RecipientEmail: "client@example.com",
			Subject:        "Invoice INV-12345678",
			InvoiceID:      "invoice-12345678",
			ProjectID:      "project-123",
			Status:         "unpaid",
			CurrencyName:   "US Dollar",
			CurrencyCode:   "USD",
			DueAt:          &dueAt,
			SubTotal:       "1000.00",
			TotalTax:       "100.00",
			TotalDiscount:  "50.00",
		}

		err = m.SendInvoiceEmail(
			ctx,
			email,
		)
		if err != nil {
			t.Fatalf(
				"SendInvoiceEmail() error = %v",
				err,
			)
		}

		message := waitForLatestMessage(
			t,
			sharedMailpit,
		)

		if message.Subject != email.Subject {
			t.Errorf(
				"subject = %q, want %q",
				message.Subject,
				email.Subject,
			)
		}

		if len(message.To) != 1 {
			t.Fatalf(
				"recipient count = %d, want 1",
				len(message.To),
			)
		}

		if message.To[0].Address != email.RecipientEmail {
			t.Errorf(
				"recipient = %q, want %q",
				message.To[0].Address,
				email.RecipientEmail,
			)
		}

		body := getLatestHTML(
			t,
			sharedMailpit,
		)

		for _, want := range []string{
			"#12345678 - Unpaid",
			"15 Oct 2026 UTC",
			"USD",
			"US Dollar",
			"1000.00",
			"100.00",
			"50.00",
		} {
			if !strings.Contains(body, want) {
				t.Errorf(
					"email body does not contain %q",
					want,
				)
			}
		}
	})
}

func TestSMTPMailer_Send(t *testing.T) {
	t.Run("sends multiple emails sequentially", func(t *testing.T) {
		clearMailpit(t)

		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			sharedMailpit.config,
			templates,
			10,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)

		m.Start(ctx)

		defer func() {
			cancel()
			m.Stop()
		}()

		first, err := iam.NewEmail(
			"first@example.com",
		)
		if err != nil {
			t.Fatal(err)
		}

		second, err := iam.NewEmail(
			"second@example.com",
		)
		if err != nil {
			t.Fatal(err)
		}

		if err := m.SendVerifiedEmail(
			ctx,
			first,
		); err != nil {
			t.Fatalf(
				"SendVerifiedEmail(first) error = %v",
				err,
			)
		}

		if err := m.SendVerifiedEmail(
			ctx,
			second,
		); err != nil {
			t.Fatalf(
				"SendVerifiedEmail(second) error = %v",
				err,
			)
		}

		messages := waitForMessages(
			t,
			sharedMailpit,
			2,
		)

		if len(messages) != 2 {
			t.Fatalf(
				"message count = %d, want 2",
				len(messages),
			)
		}

		recipients := make(map[string]bool)

		for _, message := range messages {
			if len(message.To) != 1 {
				t.Fatalf(
					"recipient count = %d, want 1",
					len(message.To),
				)
			}

			recipients[message.To[0].Address] = true
		}

		if !recipients["first@example.com"] {
			t.Error("first email was not received")
		}

		if !recipients["second@example.com"] {
			t.Error("second email was not received")
		}
	})
}

func TestSMTPMailer_Closed(t *testing.T) {
	t.Run("returns error when mailer is closed", func(t *testing.T) {
		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			SMTPConfig{},
			templates,
			1,
		)

		m.closed.Store(true)

		email, err := iam.NewEmail(
			"user@example.com",
		)
		if err != nil {
			t.Fatal(err)
		}

		verificationID, err := iam.NewVerificationID(
			"verification-token",
		)
		if err != nil {
			t.Fatal(err)
		}

		err = m.SendVerificationEmail(
			context.Background(),
			email,
			verificationID,
		)

		if err == nil {
			t.Fatal(
				"SendVerificationEmail() error = nil, want error",
			)
		}

		if !strings.Contains(
			err.Error(),
			"mailer closed",
		) {
			t.Errorf(
				"error = %q, want mailer closed error",
				err,
			)
		}
	})
}

func TestSMTPMailer_ContextCancellation(t *testing.T) {
	t.Run("returns context cancellation when queueing is blocked", func(t *testing.T) {
		templates, err := NewTemplateManager("")
		if err != nil {
			t.Fatal(err)
		}

		m := NewSMTPMailer(
			SMTPConfig{},
			templates,
			0,
		)

		ctx, cancel := context.WithCancel(
			context.Background(),
		)
		cancel()

		email, err := iam.NewEmail(
			"user@example.com",
		)
		if err != nil {
			t.Fatal(err)
		}

		verificationID, err := iam.NewVerificationID(
			"verification-token",
		)
		if err != nil {
			t.Fatal(err)
		}

		err = m.SendVerificationEmail(
			ctx,
			email,
			verificationID,
		)

		if !errors.Is(
			err,
			context.Canceled,
		) {
			t.Errorf(
				"error = %v, want %v",
				err,
				context.Canceled,
			)
		}
	})
}

type testMailpitMessage struct {
	ID      string `json:"ID"`
	Subject string `json:"Subject"`

	To []struct {
		Name    string `json:"Name"`
		Address string `json:"Address"`
	} `json:"To"`
}

type testMailpitMessagesResponse struct {
	Messages []testMailpitMessage `json:"messages"`
	Total    int                  `json:"total"`
}

func waitForMessages(
	t *testing.T,
	mailpit testMailpit,
	count int,
) []testMailpitMessage {
	t.Helper()

	url := fmt.Sprintf(
		"http://localhost:%d/api/v1/messages",
		mailpit.httpPort,
	)

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()

			if readErr == nil &&
				resp.StatusCode == http.StatusOK {
				var result testMailpitMessagesResponse

				if err := json.Unmarshal(
					body,
					&result,
				); err == nil {
					if len(result.Messages) >= count {
						return result.Messages
					}
				}
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf(
		"timed out waiting for %d Mailpit messages",
		count,
	)

	return nil
}

func waitForLatestMessage(
	t *testing.T,
	mailpit testMailpit,
) testMailpitMessage {
	t.Helper()

	messages := waitForMessages(
		t,
		mailpit,
		1,
	)

	return messages[0]
}

func getLatestHTML(
	t *testing.T,
	mailpit testMailpit,
) string {
	t.Helper()

	url := fmt.Sprintf(
		"http://localhost:%d/view/latest.html",
		mailpit.httpPort,
	)

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()

			if readErr == nil &&
				resp.StatusCode == http.StatusOK {
				return string(body)
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal(
		"timed out waiting for Mailpit HTML message",
	)

	return ""
}
