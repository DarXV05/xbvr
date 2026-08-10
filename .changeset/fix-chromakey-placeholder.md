---
"xbvr": patch
---

Greenscreen scenes now use DeoVR's automatic chroma key again. Scenes scraped with an empty, disabled chroma key were sending that placeholder to the player, which turned its own keying off and left the green background visible until you set the channels by hand.
