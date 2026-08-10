package api

import "testing"

func TestSceneSlug(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want string
	}{
		{"studio url", "https://www.naughtyamerica.com/scene/fuck-the-masseuse-25635", "fuck-the-masseuse"},
		{"aggregator url for the same scene", "https://povr.com/vr-porn/fuck-the-masseuse-5344801", "fuck-the-masseuse"},
		{"trailing slash", "https://example.com/scene/some-title-123/", "some-title"},
		{"query string is ignored", "https://example.com/scene/some-title-123?utm=x", "some-title"},
		{"no trailing id", "https://example.com/scene/some-title", "some-title"},
		{"digits inside the slug survive", "https://example.com/scene/scene-4k-edition", "scene-4k-edition"},
		{"empty path", "https://example.com", ""},
		{"unparseable", "://nope", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sceneSlug(c.url); got != c.want {
				t.Errorf("sceneSlug(%q) = %q, want %q", c.url, got, c.want)
			}
		})
	}
}

func TestSceneSlugMatchesAcrossSites(t *testing.T) {
	studio := sceneSlug("https://www.naughtyamerica.com/scene/fuck-the-masseuse-25635")
	aggregator := sceneSlug("https://povr.com/vr-porn/fuck-the-masseuse-5344801")
	if studio == "" || studio != aggregator {
		t.Errorf("slugs differ: studio=%q aggregator=%q", studio, aggregator)
	}
}

func TestPadOsHash(t *testing.T) {
	if got := padOsHash("fd4e88f6fa1e5ff1"); got != "fd4e88f6fa1e5ff1" {
		t.Errorf("16 chars should be unchanged, got %q", got)
	}
	if got := padOsHash("4e88f6fa1e5ff1"); got != "004e88f6fa1e5ff1" {
		t.Errorf("short hash = %q, want 004e88f6fa1e5ff1", got)
	}
	if got := len(padOsHash("")); got != 16 {
		t.Errorf("empty hash padded to %d chars, want 16", got)
	}
}
