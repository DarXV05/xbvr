# xbvr

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
