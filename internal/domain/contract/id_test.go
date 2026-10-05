package contract

import "testing"

func TestNewContractID(t *testing.T) {
	t.Run("creates contract ID", func(t *testing.T) {
		id, err := NewContractID("contract-123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if id.String() != "contract-123" {
			t.Fatalf("expected %q, got %q", "contract-123", id.String())
		}
	})

	t.Run("rejects empty ID", func(t *testing.T) {
		_, err := NewContractID("")
		if err != ErrContractIDCannotBeEmpty {
			t.Fatalf("expected %v, got %v", ErrContractIDCannotBeEmpty, err)
		}
	})
}
