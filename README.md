## LinkedIn Automation Proof‑of‑Concept (Go + Rod)

This repository contains a **LinkedIn automation proof‑of‑concept** built in Go using the [Rod](https://go-rod.github.io/) browser automation library.  
It demonstrates **clean modular architecture**, **stealthy browser automation techniques**, and **SQLite‑backed state persistence**.

> **Disclaimer**  
> This project is for **educational and evaluation purposes only**.  
> Do **not** use it to violate LinkedIn’s Terms of Service, rate limits, or any applicable law.  
> You are solely responsible for how you use this code.

---

### Project Overview

- **Language / Runtime**: Go (with Go modules)
- **Browser Automation**: Rod + Chrome/Chromium
- **Persistence**: SQLite
- **Configuration**: YAML with environment overrides
- **Logging**: Structured logging via Logrus

The PoC performs a minimal end‑to‑end flow:

1. Start a **stealth‑configured Rod browser**.
2. **Log in to LinkedIn** (using credentials from environment variables, with session cookie reuse).
3. Perform **keyword‑based people search**.
4. Send **connection requests** with a personalised note (with simple daily limits).
5. Send **follow‑up messages** to connections (template‑driven, rate‑limited).

All *outgoing actions* (requests/messages) are recorded to SQLite so the PoC can **enforce limits** and **resume after crashes**.

---

### Architecture

High‑level structure:

```text
cmd/app/main.go               – entry point, wiring and demo flow

internal/browser              – Rod browser bootstrap + stealth
internal/config               – YAML config, env overrides, validation/defaults
internal/logger               – shared Logrus logger
internal/storage              – SQLite state persistence

internal/auth                 – LinkedIn login + cookie management
internal/search               – keyword search and profile URL extraction
internal/connect              – connection request workflow
internal/messaging            – follow‑up messaging workflow

internal/stealth/mouse.go     – human‑like mouse movement
internal/stealth/typing.go    – human‑like typing
internal/stealth/scroll.go    – human‑like scrolling behaviour
internal/stealth/timing.go    – think‑time / random delays
```

**Key design principles**

- **Internal packages**: All implementation is under `internal/` to keep the public surface small.
- **Small, focused packages**: Each concern (browser, auth, search, connect, messaging, storage, stealth) has its own package.
- **Dependency injection**: Core dependencies (`*rod.Browser`, `*storage.Storage`, config, logger) are constructed in `cmd/app/main.go` and passed into internal packages.
- **Idempotent operations**: Storage layer checks and deduplicates actions (e.g. one connection request per profile URL).

---

### Stealth Techniques Implemented

The **browser initializer** in `internal/browser/browser.go` configures Rod with several stealth features, each documented in code comments:

- **Randomised User‑Agent**
  - Chooses from a small pool of modern desktop Chrome user‑agents on each run.
  - Avoids a static, easy‑to‑fingerprint user‑agent string.

- **Custom Viewport Size**
  - Sets `window-size` based on config (default: 1366×768).
  - Mimics realistic resolutions instead of obvious defaults.

- **Headless Mode (Configurable)**
  - `browser.headless` in `config.yaml` toggles headless/headful mode.
  - Running headful can sometimes be less suspicious; both are supported.

- **Navigator WebDriver Masking**
  - Uses JS injection to set `navigator.webdriver` to `undefined`.
  - Some detection scripts explicitly check this flag.

- **Additional Automation Flag Masking**
  - Injects JS to provide non‑empty `navigator.plugins` and a plausible `navigator.languages` value.
  - Combined with Rod’s `stealth.Page`, this helps blend in with real browsers.

- **Human‑Like Interaction Helpers (internal/stealth)**
  - **Mouse** (`mouse.go`): cubic Bézier paths, variable velocity, overshoots and micro‑corrections.
  - **Typing** (`typing.go`): variable key delays, occasional typo/backspace cycles, natural rhythm.
  - **Scrolling** (`scroll.go`): random distances, easing‑based acceleration/deceleration, scroll‑back.
  - **Timing** (`timing.go`): random think‑time delays between actions.

> **Note**: These techniques **do not guarantee undetectability**; they only demonstrate how to implement more realistic automation.

---

### Configuration Management

Configuration is defined in `config.yaml` and loaded/validated in `internal/config/config.go`.

- **YAML file**: `config.yaml` at the project root.
- **Environment overrides**:
  - `SQLITE_DSN` overrides the database DSN.
  - `LOG_LEVEL` controls log verbosity.
- **Defaults & validation**:
  - Safe defaults for viewport, limits, delays, and DSN.
  - Basic validation for fields like search keywords and limits.

**Example `config.yaml` (included):**

```yaml
browser:
  headless: true
  viewport_width: 1366
  viewport_height: 768
  user_agent: ""

database:
  dsn: "file:linkedin_poc.db?_fk=1"

search:
  keywords:
    - "golang developer"
    - "backend engineer"
  max_pages: 1
  page_delay_min: 2s
  page_delay_max: 5s

connect:
  daily_limit: 10
  note_template: "Hi there, I found your profile ({{PROFILE_URL}}) while experimenting with a LinkedIn automation proof-of-concept and wanted to connect."
  action_delay_min: 2s
  action_delay_max: 5s

messaging:
  templates:
    - "Thanks for connecting! This message was sent as part of an educational LinkedIn automation proof-of-concept."
    - "Great to connect with you. I'm currently evaluating a browser automation PoC and your profile was included in a demo search."
  check_interval: 10m
  daily_limit: 10
  action_delay_min: 2s
  action_delay_max: 5s
```

---

### State Persistence & Rate Limiting

`internal/storage/sqlite.go` implements a light‑weight persistence layer on top of SQLite:

- **Sent requests table**
  - Stores `profile_url` and `sent_at`.
  - `UNIQUE(profile_url)` ensures one request per profile.
  - Used to **avoid duplicates** and **enforce daily limits**.

- **Messages table**
  - Stores `profile_url`, `message_type` (e.g. `"followup"`), and `sent_at`.
  - Used to **rate‑limit follow‑up messages** and avoid multiple messages per user.

- **Rate‑limit helpers**
  - `CountRequestsSince` and `CountMessagesSince` are used by the `connect` and `messaging` packages to enforce **daily caps** from config.

This site‑aware state allows the PoC to **resume safely after crashes** without re‑sending actions.

---

### Setup Instructions

#### 1. Prerequisites

- Go **1.21+**
- A recent **Chrome/Chromium** installation
- SQLite (for debugging / inspection, optional at runtime)

#### 2. Clone and install dependencies

```bash
git clone <this-repo-url> linkedin-automation-poc
cd linkedin-automation-poc
go mod tidy
```

#### 3. Configure environment variables

Set your **LinkedIn test account** credentials in the environment:

```bash
export LINKEDIN_EMAIL="your-email@example.com"
export LINKEDIN_PASSWORD="your-password"
```

> **Strongly recommended**: Use a **non‑production**, non‑critical test account only.

Optionally:

```bash
export SQLITE_DSN="file:linkedin_poc.db?_fk=1"
export LOG_LEVEL="debug"
```

#### 4. Adjust `config.yaml`

- Tune `search.keywords` to keep queries small and controlled.
- Adjust `connect.daily_limit` and `messaging.daily_limit` to **low** values.
- Decide whether to run in `browser.headless: true` or `false`.

---

### Running the Demo

From the project root:

```bash
go run ./cmd/app
```

What the demo does:

1. **Starts** a stealth‑configured Rod browser.
2. **Loads config** and connects to SQLite (creating tables if needed).
3. **Logs in** to LinkedIn:
   - Re‑uses cookies from `linkedin_session_cookies.json` when possible.
   - Falls back to full login if cookies are invalid or missing.
4. Performs **keyword search** and collects unique profile URLs.
5. Sends **connection requests** with a personalised note, **respecting daily limits** and skipping profiles already contacted.
6. Sends **template‑based follow‑up messages** to some connections, also rate‑limited.

You can safely terminate the process (Ctrl‑C); on the next run, the PoC will reuse **session cookies** and **SQLite state** where possible.

---

### Folder Structure

```text
.
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── auth/
│   │   └── login.go
│   ├── browser/
│   │   └── browser.go
│   ├── config/
│   │   └── config.go
│   ├── connect/
│   │   └── connect.go
│   ├── logger/
│   │   └── logger.go
│   ├── messaging/
│   │   └── messaging.go
│   ├── search/
│   │   └── search.go
│   ├── stealth/
│   │   ├── mouse.go
│   │   ├── typing.go
│   │   ├── scroll.go
│   │   └── timing.go
│   └── storage/
│       └── sqlite.go
├── config.yaml
├── go.mod
└── README.md
```

---

### Evaluation Alignment

This PoC is designed to be **readable** and **easy to audit** for educational and evaluation purposes:

- **Rod usage** is straightforward and isolated in `internal/browser` and feature packages.
- **Stealth behaviours** are clearly explained with in‑line comments and separated in `internal/stealth`.
- **LinkedIn workflows** (login, search, connect, messaging) are minimal, heuristic, and non‑abusive by default.
- **State persistence and limits** are explicit via the SQLite schema and config.

If you are evaluating browser automation patterns, focus on:

- How **Rod** is initialised and configured.
- How **stealth techniques** are layered (browser‑level vs. interaction‑level).
- How **state and limits** are enforced to avoid repeated or excessive actions.



