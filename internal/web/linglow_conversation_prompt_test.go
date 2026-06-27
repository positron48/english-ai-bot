package web

import (
	"strings"
	"testing"

	"tgbot-skeleton/internal/repository"
)

func TestBuildQuestEvalPromptIncludesTasks(t *testing.T) {
	scenario := &repository.ConversationScenario{
		NPCName:    "Mara",
		NPCPersona: "barista",
		SceneSetup: "A cozy cafe.",
		IsQuest:    true,
		CEFRLevel:  "A0",
	}
	tasks := []repository.ConversationTask{
		{ID: 1, Code: "greet", CompletionCriteria: "Say hello", IsRequired: true},
	}
	out := buildQuestEvalPrompt("quest base", "en", scenario, tasks, map[int64]bool{})
	if !strings.Contains(out, "quest base") {
		t.Error("missing quest base")
	}
	if !strings.Contains(out, "[greet]") {
		t.Error("missing task code")
	}
	if strings.Contains(out, "###CONTROL###") {
		t.Error("quest prompt must not mention control sentinel")
	}
}

func TestBuildCorrectionPromptIncludesCEFR(t *testing.T) {
	out := buildCorrectionPrompt("correction base", "es", "A2")
	if !strings.Contains(out, "correction base") || !strings.Contains(out, "A2") {
		t.Errorf("unexpected prompt: %q", out)
	}
}

func TestBuildNPCReplyPromptOmitsTaskList(t *testing.T) {
	scenario := &repository.ConversationScenario{
		NPCName:    "Mara",
		NPCPersona: "barista",
		SceneSetup: "A cozy cafe.",
		IsQuest:    true,
		CEFRLevel:  "A0",
	}
	tasks := []repository.ConversationTask{
		{ID: 1, Code: "greet", CompletionCriteria: "Say hello", IsRequired: true},
	}
	// NPC prompt builder does not receive tasks — ensure no task codes leak from scenario-only input.
	_ = tasks
	out := buildNPCReplyPrompt("npc base", "en", scenario, false)
	if !strings.Contains(out, "npc base") || !strings.Contains(out, "Mara") {
		t.Errorf("unexpected prompt: %q", out)
	}
	if strings.Contains(out, "[greet]") {
		t.Error("NPC reply prompt must not include task checklist")
	}
	if strings.Contains(out, "###CONTROL###") {
		t.Error("NPC reply prompt must not mention control sentinel")
	}
}

func TestBuildOpeningNPCPromptOmitsTasks(t *testing.T) {
	scenario := &repository.ConversationScenario{
		NPCName:    "Mara",
		NPCPersona: "barista",
		IsQuest:    true,
		CEFRLevel:  "A0",
	}
	out := buildOpeningNPCPrompt("npc base", "en", scenario)
	if !strings.Contains(out, "Greet the learner") {
		t.Error("missing opening instruction")
	}
	if strings.Contains(out, "TASKS") {
		t.Error("opening prompt must not include tasks")
	}
}
