---
"xbvr": patch
---

Fix the CovertJapan scraper importing nothing. Its studio URL uses a hyphenated slug that did not match the internal studio lookup, so the scraper never queried SexLikeReal at all while still reporting a successful run.
