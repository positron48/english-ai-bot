package service

import "testing"

func TestNormalizeTargetVerbDisplay(t *testing.T) {
	t.Run("English keeps to prefix for verbs", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("en", "verb", "to make")
		if got != "to make" {
			t.Fatalf("got %q, want to make", got)
		}
	})

	t.Run("Spanish verbo strips to prefix", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("es", "verbo", "to hablar")
		if got != "hablar" {
			t.Fatalf("got %q, want hablar", got)
		}
	})

	t.Run("Spanish strips legacy to prefix without POS", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("es", "", "to comer")
		if got != "comer" {
			t.Fatalf("got %q, want comer", got)
		}
	})

	t.Run("Spanish noun keeps word without to", func(t *testing.T) {
		got := normalizeTargetVerbDisplay("es", "noun", "casa")
		if got != "casa" {
			t.Fatalf("got %q, want casa", got)
		}
	})
}

func TestTargetLangForCourse(t *testing.T) {
	tests := []struct {
		course   string
		fallback string
		want     string
	}{
		{course: "es_ru", fallback: "en", want: "es"},
		{course: "en_ru", fallback: "es", want: "en"},
		{course: "", fallback: "en", want: "en"},
		{course: "garbage", fallback: "en", want: "en"},
	}
	for _, tt := range tests {
		t.Run(tt.course, func(t *testing.T) {
			if got := TargetLangForCourse(tt.course, tt.fallback); got != tt.want {
				t.Fatalf("TargetLangForCourse(%q, %q) = %q, want %q", tt.course, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestNativeLangForCourse(t *testing.T) {
	if got := NativeLangForCourse("es_ru", "en"); got != "ru" {
		t.Fatalf("got %q, want ru", got)
	}
	if got := NativeLangForCourse("", "en"); got != "en" {
		t.Fatalf("got %q, want en", got)
	}
}
