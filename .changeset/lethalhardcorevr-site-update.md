---
"xbvr": patch
---

Fix the LethalHardcoreVR scraper, which had stopped importing anything after the site moved to a client-rendered app. It now reads the site's own search index instead of the page HTML. WhorecraftVR remains registered but that site no longer has any scenes.
