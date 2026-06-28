package learning

import "testing"

func TestTargetLangNameRUAccusative_AllCodes(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"en", "английский"},
		{"es", "испанский"},
		{"de", "немецкий"},
		{"fr", "французский"},
		{"it", "итальянский"},
		{"pt", "португальский"},
		{"", "английский"},
		{"xx", "xx"},
	}
	for _, tc := range tests {
		if got := TargetLangNameRUAccusative(tc.code); got != tc.want {
			t.Fatalf("TargetLangNameRUAccusative(%q) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestTargetLangNameRUPrepositional_AllCodes(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"en", "английском"},
		{"es", "испанском"},
		{"de", "немецком"},
		{"fr", "французском"},
		{"it", "итальянском"},
		{"pt", "португальском"},
		{"", "английском"},
		{"XX", "xx"},
	}
	for _, tc := range tests {
		if got := TargetLangNameRUPrepositional(tc.code); got != tc.want {
			t.Fatalf("TargetLangNameRUPrepositional(%q) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestTargetLangNameEN_AllCodes(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"en", "English"},
		{"es", "Spanish"},
		{"de", "German"},
		{"fr", "French"},
		{"it", "Italian"},
		{"pt", "Portuguese"},
		{"", "English"},
		{"XX", "xx"},
	}
	for _, tc := range tests {
		if got := TargetLangNameEN(tc.code); got != tc.want {
			t.Fatalf("TargetLangNameEN(%q) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestTargetLangNameES_AllCodes(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"en", "inglés"},
		{"es", "español"},
		{"de", "alemán"},
		{"fr", "francés"},
		{"it", "italiano"},
		{"pt", "portugués"},
		{"", "inglés"},
		{"XX", "xx"},
	}
	for _, tc := range tests {
		if got := TargetLangNameES(tc.code); got != tc.want {
			t.Fatalf("TargetLangNameES(%q) = %q, want %q", tc.code, got, tc.want)
		}
	}
}
