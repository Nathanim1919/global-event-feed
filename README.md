# Global Event Feed

Global Event Feed is a deployable, real-time event system designed to ingest, process, and distribute events globally. It demonstrates production-ready Go backend architecture, system design principles, and minimal frontend integration for testing and visualization.

## Features

- Ingest events via REST API (`POST /events`)
- Retrieve events via REST API (`GET /events`)
- Store events in PostgreSQL
- Minimal frontend to submit and view events
- Production-ready folder structure and configuration

## Tech Stack

- **Backend:** Go (chi router, structured handlers/services/repositories)
- **Database:** PostgreSQL
- **Real-time:** WebSocket support
- **Frontend:** Minimal HTML/JS for testing

## Project Goals

1. Learn and apply Go in a production-style project.
2. Implement real system design patterns: separation of concerns, async processing, and eventual scaling.
3. Build a deployable system that demonstrates real-time event handling.
4. Create a clean, portfolio-ready project that showcases technical skill.

## Getting Started

1. Clone the repo
2. Set up PostgreSQL and configure `.env`
3. Run `go run ./cmd/main.go` to start the server
4. Open `index.html` in the browser to test event submission

---

**Note:** This is Phase 1 (MVP) — basic event ingestion and retrieval. Later phases will include async processing, worker pools, and real-time WebSocket updates.
