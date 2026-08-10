---
"xbvr": patch
---

Fix a crash when DeoVR or HereSphere opened a scene that has no matched video file. The request failed with no response at all, which appears as a gateway error behind a reverse proxy.
