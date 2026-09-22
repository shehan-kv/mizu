package billing

import "testing"

func TestNewInvoiceID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    InvoiceID
		wantErr error
	}{
		{
			name:  "valid ID",
			input: "invoice-1",
			want:  InvoiceID("invoice-1"),
		},
		{
			name:    "empty ID",
			input:   "",
			wantErr: ErrBillingInvoiceIDCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NewInvoiceID(tt.input)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if id != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, id)
			}

			if id.String() != tt.input {
				t.Fatalf("expected String() %q, got %q", tt.input, id.String())
			}
		})
	}
}
