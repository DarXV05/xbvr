# xbvr

## 0.8.1

### Patch Changes

- 6b5a185: Fix the CovertJapan scraper importing nothing. Its studio URL uses a hyphenated slug that did not match the internal studio lookup, so the scraper never queried SexLikeReal at all while still reporting a successful run.

## 0.8.0

### Minor Changes

- bbcf539: Show which scrapers have silently stopped working. The scraper list now marks a site suspect when it has recognised no scenes on three consecutive runs, which distinguishes a scraper broken by a site change from one that simply has nothing new. Sites that report no signal are shown as unknown rather than healthy.

## 0.7.0

### Minor Changes

- b1607c9: Add a scraper for vrxdb.com. It brings in title, studio, cast, tags, cover, gallery, duration, release date and trailer. It cannot supply filenames, so its scenes will not match files on disk automatically and have to be matched by hand.

### Patch Changes

- ca9df27: Update golang.org/x/net, x/crypto and x/image to versions with security fixes. The x/net one is the meaningful change: its HTML parser handles pages fetched by the scrapers, so a malformed page could previously stall a scrape.
- c351b9b: Fix the LethalHardcoreVR scraper, which had stopped importing anything after the site moved to a client-rendered app. It now reads the site's own search index instead of the page HTML. WhorecraftVR remains registered but that site no longer has any scenes.

## 0.6.0

### Minor Changes

- 39d39b5: When matching a file to a scene, rank scenes of a similar length higher. The filename text search alone often misses the right scene, and duration is a strong signal it ignored.

### Patch Changes

- 067ab72: Fix scenes whose titles contain an apostrophe being unmatchable. Searching for such a title found nothing whether the apostrophe was typed or omitted, which also made the affected files impossible to match from the file list.

## 0.5.2

### Patch Changes

- 839f932: Fix RealJamVR and PornCornVR returning 403 on every scrape. Both sites blocklist the exact User-Agent string XBVR shipped, so the scrapers failed silently on every scheduled run while still reporting success.

## 0.5.1

### Patch Changes

- 2ad1f0b: Releases no longer create a GitHub release page or upload tarballs. The Docker image is the deliverable and CHANGELOG.md carries the release notes.

## 0.5.0

### Minor Changes

- 922092c: Releases now ship a single linux/amd64 Docker image and Linux x86_64 tarball. macOS, Windows and ARM artifacts are no longer built, cutting release time from roughly 12 minutes to 3.

### Patch Changes

- 922092c: Fix the VirtualRealPorn scraper after the site moved to Livewire. Scenes from that site stopped importing entirely, with no error reported. Performer names are also stripped of the site's " VR" suffix so they no longer create duplicate actor records, and trailers resolve again.
