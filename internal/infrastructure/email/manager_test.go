package email

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateManager(t *testing.T) {
	t.Run("loads embedded templates", func(t *testing.T) {
		manager, err := NewTemplateManager("")
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}

		names := []string{
			TemplateVerifyEmail,
			TemplateVerifiedEmail,
			TemplateContractEmail,
			TemplateInvoiceEmail,
			TemplateRecoveryEmail,
		}

		for _, name := range names {
			t.Run(name, func(t *testing.T) {
				template, err := manager.Get(name)
				if err != nil {
					t.Fatalf("failed to get template: %v", err)
				}

				if template == nil {
					t.Fatal("expected template, got nil")
				}
			})
		}
	})

	t.Run("returns error for unknown template", func(t *testing.T) {
		manager, err := NewTemplateManager("")
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}

		_, err = manager.Get("does-not-exist")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("loads custom template", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(
			filepath.Join(dir, "custom.html"),
			[]byte(`<html>{{.Name}}</html>`),
			0644,
		)
		if err != nil {
			t.Fatalf("failed to write custom template: %v", err)
		}

		manager, err := NewTemplateManager(dir)
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}

		template, err := manager.Get("custom")
		if err != nil {
			t.Fatalf("failed to get custom template: %v", err)
		}

		var buf bytes.Buffer

		if err := template.Execute(&buf, map[string]string{
			"Name": "Shehan",
		}); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}

		if got := buf.String(); got != `<html>Shehan</html>` {
			t.Fatalf("expected rendered template, got %q", got)
		}
	})

	t.Run("custom template overrides embedded template", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(
			filepath.Join(dir, TemplateVerifyEmail+".html"),
			[]byte(`custom verification email`),
			0644,
		)
		if err != nil {
			t.Fatalf("failed to write custom template: %v", err)
		}

		manager, err := NewTemplateManager(dir)
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}

		template, err := manager.Get(TemplateVerifyEmail)
		if err != nil {
			t.Fatalf("failed to get overridden template: %v", err)
		}

		var buf bytes.Buffer

		if err := template.Execute(&buf, nil); err != nil {
			t.Fatalf("failed to execute template: %v", err)
		}

		if got := buf.String(); got != "custom verification email" {
			t.Fatalf("expected custom template, got %q", got)
		}
	})

	t.Run("ignores non-html files", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(
			filepath.Join(dir, "ignored.txt"),
			[]byte("not a template"),
			0644,
		)
		if err != nil {
			t.Fatalf("failed to write file: %v", err)
		}

		manager, err := NewTemplateManager(dir)
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}

		_, err = manager.Get("ignored")
		if err == nil {
			t.Fatal("expected template to be ignored")
		}
	})

	t.Run("returns error for invalid custom template", func(t *testing.T) {
		dir := t.TempDir()

		err := os.WriteFile(
			filepath.Join(dir, "invalid.html"),
			[]byte(`{{if}}`),
			0644,
		)
		if err != nil {
			t.Fatalf("failed to write template: %v", err)
		}

		_, err = NewTemplateManager(dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error for missing custom directory", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "does-not-exist")

		_, err := NewTemplateManager(dir)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
