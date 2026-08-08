package scrape

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/nleeper/goment"
	"github.com/thoas/go-funk"
	"github.com/xbapps/xbvr/pkg/models"
)

var (
	vrxdbPostID   = regexp.MustCompile(`(?:^|\s)post-(\d+)(?:\s|$)`)
	vrxdbStudio   = regexp.MustCompile(`(?:^|\s)vr_porn_studios-([a-z0-9-]+)(?:\s|$)`)
	vrxdbCategory = regexp.MustCompile(`(?:^|\s)video_categories-([a-z0-9-]+)`)
)

func vrxdbSlugToName(slug string) string {
	return strings.TrimSpace(strings.ReplaceAll(slug, "-", " "))
}

func VRXDB(wg *models.ScrapeWG, updateSite bool, knownScenes []string, out chan<- models.ScrapedScene, singleSceneURL string, singeScrapeAdditionalInfo string, limitScraping bool) error {
	defer wg.Done()
	scraperID := "vrxdb"
	siteID := "VRXDB"
	logScrapeStart(scraperID, siteID)

	sceneCollector := createCollector("vrxdb.com")
	siteCollector := createCollector("vrxdb.com")

	sceneCollector.OnHTML(`html`, func(e *colly.HTMLElement) {
		sc := models.ScrapedScene{}
		sc.ScraperID = scraperID
		sc.SceneType = "VR"
		sc.Studio = siteID
		sc.Site = siteID
		sc.HomepageURL = strings.Split(e.Request.URL.String(), "?")[0]

		// This wrapper carries the post id, studio and categories as taxonomy
		// slugs, which are stable where the sidebar links are not.
		postClass := e.ChildAttr(`.elementor-location-single`, "class")

		if m := vrxdbPostID.FindStringSubmatch(postClass); m != nil {
			sc.SiteID = m[1]
			sc.SceneID = scraperID + "-" + sc.SiteID
		}
		if m := vrxdbStudio.FindStringSubmatch(postClass); m != nil {
			sc.Studio = vrxdbSlugToName(m[1])
		}
		for _, m := range vrxdbCategory.FindAllStringSubmatch(postClass, -1) {
			tag := vrxdbSlugToName(m[1])
			if tag != "" && !funk.ContainsString(sc.Tags, tag) {
				sc.Tags = append(sc.Tags, tag)
			}
		}

		e.ForEach(`dl8-video`, func(id int, e *colly.HTMLElement) {
			if id == 0 {
				sc.Title = strings.TrimSpace(e.Attr("title"))
			}
		})
		if sc.Title == "" {
			sc.Title = strings.TrimSpace(strings.Split(e.ChildAttr(`meta[property="og:title"]`, "content"), "|")[0])
		}

		if cover := e.ChildAttr(`meta[property="og:image"]`, "content"); cover != "" {
			sc.Covers = append(sc.Covers, strings.Split(cover, "?")[0])
		}

		// The gallery <img> is lazy-loaded; the full image sits on the lightbox anchor.
		e.ForEach(`.elementor-widget-gallery a[href]`, func(id int, e *colly.HTMLElement) {
			img := strings.Split(e.Request.AbsoluteURL(e.Attr("href")), "?")[0]
			if img != "" && !funk.ContainsString(sc.Gallery, img) {
				sc.Gallery = append(sc.Gallery, img)
			}
		})

		// Elementor renders desktop and mobile copies of each field, and the
		// wrapper classes are identical, so the label prefix is the only
		// reliable discriminator.
		sc.ActorDetails = make(map[string]models.ActorDetails)
		e.ForEach(`.jet-listing-dynamic-field__content`, func(id int, e *colly.HTMLElement) {
			text := strings.TrimSpace(e.Text)
			switch {
			case strings.HasPrefix(text, "Starring:"):
				for _, name := range strings.Split(strings.TrimPrefix(text, "Starring:"), ",") {
					name = strings.TrimSpace(name)
					if name != "" && !funk.ContainsString(sc.Cast, name) {
						sc.Cast = append(sc.Cast, name)
						sc.ActorDetails[name] = models.ActorDetails{Source: scraperID + " scrape"}
					}
				}
			case strings.HasPrefix(text, "Released:"):
				if sc.Released == "" {
					if d, err := goment.New(strings.TrimSpace(strings.TrimPrefix(text, "Released:")), "MMMM D, YYYY"); err == nil {
						sc.Released = d.Format("YYYY-MM-DD")
					}
				}
			case strings.HasPrefix(text, "Duration:"):
				if sc.Duration == 0 {
					if m := regexp.MustCompile(`(\d+)`).FindStringSubmatch(text); m != nil {
						sc.Duration, _ = strconv.Atoi(m[1])
					}
				}
			}
		})

		if desc := e.ChildAttr(`meta[property="og:description"]`, "content"); desc != "" {
			sc.Synopsis = strings.TrimSpace(desc)
		}

		sc.TrailerType = "scrape_html"
		params := models.TrailerScrape{
			SceneUrl:    sc.HomepageURL,
			HtmlElement: "dl8-video source",
			ContentPath: "src",
			QualityPath: "quality",
		}
		tmp, _ := json.Marshal(params)
		sc.TrailerSrc = string(tmp)

		if sc.SceneID != "" {
			out <- sc
		}
	})

	siteCollector.OnHTML(`a[href*="/vr-porn-videos/"]`, func(e *colly.HTMLElement) {
		sceneURL := strings.Split(e.Request.AbsoluteURL(e.Attr("href")), "?")[0]
		if !strings.Contains(sceneURL, "/vr-porn-videos/") ||
			strings.Contains(sceneURL, "/page/") ||
			strings.TrimSuffix(sceneURL, "/") == "https://vrxdb.com/vr-porn-videos" {
			return
		}
		if !isKnownScene(scraperID, knownScenes, sceneURL) {
			sceneCollector.Visit(sceneURL)
		}
	})

	siteCollector.OnHTML(`a.next.page-numbers`, func(e *colly.HTMLElement) {
		if !limitScraping {
			siteCollector.Visit(e.Request.AbsoluteURL(e.Attr("href")))
		}
	})

	if singleSceneURL != "" {
		sceneCollector.Visit(singleSceneURL)
	} else {
		siteCollector.Visit("https://vrxdb.com/vr-porn-videos/")
	}

	if updateSite {
		updateSiteLastUpdate(scraperID)
	}
	logScrapeFinished(scraperID, siteID)
	return nil
}

func init() {
	registerScraper("vrxdb", "VRXDB", "https://vrxdb.com/favicon.ico", "vrxdb.com", VRXDB)
}
