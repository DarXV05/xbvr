package scrape

import (
	"strconv"
	"sync"
	"testing"
)

func TestIsKnownSceneCountsRecognisedAndAdded(t *testing.T) {
	resetScrapeCounts("t-basic")
	known := []string{"https://a/1", "https://a/2"}

	if !isKnownScene("t-basic", known, "https://a/1") {
		t.Error("expected an already-stored scene to report as known")
	}
	if isKnownScene("t-basic", known, "https://a/9") {
		t.Error("expected an unseen scene to report as not known")
	}

	recognised, added := ScrapeCounts("t-basic")
	if recognised != 2 {
		t.Errorf("recognised = %d, want 2 - known scenes must still count", recognised)
	}
	if added != 1 {
		t.Errorf("added = %d, want 1", added)
	}
}

// A site with nothing new still walks its listing, so recognised stays above
// zero. Conflating that with a broken scraper is the bug this guards.
func TestQuietSiteStillRecognises(t *testing.T) {
	resetScrapeCounts("t-quiet")
	known := []string{"https://a/1", "https://a/2", "https://a/3"}
	for _, u := range known {
		isKnownScene("t-quiet", known, u)
	}

	recognised, added := ScrapeCounts("t-quiet")
	if recognised != 3 {
		t.Errorf("recognised = %d, want 3", recognised)
	}
	if added != 0 {
		t.Errorf("added = %d, want 0", added)
	}
}

func TestBrokenScraperRecognisesNothing(t *testing.T) {
	resetScrapeCounts("t-broken")
	if recognised, _ := ScrapeCounts("t-broken"); recognised != 0 {
		t.Errorf("recognised = %d, want 0", recognised)
	}
}

func TestResetClearsPreviousRun(t *testing.T) {
	resetScrapeCounts("t-reset")
	isKnownScene("t-reset", nil, "https://a/1")
	resetScrapeCounts("t-reset")

	recognised, added := ScrapeCounts("t-reset")
	if recognised != 0 || added != 0 {
		t.Errorf("after reset got recognised=%d added=%d, want 0/0", recognised, added)
	}
}

// Scrapers run concurrently against one shared store, so a count landing on the
// wrong site would silently mark a working scraper broken.
func TestConcurrentScrapersDoNotCrossContaminate(t *testing.T) {
	const sites = 8
	const perSite = 200

	var wg sync.WaitGroup
	for s := 0; s < sites; s++ {
		id := "t-conc-" + strconv.Itoa(s)
		resetScrapeCounts(id)
		wg.Add(1)
		go func(id string, n int) {
			defer wg.Done()
			for i := 0; i < n; i++ {
				isKnownScene(id, nil, "https://a/"+strconv.Itoa(i))
			}
		}(id, perSite*(s+1))
	}
	wg.Wait()

	for s := 0; s < sites; s++ {
		id := "t-conc-" + strconv.Itoa(s)
		want := perSite * (s + 1)
		recognised, added := ScrapeCounts(id)
		if recognised != want {
			t.Errorf("%s recognised = %d, want %d", id, recognised, want)
		}
		if added != want {
			t.Errorf("%s added = %d, want %d", id, added, want)
		}
	}
}
