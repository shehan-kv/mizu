package iam

import "testing"

func TestNewName(t *testing.T) {
	tests := []struct {
		name      string
		firstName string
		lastName  string
		wantError error
	}{
		{
			name:      "valid name",
			firstName: "John",
			lastName:  "Doe",
		},
		{
			name:      "empty first name",
			firstName: "",
			lastName:  "Doe",
			wantError: ErrUserFirstNameCannotBeEmpty,
		},
		{
			name:      "empty last name",
			firstName: "John",
			lastName:  "",
			wantError: ErrUserLastNameCannotBeEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := NewName(tt.firstName, tt.lastName)

			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("expected error %v, got %v", tt.wantError, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if name.FirstName() != tt.firstName {
				t.Errorf("expected first name %q, got %q", tt.firstName, name.FirstName())
			}

			if name.LastName() != tt.lastName {
				t.Errorf("expected last name %q, got %q", tt.lastName, name.LastName())
			}
		})
	}
}

func TestNameEquals(t *testing.T) {
	name1, err := NewName("John", "Doe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name2, err := NewName("John", "Doe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name3, err := NewName("Jane", "Doe")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !name1.Equals(name2) {
		t.Error("expected identical names to be equal")
	}

	if name1.Equals(name3) {
		t.Error("expected different names to not be equal")
	}
}
