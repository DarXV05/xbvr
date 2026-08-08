---
"xbvr": patch
---

Fix RealJamVR and PornCornVR returning 403 on every scrape. Both sites blocklist the exact User-Agent string XBVR shipped, so the scrapers failed silently on every scheduled run while still reporting success.
