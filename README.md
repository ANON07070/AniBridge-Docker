<p align="center">
  <img src="docs/logo.png" alt="AniBridge logo" width="480">
</p>

<p align="center">
  <a href="README.pt-BR.md">🇧🇷 Português</a> ·
  <b>🇺🇸 English</b>
</p>

# AniBridge

[![Go](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Python](https://img.shields.io/badge/Python-3.10%2B-3776AB?style=flat&logo=python&logoColor=white)](https://www.python.org/)
[![FastAPI](https://img.shields.io/badge/FastAPI-009688?style=flat&logo=fastapi&logoColor=white)](https://fastapi.tiangolo.com/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?style=flat&logo=docker&logoColor=white)](https://www.docker.com/)

A self-hosted web app for watching anime through your browser. Search,
browse episode lists, and stream — no client-side software required
beyond a modern browser.

## Architecture

```
Browser → Frontend (HTML/CSS/JS) → FastAPI → Go Bridge → pkg/goanime
```

- **Frontend**: static, served by FastAPI itself (no CORS to configure).
- **Backend**: FastAPI, exposes the API and calls the Go Bridge via
  subprocess.
- **Go Bridge**: adapts [`pkg/goanime`](https://github.com/alvarorichard/Goanime)
  (a Go scraping library) into a small CLI that speaks JSON, since the
  backend is written in Python.

More detail on each layer: [`bridge/README.md`](bridge/README.md),
[`backend/README.md`](backend/README.md),
[`frontend/README.md`](frontend/README.md).

## Status

End-to-end flow working for **AnimeFire**: search → episode list →
in-browser playback, including automatic best-quality selection and
remote (non-localhost) access.

**Known limitations:**
- **Goyabu and SuperFlix**: show up in search results, but have no
  episode/stream support — `pkg/goanime` v1.8.6 doesn't expose these as
  values of the public `types.Source` type, so there's no way to fetch
  episodes for them in this version.
- **AllAnime**: search and episode listing work, but fetching the
  stream URL currently fails (timeout reaching a required referer page,
  `mkissa.to`). Deprioritized for now.
- **AnimeFire**: fully working, including a proxy layer that fetches
  video bytes server-side (some CDNs require a specific `Referer` that
  a browser can't forge on its own).

Project currently paused for beta testing.

## Getting Started

### Docker — prebuilt image (recommended, nothing to compile)

The image is built automatically by GitHub Actions, so it doesn't
matter how weak your machine is — nothing is compiled locally
(compiling the Go Bridge locally has caused out-of-memory build
failures on machines with ~4GB of RAM):

```bash
docker pull ghcr.io/YOUR_USERNAME/anibridge:latest
docker run -p 8000:8000 ghcr.io/YOUR_USERNAME/anibridge:latest
```

(replace `YOUR_USERNAME` with the GitHub user/org this was published
under)

### Docker — build locally

```bash
docker compose up --build
```

**Warning**: this compiles the Go Bridge on your own machine. If your
machine has limited RAM, prefer the prebuilt image above.

### Windows, without Docker

```powershell
cd bridge
go build -o goanime-bridge.exe .

cd ..\backend
$env:BRIDGE_PATH = "FULL_PATH_TO\bridge\goanime-bridge.exe"
pip install -r requirements.txt
uvicorn main:app --reload --port 8000
```

### Linux, without Docker

```bash
./run.sh
```

The script compiles the bridge, creates a Python venv, installs
dependencies, and starts the server. Accepts optional `HOST`/`PORT`:

```bash
PORT=9000 ./run.sh
```

Requires Go 1.26.5+ and Python 3.10+ already installed.

### Accessing it

In every case above: `http://localhost:8000`. To access from another
device on the same network, start with `--host 0.0.0.0` (Windows) or
`HOST=0.0.0.0 ./run.sh` (Linux), then use the machine's local IP.

## Project Rules

- Use `pkg/goanime` directly, don't reimplement its scraping logic.
- Verify the library's source code before assuming any behavior — its
  documentation has been out of date more than once.
- Incremental scope: each new feature is validated before moving to
  the next.
- No Redis, PostgreSQL, or authentication for now — simplicity first.
- After publishing a new image: GHCR packages are **private** by
  default even with a public repository — mark it public manually
  under Settings → Packages the first time, or `docker pull` will fail
  with an authentication error for other people.

## Project Structure

```
anibridge/
├── .github/workflows/  # CI: builds and publishes the Docker image to GHCR
├── bridge/      # Go: adapts pkg/goanime into JSON over a CLI
├── backend/     # FastAPI: API + serves the frontend
├── frontend/    # static HTML/CSS/JS
├── Dockerfile
├── docker-compose.yml
└── run.sh       # brings everything up at once on Linux, without Docker
```

## Credits

Built on top of [`pkg/goanime`](https://github.com/alvarorichard/Goanime)
by [alvarorichard](https://github.com/alvarorichard).
