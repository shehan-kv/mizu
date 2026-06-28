package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"html/template"
	"mizu/internal/application/mailer"
	"mizu/internal/domain/iam"
	"mizu/internal/domain/verification"
	"net"
	"net/smtp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	BaseURL  string
}

type message struct {
	to      string
	subject string
	body    string
	result  chan error
}

// Mailer maintains a persistent TLS SMTP connection and sends
// emails sequentially through an internal queue.
// Call Start to begin processing and Stop for graceful shutdown.
type SMTPMailer struct {
	config    SMTPConfig
	templates *TemplateManager

	// Client is only accessed by the worker goroutine.
	client *smtp.Client

	ch     chan message
	wg     sync.WaitGroup
	closed atomic.Bool
}

func NewSMTPMailer(config SMTPConfig, templates *TemplateManager, buffer int) *SMTPMailer {
	return &SMTPMailer{
		config:    config,
		templates: templates,
		ch:        make(chan message, buffer),
	}
}

// Start begins processing outbound emails in a background goroutine.
// It should be called once before sending emails.
func (m *SMTPMailer) Start(ctx context.Context) {

	m.wg.Go(func() {

		for {
			select {
			case msg, ok := <-m.ch:
				if !ok {
					m.close()
					return
				}

				err := m.send(msg.to, msg.subject, msg.body)

				select {
				case msg.result <- err:
				default:
				}

			case <-ctx.Done():
				m.closed.Store(true)

				// Drain remaining messages.
				for {
					select {
					case msg, ok := <-m.ch:
						if !ok {
							m.close()
							return
						}

						err := m.send(msg.to, msg.subject, msg.body)

						select {
						case msg.result <- err:
						default:
						}

					default:
						m.close()
						return
					}
				}
			}
		}
	})
}

// Stop waits for all in-flight emails to finish sending.
func (m *SMTPMailer) Stop() {
	m.wg.Wait()
}

func (m *SMTPMailer) SendRecoveryEmail(ctx context.Context, e iam.Email, token iam.RecoveryToken) error {
	if m.closed.Load() {
		return errors.New("email.Mailer.SendRecoveryEmail: mailer closed")
	}

	t, err := m.templates.Get(TemplateRecoveryEmail)
	if err != nil {
		return fmt.Errorf("email.Mailer.SendRecoveryEmail: %w", err)
	}

	data := struct {
		BaseURL       string
		RecoveryToken string
	}{
		BaseURL:       m.config.BaseURL,
		RecoveryToken: token.String(),
	}

	var body bytes.Buffer

	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("email.Mailer.SendRecoveryEmail: render template: %w", err)
	}

	result := make(chan error, 1)

	msg := message{
		to:      sanitizeHeader(e.String()),
		subject: sanitizeHeader("Recover Your Account"),
		body:    body.String(),
		result:  result,
	}

	select {
	case m.ch <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *SMTPMailer) SendVerificationEmail(
	ctx context.Context,
	email iam.Email,
	vID verification.VerificationID,
) error {
	if m.closed.Load() {
		return errors.New("email.Mailer.SendVerificationEmail: mailer closed")
	}

	t, err := m.templates.Get(TemplateVerifyEmail)
	if err != nil {
		return fmt.Errorf("email.Mailer.SendVerificationEmail: %w", err)
	}

	data := struct {
		BaseURL        string
		VerificationID string
	}{
		BaseURL:        m.config.BaseURL,
		VerificationID: vID.String(),
	}

	var body bytes.Buffer

	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("email.Mailer.SendVerificationEmail: render template: %w", err)
	}

	result := make(chan error, 1)

	msg := message{
		to:      sanitizeHeader(email.String()),
		subject: sanitizeHeader("Verify Your Account"),
		body:    body.String(),
		result:  result,
	}

	select {
	case m.ch <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *SMTPMailer) SendVerifiedEmail(ctx context.Context, email iam.Email) error {
	if m.closed.Load() {
		return errors.New("email.Mailer.SendVerifiedEmail: mailer closed")
	}

	t, err := m.templates.Get(TemplateVerifiedEmail)
	if err != nil {
		return fmt.Errorf("email.Mailer.SendVerifiedEmail: %w", err)
	}

	var body bytes.Buffer

	if err := t.Execute(&body, nil); err != nil {
		return fmt.Errorf("email.Mailer.SendVerifiedEmail: render template: %w", err)
	}

	result := make(chan error, 1)

	msg := message{
		to:      sanitizeHeader(email.String()),
		subject: sanitizeHeader("Your Account Has Been Successfully Verified"),
		body:    body.String(),
		result:  result,
	}

	select {
	case m.ch <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *SMTPMailer) SendContractEmail(ctx context.Context, email mailer.ContractEmail) error {
	if m.closed.Load() {
		return errors.New("email.Mailer.SendContractEmail: mailer closed")
	}

	t, err := m.templates.Get(TemplateContractEmail)
	if err != nil {
		return fmt.Errorf("email.Mailer.SendContractEmail: %w", err)
	}

	var body bytes.Buffer

	data := struct {
		ContractID    string
		ContractName  string
		ContractTerms string
		ProjectID     string
		ProjectName   string
		Signatories   []mailer.ContractSignatory
	}{
		ContractID:    email.ContractID,
		ContractName:  email.ContractName,
		ContractTerms: email.ContractTerms,
		ProjectID:     email.ProjectID,
		ProjectName:   email.ProjectName,
		Signatories:   email.Signatories,
	}

	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("email.Mailer.SendVerifiedEmail: render template: %w", err)
	}

	result := make(chan error, 1)

	msg := message{
		to:      sanitizeHeader(email.RecipientEmail),
		subject: sanitizeHeader(email.Subject),
		body:    body.String(),
		result:  result,
	}

	select {
	case m.ch <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *SMTPMailer) SendInvoiceEmail(ctx context.Context, email mailer.InvoiceEmail) error {
	if m.closed.Load() {
		return errors.New("email.Mailer.SendInvoiceEmail: mailer closed")
	}

	t, err := m.templates.Get(TemplateInvoiceEmail)
	if err != nil {
		return fmt.Errorf("email.Mailer.SendInvoiceEmail: %w", err)
	}

	var body bytes.Buffer

	var dueAt string
	if email.DueAt != nil {
		dueAt = email.DueAt.UTC().Format("02 Jan 2006 UTC")
	}
	data := struct {
		InvoiceID          string
		FormattedInvoiceID string
		ProjectID          string
		Status             string
		CurrencyName       string
		CurrencyCode       string
		DueAt              string
		Note               *string
		Items              []mailer.InvoiceItem
		SubTotal           string
		TotalTax           string
		TotalDiscount      string
	}{
		InvoiceID:          email.InvoiceID,
		FormattedInvoiceID: strings.ToUpper(email.InvoiceID[len(email.InvoiceID)-8:]),
		ProjectID:          email.ProjectID,

		Status:       email.Status,
		CurrencyName: email.CurrencyName,
		CurrencyCode: email.CurrencyCode,

		DueAt: dueAt,
		Note:  email.Note,

		Items: email.Items,

		SubTotal:      email.SubTotal,
		TotalTax:      email.TotalTax,
		TotalDiscount: email.TotalDiscount,
	}

	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("email.Mailer.SendInvoiceEmail: render template: %w", err)
	}

	result := make(chan error, 1)

	msg := message{
		to:      sanitizeHeader(email.RecipientEmail),
		subject: sanitizeHeader(email.Subject),
		body:    body.String(),
		result:  result,
	}

	select {
	case m.ch <- msg:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *SMTPMailer) send(to, subject, body string) error {
	if err := m.ensureConnected(); err != nil {
		return fmt.Errorf("email.Mailer.send: connect: %w", err)
	}

	if err := m.client.Mail(m.config.From); err != nil {
		// Connection dropped, reconnect and retry once.
		m.client = nil

		if err := m.ensureConnected(); err != nil {
			return fmt.Errorf("email.Mailer.send: reconnect: %w", err)
		}

		if err := m.client.Mail(m.config.From); err != nil {
			return fmt.Errorf("email.Mailer.send: MAIL FROM: %w", err)
		}
	}

	if err := m.client.Rcpt(to); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: RCPT TO: %w", err)
	}

	wc, err := m.client.Data()
	if err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: DATA: %w", err)
	}

	if _, err := fmt.Fprintf(wc, "From: %s\r\n", sanitizeHeader(m.config.From)); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write from: %w", err)
	}

	if _, err := fmt.Fprintf(wc, "To: %s\r\n", to); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write to: %w", err)
	}

	if _, err := fmt.Fprintf(wc, "Subject: %s\r\n", subject); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write subject: %w", err)
	}

	if _, err := fmt.Fprint(wc, "MIME-Version: 1.0\r\n"); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write headers: %w", err)
	}

	if _, err := fmt.Fprintf(wc, "Date: %s\r\n", time.Now().UTC().Format(time.RFC1123Z)); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write headers: %w", err)
	}

	if _, err := fmt.Fprint(wc, "Content-Type: text/html; charset=UTF-8\r\n"); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write headers: %w", err)
	}

	if _, err := fmt.Fprint(wc, "\r\n"); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write headers: %w", err)
	}

	if _, err := fmt.Fprint(wc, body); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: write body: %w", err)
	}

	if err := wc.Close(); err != nil {
		m.client = nil
		return fmt.Errorf("email.Mailer.send: close data writer: %w", err)
	}

	return nil
}

func (m *SMTPMailer) ensureConnected() error {
	if m.client != nil {
		if err := m.client.Noop(); err == nil {
			return nil
		}

		m.client.Close()
		m.client = nil
	}

	addr := net.JoinHostPort(m.config.Host, fmt.Sprint(m.config.Port))

	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	tlsConfig := &tls.Config{
		ServerName: m.config.Host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("tls dial %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, m.config.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp client: %w", err)
	}

	auth := smtp.PlainAuth(
		"",
		m.config.Username,
		m.config.Password,
		m.config.Host,
	)

	if err := client.Auth(auth); err != nil {
		client.Close()
		return fmt.Errorf("auth: %w", err)
	}

	m.client = client

	return nil
}

// Close sends QUIT and closes the connection cleanly.
func (m *SMTPMailer) close() {
	if m.client != nil {
		m.client.Quit()
		m.client = nil
	}
}

func sanitizeHeader(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return template.HTMLEscapeString(s)
}
