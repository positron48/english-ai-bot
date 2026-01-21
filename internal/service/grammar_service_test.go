package service

import "testing"

func TestNormalizeTrueFalseValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantVal  string
		wantOK   bool
	}{
		{"string true", "true", "true", true},
		{"string True", "True", "true", true},
		{"string FALSE", "FALSE", "false", true},
		{"string Да", "Да", "true", true},
		{"string да", "да", "true", true},
		{"string Нет", "Нет", "false", true},
		{"string нет", "нет", "false", true},
		{"string yes", "yes", "true", true},
		{"string no", "no", "false", true},
		{"bool true", true, "true", true},
		{"bool false", false, "false", true},
		{"unknown string", "x", "", false},
		{"empty string", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeTrueFalseValue(tt.input)
			if ok != tt.wantOK || got != tt.wantVal {
				t.Errorf("normalizeTrueFalseValue(%v) = (%q, %v), want (%q, %v)", tt.input, got, ok, tt.wantVal, tt.wantOK)
			}
		})
	}
}
