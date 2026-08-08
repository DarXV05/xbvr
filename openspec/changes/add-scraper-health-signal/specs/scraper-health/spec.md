## Purpose

Records whether a scrape run actually extracted anything from a site, so that a scraper
broken by a site change is distinguishable from a working scraper with no new content,
and surfaces sites whose scrapers have silently stopped working.

## ADDED Requirements

### Requirement: Record what a scrape run found

Each scrape run of a site SHALL record how many scene links the run recognised on that
site's listing pages, counted independently of whether those scenes were already stored.

The recorded count MUST include scenes already in the library. A run that recognises only
already-stored scenes has worked correctly and MUST NOT be recorded as having found
nothing.

#### Scenario: Site with new scenes

- **WHEN** a scrape run recognises 40 scene links, 3 of which are not yet stored
- **THEN** the run records 40 recognised and 3 added

#### Scenario: Site with nothing new since the last run

- **WHEN** a scrape run recognises 40 scene links and all 40 are already stored
- **THEN** the run records 40 recognised and 0 added

#### Scenario: Site whose listing no longer matches

- **WHEN** a scrape run recognises no scene links at all
- **THEN** the run records 0 recognised and 0 added

### Requirement: Distinguish a quiet site from a broken scraper

The system SHALL treat zero recognised scene links as evidence that a scraper is broken,
and SHALL NOT treat zero added scenes as evidence of anything.

#### Scenario: Quiet site is not flagged

- **WHEN** a site records a run with scenes recognised and none added
- **THEN** the site is not flagged as suspect, however many consecutive runs add nothing

#### Scenario: Broken scraper is flagged

- **WHEN** a site records consecutive runs recognising no scene links, reaching the
  configured threshold
- **THEN** the site is flagged as suspect

### Requirement: Require repeated failure before flagging

A site SHALL be flagged as suspect only after recognising no scene links on a configured
number of consecutive runs, so that a single blocked request, rate limit or network error
does not flag a working scraper.

The consecutive count SHALL reset whenever a run recognises at least one scene link.

#### Scenario: Single failed run does not flag

- **WHEN** a site recognises no scene links on one run and the threshold is greater than one
- **THEN** the site is not flagged as suspect

#### Scenario: Recovery clears the count

- **WHEN** a site has consecutive runs recognising nothing, below the threshold, and a
  later run recognises at least one scene link
- **THEN** the consecutive count returns to zero and the site is not flagged

#### Scenario: Recovery clears an existing flag

- **WHEN** a site is flagged as suspect and a later run recognises at least one scene link
- **THEN** the site is no longer flagged as suspect

### Requirement: Preserve the meaning of the last-update time

Recording scrape health SHALL NOT change when a site's last-update time is set. That time
continues to report when the scraper last ran, independently of whether the run found
anything.

#### Scenario: Broken scraper still records that it ran

- **WHEN** a scrape run recognises no scene links
- **THEN** the site's last-update time is still set to the time of that run

### Requirement: Report absence of a health signal as unknown

A site whose scraper does not report recognised scene links SHALL be presented as having
an unknown health state, and MUST NOT be presented as healthy.

#### Scenario: Site that never reports counts

- **WHEN** a site has completed scrape runs but has never recorded a recognised count
- **THEN** its health state is presented as unknown

#### Scenario: Site not yet scraped since the feature was added

- **WHEN** a site has no recorded scrape health
- **THEN** its health state is presented as unknown rather than suspect

### Requirement: Surface suspect sites to the user

Sites flagged as suspect SHALL be visible where a user manages sites and scrapers, showing
that the scraper appears to have stopped working and when it last recognised any scenes.

#### Scenario: Suspect site is visible in the site list

- **WHEN** a user views the list of sites and one is flagged as suspect
- **THEN** that site is shown as suspect alongside the time it last recognised any scenes
