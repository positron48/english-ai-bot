package readingcms

import "testing"

func TestBatchPercentTwoPhase(t *testing.T) {
	cases := []struct {
		name           string
		total          int
		phaseDone1     int
		phaseDone2     int
		current        int
		currentPercent int
		running        bool
		done           bool
		errMsg         string
		want           int
	}{
		{"empty", 0, 0, 0, 0, 0, false, false, "", 0},
		{"done", 5, 5, 5, 0, 0, false, true, "", 100},
		{"half first phase one at 50", 4, 1, 0, 2, 50, true, false, "", 18},
		{"first phase done", 3, 3, 0, 0, 0, false, false, "", 50},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := batchPercentTwoPhase(tc.total, tc.phaseDone1, tc.phaseDone2, tc.current, tc.currentPercent, tc.running, tc.done, tc.errMsg)
			if got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}
}
