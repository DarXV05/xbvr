package api

import (
	"testing"

	"github.com/tidwall/gjson"
)

func TestUsableChromaKey(t *testing.T) {
	const placeholder = `{"enabled":false,"hue":0,"opacity":1,"saturation":0,"threshold":0,"value":0}`

	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"empty is absent", "", false},
		{"invalid json is absent", "not json", false},
		{"scraper placeholder is absent", placeholder, false},
		{"placeholder with short names is absent", `{"enabled":false,"h":0,"s":0,"v":0,"threshold":0,"opacity":1}`, false},
		{"enabled is usable", `{"enabled":true,"hue":0,"saturation":0,"value":0}`, true},
		{"non-zero hue is usable even when disabled", `{"enabled":false,"hue":120,"saturation":0,"value":0}`, true},
		{"non-zero short-name channel is usable", `{"enabled":false,"h":0,"s":0.4,"v":0}`, true},
		{"non-zero threshold is usable", `{"enabled":false,"threshold":0.2}`, true},
		{"hasAlpha is usable", `{"enabled":false,"hasAlpha":true}`, true},
		{"opacity alone is not a key", `{"enabled":false,"opacity":1}`, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := usableChromaKey(c.raw); got != c.want {
				t.Errorf("usableChromaKey(%s) = %v, want %v", c.raw, got, c.want)
			}
		})
	}
}

func TestChromaKeyFloatFallsBackToLongName(t *testing.T) {
	long := gjson.Parse(`{"hue":120,"saturation":0.5,"value":0.25}`)
	if got := chromaKeyFloat(long, "h", "hue"); got != 120 {
		t.Errorf("hue = %v, want 120", got)
	}
	if got := chromaKeyFloat(long, "s", "saturation"); got != 0.5 {
		t.Errorf("saturation = %v, want 0.5", got)
	}

	// The short name wins when it carries a value.
	both := gjson.Parse(`{"h":90,"hue":120}`)
	if got := chromaKeyFloat(both, "h", "hue"); got != 90 {
		t.Errorf("h = %v, want 90", got)
	}
}
