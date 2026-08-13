package aiprovider

import "testing"

func TestParseVerdict(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		decision string
		conf     float64
	}{
		{"plain approved", `{"decision":"approved","confidence":0.95,"reason":"ok"}`, "approved", 0.95},
		{"rejected with prose", `Here's my assessment: {"decision":"rejected","confidence":0.1,"reason":"spam"}`, "rejected", 0.1},
		{"uncertain code fence", "```json\n{\"decision\":\"uncertain\",\"confidence\":0.5}\n```", "uncertain", 0.5},
		{"unknown decision normalised", `{"decision":"maybe","confidence":0.7}`, "uncertain", 0.7},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := parseVerdict(c.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r.Decision != c.decision {
				t.Errorf("expected decision %q, got %q", c.decision, r.Decision)
			}
			if r.Confidence != c.conf {
				t.Errorf("expected confidence %v, got %v", c.conf, r.Confidence)
			}
		})
	}
}

func TestParseVerdict_Invalid(t *testing.T) {
	if _, err := parseVerdict("not json at all"); err == nil {
		t.Error("expected error for invalid verdict, got nil")
	}
}
