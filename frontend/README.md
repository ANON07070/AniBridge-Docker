<p align="center">
  <a href="README.pt-BR.md">🇧🇷 Português</a> ·
  <b>🇺🇸 English</b>
</p>

# AniBridge — Frontend

Plain HTML/CSS/JavaScript, no framework and no build step. Served
directly by FastAPI (`StaticFiles` mounted at `/`), on the same origin
as the API — no CORS to configure.

## Structure

```
frontend/
├── index.html    # search form, episode list, player
├── css/style.css
└── js/app.js
```

## How it works

`index.html` has two sections toggled by JavaScript, no page reloads:

1. **Search view**: search form + result cards. Clicking a card opens
   the episode view for that anime (`GET /api/episodes`).
2. **Episode view**: episode list for the selected anime. Each episode
   gets a "Watch" button *only* when its source has playback support
   (`isEmbedURL`/`isPlayableSource` in `app.js` — currently just
   AnimeFire).

### Player

Two elements exist inside `#player-container`, only one visible at a
time depending on the source of the resolved stream URL:

- **`<video>`**: used for direct file URLs (e.g. AnimeFire's
  `lightspeedst.net` links). Points to `/api/stream/proxy?...`
  (server-side proxy) instead of the CDN URL directly — required
  because these CDNs need a specific `Referer` a browser can't forge
  cross-origin (see `backend/README.md`).
- **`<iframe>`**: used when the resolved URL is a Blogger embed
  (`blogger.com`/`*.blogspot.com`). Some AnimeFire episodes are hosted
  on Blogger; AnimeFire's own page embeds that URL as an `<iframe>`,
  not as a direct video file — a `<video src>` pointed at it just
  fetches an HTML page, not media, and fails to play. Confirmed by
  reading `pkg/goanime`'s scraper source
  (`extractAnimefireBloggerURL`). This path bypasses the proxy: it's a
  normal page load, not raw video bytes.

## Endpoints consumed

```
GET /api/search?q=...&source=...
GET /api/episodes?url=...&source=...
GET /api/stream?animeUrl=...&source=...&episodeUrl=...&episodeNumber=...
GET /api/stream/proxy?url=...&source=...&referer=...
```

## Design notes

- No bundler, no framework: kept intentionally simple, matching the
  project's incremental-scope approach.
- The player doesn't assume any stream is directly playable — this was
  validated empirically per source (see the "Status" section of the
  root README) before writing the `<video>`/`<iframe>` branching logic.
