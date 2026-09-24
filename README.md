# ShopKart CTF — win CyberRange Pilot (Local Build)

A small, self-contained, intentionally-vulnerable e-commerce demo ("ShopKart")
with a built-in 5-flag Capture-The-Flag layer ("Mission Control"), built for
the Foundation-level VA course pilot.

This build is for **local, offline demo purposes only** — nothing here talks
to the internet or to any win upskill production system. It's meant to be
played on your own laptop so the mentor can review it before anything goes
near Plesk / production.

---

## 1. Requirements

- **Go 1.22 or newer.** Check with `go version`. If you don't have Go, install
  it from https://go.dev/dl/ (pick the installer for your OS — it's a
  2-minute install, no extra setup needed).
- A C compiler (needed for the SQLite database driver). Windows/Mac users:
  if `go build` complains about a missing compiler, install
  [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) (Windows) or Xcode Command
  Line Tools (`xcode-select --install` on Mac). Linux: `gcc` is usually
  already present.
- A modern browser (Chrome/Edge/Firefox).

No Docker, no internet connection, and no database server needed — everything
runs from a single Go process with an in-memory database.

## 2. How to run it

```bash
cd shopkart-ctf
go run .
```

You should see:

```
ShopKart CTF running: http://localhost:8080
```

Open **http://localhost:8080** in your browser. That's it.

To stop the server, press `Ctrl+C` in the terminal. Every time you restart
it, all progress and review data resets to a clean state — this is
deliberate, so every playthrough starts fresh.

## 3. How to play

1. Land on the ShopKart storefront — browse it like a normal shopper first.
2. Click the amber **"Mission Control"** button (top right) to enter the CTF
   layer.
3. Read the case briefing, then work through the 5 challenge cards in order
   (they don't have to be done in order, but they get progressively less
   guided).
4. Each card has:
   - A **briefing** (the story/context)
   - A **link to where to go** on the ShopKart site (or a note that you'll
     need to find the spot yourself)
   - An **"Ask the Security Consultant for a hint"** button — click it as
     many times as you want, each click reveals one more hint
   - A **flag submission box** — paste in the flag once you find it
5. Capturing a flag gives you an instant confetti burst and a short
   "why this matters" explanation. Capture all 5 to unlock the mission
   completion screen.

## 4. What's actually vulnerable (for the facilitator)

Every bug below is seeded **on purpose** for this training build — see the
`// VULNERABLE ON PURPOSE` comments in `handlers.go` for exactly where and
why. See `SOLUTIONS.md` for the answers/walkthrough if you want to check
your own play-through or skip ahead.

| # | Challenge | Vulnerability class | Where |
|---|---|---|---|
| 1 | The Front Door | SQL Injection | Login form |
| 2 | The Spare Key | Broken Authentication | Hidden text on the About page |
| 3 | The Poisoned Review | Stored XSS | Product review box |
| 4 | The Leaky API | Sensitive Data Exposure | Unauthenticated `/api/users` |
| 5 | The Backdoor Left Open | Security Misconfiguration | Exposed `/.env` file |

## 5. Project structure

```
shopkart-ctf/
  main.go        - server bootstrap & routing
  db.go           - in-memory SQLite schema + seed data
  flags.go        - the 5 challenges: story text, hints, flag answers
  handlers.go     - all HTTP handlers (this is where every bug lives)
  session.go      - cookie-based per-player progress tracking
  templates/      - all HTML (Go html/template)
  static/         - style.css + app.js (Mission Control interactivity, confetti)
```

Nothing here uses a framework — plain Go standard library + one dependency
(`github.com/mattn/go-sqlite3`) for the database. Easy to read end-to-end in
one sitting.

## 6. What happens after this is approved

This is deliberately a **local-only** build. Once this is reviewed and
approved, the next steps (not done yet) are:

1. Package this into a Docker image (no code changes needed - it already
   runs as a single self-contained binary).
2. Deploy that image inside an isolated Plesk subscription/subdomain
   (e.g. `lab.winupskill.in`), fully separated from the production
   winupskill.in codebase and database.
3. Connect flag-capture completion to win's own course/certificate system.

None of that has been touched yet — this build only exists on this machine.
