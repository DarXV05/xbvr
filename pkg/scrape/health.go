package scrape

import (
	"sync"

	"github.com/thoas/go-funk"
)

type scrapeCounts struct {
	Recognised int
	Added      int
}

var scrapeHealth sync.Map

func resetScrapeCounts(scraperID string) {
	scrapeHealth.Store(scraperID, &scrapeCounts{})
}

func countsFor(scraperID string) *scrapeCounts {
	v, _ := scrapeHealth.LoadOrStore(scraperID, &scrapeCounts{})
	return v.(*scrapeCounts)
}

var scrapeHealthMu sync.Mutex

// isKnownScene records that a scene link was recognised on a listing page and
// reports whether it is already stored. Recognising nothing across a whole run
// is what separates a broken scraper from a site with no new releases.
func isKnownScene(scraperID string, knownScenes []string, sceneURL string) bool {
	c := countsFor(scraperID)
	known := funk.ContainsString(knownScenes, sceneURL)

	scrapeHealthMu.Lock()
	c.Recognised++
	if !known {
		c.Added++
	}
	scrapeHealthMu.Unlock()

	return known
}

func ScrapeCounts(scraperID string) (recognised int, added int) {
	c := countsFor(scraperID)
	scrapeHealthMu.Lock()
	defer scrapeHealthMu.Unlock()
	return c.Recognised, c.Added
}
