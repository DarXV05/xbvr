## Context

See proposal.md — Why.

The constraint that shapes everything here is that enumeration happens in ~45 separate
scraper files, each with its own listing selector, and there is no shared code path that
knows a given link is a scene link. A generic hook cannot tell a scene link from a
navigation link, so a counter installed in `createCollector` would never read zero even
when the scene selector matched nothing — which is exactly the case that must be detected.

Two facts about the existing code make a narrow intervention possible:

- 41 of 53 files in `pkg/scrape` call `funk.ContainsString(knownScenes, sceneURL)` to
  decide whether a link is worth visiting — 43 call sites, most of them character-identical
  (`if !funk.ContainsString(knownScenes, sceneURL) {`). One, `r18.go`, has no scraper id in
  scope at the call site and is left out of the first pass.
- That call sits exactly at the point where a link has been recognised as a scene and
  before the decision to skip it. It runs for known and unknown scenes alike, which is
  what separates "nothing new" from "nothing found".

The remaining scrapers — the SLR, StashDB, TPDB and JAV families — read APIs rather than
walking listing pages and have no equivalent point.

## Goals / Non-Goals

**Goals:**
- One enumeration counter per site per run, incremented at a single code path.
- Impossible to add a scraper that forgets to count.
- No change to `ScraperFunc`, so scraper registration and the 45 signatures stay as they are.

**Non-Goals:**
- Covering API-based scrapers in the first pass.
- Distinguishing "listing page returned an error" from "listing page parsed but matched
  nothing". Both are zero enumeration and both mean broken.
- Any counting inside `sceneCollector`. It only visits unknown scenes, so its volume
  tracks new content, not scraper health.

## Decisions

**Fuse the counter to the known-scene check rather than adding a separate call.**

Replace the `funk.ContainsString(knownScenes, sceneURL)` call sites with a helper in
`pkg/scrape`:

```go
func isKnownScene(scraperID string, knownScenes []string, sceneURL string) bool
```

which records the enumeration and returns the same boolean. For 29 call sites this is a
mechanical substitution; the 12 with extra conditions keep their conditions around it.

Alternative considered: a separate `markEnumerated(scraperID)` line above each check.
Rejected because a new scraper can be written without it and nothing would fail — the
site would simply look permanently broken or permanently healthy depending on the default.
Fusing it to a call the scraper must already make removes that failure mode.

Alternative considered: threading a counter through `ScraperFunc`. Rejected — it changes
the signature of every scraper and every registration for no gain over a keyed counter.

**Key the counter by scraperID in package-level state, not on `ScrapeWG`.**

`runScrapers` creates one `ScrapeWG` for the whole run and shares it across every site,
so it cannot hold per-site counts without being keyed anyway. Scrapers run concurrently,
so the store must be concurrency-safe — a `sync.Map` of `scraperID` to counter, reset when
`logScrapeStart` fires and read when `logScrapeFinished` fires. Those two functions already
bracket every scraper and already carry `scraperID`.

**Persist a health record separate from `Site.LastUpdate`.**

`LastUpdate` keeps meaning "when the scraper last ran". A new record holds the last run's
enumerated and emitted counts and a consecutive-zero-enumeration counter. Conflating them
would break the scheduling that reads `LastUpdate`.

**Mark suspect after N consecutive zero-enumeration runs, not the first.**

A single zero can be a transient block, a rate limit, or a network blip. N is a
configuration value; 3 is a reasonable default given the scrape schedule, and the counter
resets on any run that enumerates.

## Risks / Trade-offs

**A site legitimately publishing nothing still enumerates its back catalogue, so zero
enumeration really does mean broken** → This holds only while scrapers walk listing pages
that show existing scenes. A site that changes to an empty-by-default listing would read
as broken. Accepted: that is indistinguishable from broken without parsing further, and a
false positive that says "check this scraper" is cheap.

**The 12 non-identical call sites are where a mechanical edit goes wrong** → They must be
converted individually and each read, not swept with a regex. Their surrounding conditions
(`strings.Contains(sceneURL, "/vr-porn/")` and similar) filter links after the known check
and must keep their current order.

**API-based scrapers stay invisible to this** → They are the families most likely to fail
loudly (an API returns an error status) rather than silently, so they are lower priority.
Worth stating in the UI that a site has no health signal rather than showing it as healthy.

**Suspect status is only as visible as the UI makes it** → If it lands as a subtle badge
on a settings page nobody opens, the change achieves nothing. Placement matters more than
the detection here.

## Migration Plan

The health record needs a new table and a migration in `pkg/migrations/migrations.go`,
appended as a new sequentially numbered entry. Existing sites get no health record until
their first scrape after upgrade, so the UI must treat absent as unknown rather than
suspect.

Rollback is dropping the table; nothing else reads it.

## Resolved

**N is a constant of 3, not a configuration field.** The spec requires a configured
threshold but does not require it to be user-editable, and adding an `ObjectConfig` field
pulls in an API endpoint, a UI control and a backup-schema entry for a value nobody can
tune sensibly before seeing whether 3 is right in practice. Promote it to configuration
if real use shows 3 is wrong.

**Both counts are persisted.** The spec's scenarios record recognised and added counts
together, so this is settled there rather than being a design choice.
