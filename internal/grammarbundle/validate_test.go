package grammarbundle

import "testing"

func TestValidateEmbeddedBundleID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{name: "en", id: "en", wantErr: false},
		{name: "es", id: "es", wantErr: false},
		{name: "EN case", id: "EN", wantErr: false},
		{name: "empty", id: "", wantErr: true},
		{name: "whitespace", id: "   ", wantErr: true},
		{name: "missing fr", id: "fr", wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateEmbeddedBundleID(tt.id)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
