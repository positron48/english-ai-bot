package placementbundle

import (
	"math/rand"
	"testing"
	"tgbot-skeleton/internal/placement"
)

func TestPlacementReleasedBanks(t *testing.T) {
	for _, code := range []string{"en_ru", "es_ru"} {
		t.Run(code, func(t *testing.T) {
			b, err := Load(code)
			if err != nil {
				t.Fatal(err)
			}
			if len(b.Items) < 400 {
				t.Fatal("incomplete release")
			}
			recent := map[string]int{}
			for attempt := 1; attempt <= 3; attempt++ {
				items, repeats, e := placement.Select(b, placement.Levels, recent, nil, rand.New(rand.NewSource(int64(attempt))))
				if e != nil || repeats > 0 {
					t.Fatalf("retake%d repeated=%d: %v", attempt, repeats, e)
				}
				// Simulate maximum clarification use at each level over three attempts.
				extra, repeats, e := placement.Select(b, placement.Levels, recent, items, rand.New(rand.NewSource(int64(attempt+100))))
				if e != nil || repeats > 0 {
					t.Fatalf("clarification reserve%d repeated=%d: %v", attempt, repeats, e)
				}
				for _, q := range append(items, extra...) {
					if recent[q.FamilyID] > 0 {
						t.Fatal("family repeated")
					}
					recent[q.FamilyID] = attempt
				}
			}
		})
	}
}
