---
"xbvr": patch
---

Update golang.org/x/net, x/crypto and x/image to versions with security fixes. The x/net one is the meaningful change: its HTML parser handles pages fetched by the scrapers, so a malformed page could previously stall a scrape.
