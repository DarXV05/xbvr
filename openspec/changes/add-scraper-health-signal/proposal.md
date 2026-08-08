## Why

A broken scraper is indistinguishable from a quiet one. `updateSiteLastUpdate()` sets
`Site.LastUpdate = time.Now()` at the end of every run with no check on whether anything
was found, so a site whose selectors stopped matching keeps reporting a fresh, healthy
timestamp forever.

This is not hypothetical. The VirtualRealPorn scraper matched nothing after that site
moved to Livewire — `SceneID` was never set, so `out <- sc` never fired — and the run
still logged `completed=true` and advanced `LastUpdate`. It was reported upstream as
issue #2222 by a user who noticed scenes had stopped appearing, 26 days before anyone
looked at the code. Upstream #2234 describes an independent instance: two scrapers
returning 403 on every scheduled run for two days, found by accident while chasing an
unrelated report.

A survey of all 119 listing URLs in `pkg/scrape` established what cannot be used to
detect this. HTTP status finds only network-layer failures — five sites no longer answer
at all — and is structurally blind to the failure that actually occurred: VirtualRealPorn
returned **200 on every request** while extracting nothing. The page loaded; the
selectors matched zero elements. Any check based on reachability would have passed it.

## What Changes

- Scrapers report how many scene links they enumerated on listing pages, separately from
  how many scenes they emitted. These two numbers distinguish the cases that
  `LastUpdate` currently conflates:

  | enumerated | emitted | meaning |
  |---|---|---|
  | > 0 | > 0 | working, new content |
  | > 0 | 0 | working, nothing new since last run |
  | **0** | 0 | **listing selector matched nothing — broken** |

- A per-site health record persists the last run's counts and a count of consecutive
  runs that enumerated zero links.
- A site that enumerates zero links on N consecutive runs is marked suspect and surfaced
  in the UI, rather than continuing to look current.
- `LastUpdate` keeps its present meaning — when the scraper last ran. The health record
  answers the separate question of whether it worked.

Not in scope: automatically disabling suspect sites, alerting outside the UI, and
detecting partial breakage where a listing still enumerates but scene pages fail.

## Capabilities

### New Capabilities
- `scraper-health`: recording whether a scrape run actually extracted anything, and
  exposing sites whose scrapers have silently stopped working.

### Modified Capabilities

None. No existing spec describes scrape success or failure reporting.

## Impact

- `pkg/scrape`: every scraper's listing callback is where enumeration happens. There are
  ~45 scraper files, so the counter lives in shared code rather than being duplicated per
  file.
- `pkg/models`: a new health record, and a migration to create it.
- `pkg/api` and `ui/src`: surfacing suspect sites in the scraper list.
- No change to scene or file matching, and no change to what a healthy scrape produces.

Five scrapers point at sites that no longer respond at all — `vr3000`, `stasyqvr`,
`r18`, `javlibrary`, and the amateur site in the VirtualRealPorn family. They are a
separate cleanup, but they are also the cheapest test data available for this feature:
whatever is built here should mark all five suspect on the first run.
