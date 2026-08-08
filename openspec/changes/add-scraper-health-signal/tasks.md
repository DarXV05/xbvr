## 1. Counting

- [ ] 1.1 Add a concurrency-safe per-run store in `pkg/scrape` keyed by scraper id, holding recognised and added counts
- [ ] 1.2 Add `isKnownScene(scraperID string, knownScenes []string, sceneURL string) bool` that records a recognised scene and returns the same result the existing check returns
- [ ] 1.3 Reset a scraper's counts in `logScrapeStart` and read them in `logScrapeFinished`
- [ ] 1.4 Add a unit test proving concurrent scrapers do not cross-contaminate each other's counts

## 2. Converting call sites

- [ ] 2.1 Convert the 29 identical `if !funk.ContainsString(knownScenes, sceneURL) {` sites
- [ ] 2.2 Convert `baberoticavr.go:88` and `wetvr.go:85` — both invert the test, so the recorded count must still increment on the known branch
- [ ] 2.3 Convert `povr.go:130`, `sinsvr.go:186`, `realjamvr.go:190`, `vr3000.go:95` — each filters the URL after the known check; keep the filters and their order
- [ ] 2.4 Convert `upclosevr.go:82` — the `|| singleSceneURL != ""` branch means a single-scene run must not be recorded as a listing enumeration
- [ ] 2.5 Convert `vrspy.go:370` and `vrspy.go:378` — two sites in one file, both also testing a local `sceneURLs` slice
- [ ] 2.6 Convert `vrporn.go:137`, `stashdb_studios.go:33`, `slrstudios.go:631` — these build the URL inline rather than using `sceneURL`
- [ ] 2.7 Verify no `funk.ContainsString(knownScenes` call sites remain unconverted

## 3. Persistence

- [ ] 3.1 Add a site health model holding last recognised count, last added count, consecutive-zero count, and time last recognised
- [ ] 3.2 Add a migration creating the table, appended as a new sequentially numbered entry
- [ ] 3.3 Write the health record at the end of each site's scrape, leaving `Site.LastUpdate` behaviour unchanged
- [ ] 3.4 Increment the consecutive-zero count on a zero-recognised run and reset it to zero on any run that recognises at least one scene

## 4. Suspect state

- [ ] 4.1 Derive suspect state from the consecutive-zero count against a threshold constant of 3
- [ ] 4.2 Return unknown, not healthy, for a site with no health record
- [ ] 4.3 Return unknown for sites whose scrapers never record a recognised count, so API-based scrapers are not shown as healthy

## 5. Surfacing

- [ ] 5.1 Expose health state and time last recognised on the sites API response
- [ ] 5.2 Show suspect sites in the scraper list with the time they last recognised any scenes
- [ ] 5.3 Confirm the indicator is visible without opening a per-site detail view

## 6. Verification

- [ ] 6.1 Confirm a site that recognises scenes but adds none is not flagged, across repeated runs
- [ ] 6.2 Confirm a site recognising nothing is flagged only on the third consecutive run
- [ ] 6.3 Confirm a recognising run clears both the count and an existing flag
- [ ] 6.4 Run a real scrape against the five sites that no longer respond — `vr3000`, `stasyqvr`, `r18`, `javlibrary`, and the amateur site in the VirtualRealPorn family — and confirm all five reach suspect
- [ ] 6.5 Run a real scrape against a working site and confirm it is not flagged
