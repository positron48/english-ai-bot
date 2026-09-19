package ai

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"testing"
)

func TestSentenceNormalizationPreservesLexicalDifferences(t *testing.T) {
	for _, pair := range [][2]string{{"Sí, yo bebo agua.", "sí yo bebo agua"}, {"No,I don't.", "No I don’t"}} {
		if NormalizedSentenceAnswer(pair[0]) != NormalizedSentenceAnswer(pair[1]) {
			t.Errorf("cosmetic difference: %v", pair)
		}
	}
	for _, pair := range [][2]string{{"té", "te"}, {"él", "el"}, {"we're", "were"}, {"re-sign", "resign"}, {"bebe", "bebé"}} {
		if NormalizedSentenceAnswer(pair[0]) == NormalizedSentenceAnswer(pair[1]) {
			t.Errorf("lost lexical distinction: %v", pair)
		}
	}
}

func TestSentenceGradeConsistency(t *testing.T) {
	for _, tc := range []struct {
		name, input, raw string
		wantErrors       int
		invalid          bool
	}{
		{"punctuation only", "Sí bebo agua", `{"error_count":1,"corrected_es":"Sí, bebo agua.","issues":[{"kind":"punctuation","original":"Sí","corrected":"Sí,","explanation":"Запятая."}],"explanation":"Нужна запятая."}`, 0, false},
		{"optional advice", "Bebo agua", `{"error_count":1,"corrected_es":"Bebo agua","issues":[],"explanation":"Можно добавить yo."}`, 0, false},
		{"real correction", "Nosotros bebe agua", `{"error_count":99,"outcome":"star","corrected_es":"Nosotros bebemos agua","issues":[{"kind":"verb_form","original":"bebe","corrected":"bebemos","explanation":"Для nosotros нужна форма bebemos."}],"explanation":"Для nosotros нужна форма bebemos."}`, 1, false},
		{"invisible correction", "Nosotros bebe agua", `{"error_count":1,"corrected_es":"Nosotros bebe agua","issues":[{"kind":"verb_form","original":"bebe","corrected":"bebemos","explanation":"Нужна форма bebemos."}],"explanation":"Нужна форма bebemos."}`, 0, true},
		{"missing issues", "Bebo agua", `{"error_count":0,"corrected_es":"Bebo agua"}`, 0, true},
		{"unexplained rewrite", "Bebo agua", `{"corrected_es":"Yo bebo agua","issues":[]}`, 0, true},
		{"unanchored issue", "Bebo agua", `{"corrected_es":"Bebo agua","issues":[{"kind":"meaning","explanation":"Что-то не так."}]}`, 0, true},
		{"no corrected answer", "Bebo agua", `{"corrected_es":"","issues":[]}`, 0, true},
		{"wrong explanation language", "Bebe agua", `{"corrected_es":"Bebo agua","issues":[{"kind":"verb_form","original":"Bebe","corrected":"Bebo","explanation":"Use first person."}],"explanation":"Use first person."}`, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var grade SentenceGrade
			if err := json.Unmarshal([]byte(tc.raw), &grade); err != nil {
				t.Fatal(err)
			}
			err := normalizeSentenceGrade(&grade, "Я пью воду.", tc.input)
			if (err != nil) != tc.invalid {
				t.Fatalf("err=%v, invalid=%v", err, tc.invalid)
			}
			if tc.invalid {
				return
			}
			if grade.ErrorCount != tc.wantErrors {
				t.Fatalf("grade=%+v", grade)
			}
			if tc.wantErrors == 0 && (grade.Outcome != "star" || grade.Explanation != "" || grade.CorrectedES != tc.input) {
				t.Fatalf("false penalty: %+v", grade)
			}
			if tc.wantErrors == 1 && grade.Outcome != "passed" {
				t.Fatalf("wrong outcome: %+v", grade)
			}
		})
	}
}

func TestSentenceGradeRetriesInconsistentOutput(t *testing.T) {
	for _, repair := range []bool{true, false} {
		t.Run(map[bool]string{true: "recovered", false: "not recorded"}[repair], func(t *testing.T) {
			svc := NewService("https://example.test", "general-model", "test", "", zap.NewNop())
			svc.SetSentenceModel("sentence-model")
			svc.SetSentenceGradePromptForCourse("es_ru", "Grade translation")
			calls := 0
			svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				var body ChatRequest
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Model != "sentence-model" {
					t.Fatalf("wrong model %q", body.Model)
				}
				content := `{"error_count":1,"corrected_es":"Yo bebes agua","issues":[{"kind":"verb_form"}],"explanation":"Нужна другая форма."}`
				if calls == 2 && repair {
					content = `{"error_count":1,"corrected_es":"Yo bebo agua","issues":[{"kind":"verb_form","original":"bebes","corrected":"bebo","explanation":"Для yo нужна форма bebo."}],"explanation":"Для yo нужна форма bebo."}`
				}
				return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: content}}}}), nil
			})
			g, err := svc.GradeSentenceForCourse(context.Background(), "es_ru", "Я пью воду.", "", "Bebo agua.", "Yo bebes agua")
			if (err == nil) != repair || calls != 2 {
				t.Fatalf("repair=%v calls=%d grade=%+v err=%v", repair, calls, g, err)
			}
		})
	}
}

func TestSentenceExactAnswerDoesNotCallModel(t *testing.T) {
	svc := NewService("https://example.test", "general", "test", "", zap.NewNop())
	svc.SetSentenceGradePromptForCourse("es_ru", "Grade translation")
	svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) { t.Fatal("unexpected paid call"); return nil, nil })
	g, err := svc.GradeSentenceForCourse(context.Background(), "es_ru", "Да, я пью воду.", "", "Sí, bebo agua.", "sí bebo agua")
	if err != nil || g.Outcome != "star" {
		t.Fatalf("grade=%+v err=%v", g, err)
	}
}

func TestSentenceModelRouting(t *testing.T) {
	svc := NewService("https://example.test", "general", "test", "", zap.NewNop())
	if svc.sentenceModelOr() != "general" {
		t.Fatal("missing default fallback")
	}
	svc.SetSentenceModel(" sentence ")
	if svc.sentenceModelOr() != "sentence" || svc.sentenceModelOr("override") != "override" || svc.modelOr() != "general" {
		t.Fatal("sentence override escaped its scope")
	}
}

func TestRussianAddressNeedsContext(t *testing.T) {
	for _, input := range []string{"Вы пьёте воду.", "Я помогаю вам.", "Он читает вашу книгу."} {
		if !hasRussianAddress(input) {
			t.Fatalf("missing address: %s", input)
		}
	}
	if hasRussianAddress("Он пьёт воду и выращивает овощи.") {
		t.Fatal("substring matched an address")
	}
}

func TestSentenceQualityReviewFailsClosed(t *testing.T) {
	svc := NewService("https://example.test", "general", "test", "", zap.NewNop())
	svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		var body ChatRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		if !strings.Contains(body.Messages[1].Content, "allowed_tenses") || !strings.Contains(body.Messages[0].Content, "Vosotros") {
			t.Fatal("review lacks grammar/context constraints")
		}
		return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: "not JSON"}}}}), nil
	})
	out, err := svc.reviewGeneratedSentenceQuality(context.Background(), "sentence", "es_ru", []GenSentenceWord{{"beber", "пить"}}, nil, []string{"presente"}, []GeneratedSentence{{PromptRU: "Вы пьёте воду.", ReferenceES: "Bebéis agua."}}, nil)
	if err == nil || len(out.Accepted) != 0 {
		t.Fatal("unreviewed ambiguous exercise escaped")
	}
}

func TestSentenceIssuesCoverEveryCorrection(t *testing.T) {
	for _, tc := range []struct {
		name, input, corrected string
		issues                 []SentenceGradeIssue
		ok                     bool
	}{
		{"unreported deletion", "Nosotros bebe agua", "Bebemos agua", []SentenceGradeIssue{{Original: "bebe", Corrected: "bebemos"}}, false},
		{"duplicate error", "Yo bebes agua", "Yo bebo agua", []SentenceGradeIssue{{Original: "bebes", Corrected: "bebo"}, {Original: "bebes", Corrected: "bebo"}}, false},
		{"unreported insertion", "Yo bebes agua", "Yo no bebo agua", []SentenceGradeIssue{{Original: "bebes", Corrected: "bebo"}}, false},
		{"grouped agreement", "Vosotros bebéis agua", "Usted bebe agua", []SentenceGradeIssue{{Original: "Vosotros bebéis", Corrected: "Usted bebe"}}, true},
		{"two corrections", "Nosotros bebe leche", "Nosotros no bebemos leche", []SentenceGradeIssue{{Original: "bebe", Corrected: "bebemos"}, {Original: "", Corrected: "no"}}, true},
		{"unsupported reorder", "red big house", "big red home", []SentenceGradeIssue{{Original: "house", Corrected: "home"}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sentenceIssuesCoverCorrection(tc.input, tc.corrected, tc.issues); got != tc.ok {
				t.Fatalf("got %v want %v", got, tc.ok)
			}
		})
	}
}

func TestSentenceGenerationDoesNotPublishEmptySet(t *testing.T) {
	svc := NewService("https://example.test", "default", "test", "", zap.NewNop())
	svc.SetSentenceModel("sentence-model")
	svc.SetSentenceGenPromptForCourse("es_ru", "Generate")
	calls := 0
	svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		var body ChatRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		if body.Model != "sentence-model" {
			t.Fatal("generation did not use sentence model")
		}
		calls++
		return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: `{"sentences":[]}`}}}}), nil
	})
	out, err := svc.GenerateSentenceSetForCourse(context.Background(), "es_ru", []GenSentenceWord{{"beber", "пить"}}, nil, []string{"presente"}, 1)
	if err == nil || len(out) != 0 || calls != 2 {
		t.Fatalf("out=%v err=%v calls=%d", out, err, calls)
	}
}

func TestSentenceMeaningAuditCanRemoveFalsePenalty(t *testing.T) {
	svc := NewService("https://example.test", "sentence", "test", "", zap.NewNop())
	svc.SetSentenceGradePromptForCourse("es_ru", "Grade")
	calls := 0
	svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		var body ChatRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		calls++
		raw := `{"error_count":1,"corrected_es":"Leo un libro","issues":[{"kind":"article","original":"el","corrected":"un","explanation":"Модель ошибочно выбрала артикль."}]}`
		if calls == 2 {
			// Gemini's native conversation adapter needs the continuation to end with
			// a user turn; appending a system turn after an assistant response can be empty.
			if body.Messages[len(body.Messages)-1].Role != "user" {
				t.Fatal("audit must request a new user turn")
			}
			raw = `{"error_count":0,"corrected_es":"Leo el libro","issues":[],"explanation":""}`
		}
		return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: raw}}}}), nil
	})
	grade, err := svc.GradeSentenceForCourse(context.Background(), "es_ru", "Я читаю книгу.", "На улице идёт дождь.", "Leo un libro.", "Leo el libro")
	if err != nil || calls != 2 || grade.ErrorCount != 0 || grade.Outcome != "star" {
		t.Fatalf("calls=%d grade=%+v err=%v", calls, grade, err)
	}
}

func TestSentenceReviewRequiresEvidenceForEveryChoice(t *testing.T) {
	for _, evidence := range []string{`[]`, `["Собеседник передал вам книгу."]`, `["Собеседник передал вам книгу.","Говорящий вежливо обращается к одному человеку."]`} {
		svc := NewService("https://example.test", "sentence", "test", "", zap.NewNop())
		svc.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			raw := `{"checks":[{"position":0,"accepted":true,"reason":"Проверено.","context_evidence":` + evidence + `}]}`
			return newJSONResponse(http.StatusOK, ChatResponse{Choices: []Choice{{Message: Message{Content: raw}}}}), nil
		})
		out, err := svc.reviewGeneratedSentenceQuality(context.Background(), "sentence", "es_ru", nil, nil, []string{"presente"}, []GeneratedSentence{{PromptRU: "Вы читаете книгу.", ReferenceES: "Usted lee el libro.", ClarificationRU: "Говорящий вежливо обращается к одному человеку. Собеседник передал вам книгу."}}, nil)
		want := 0
		if strings.Contains(evidence, "Говорящий") {
			want = 1
		}
		if err != nil || len(out.Accepted) != want {
			t.Fatalf("evidence=%s: accepted %d want %d err=%v", evidence, len(out.Accepted), want, err)
		}
	}
}
