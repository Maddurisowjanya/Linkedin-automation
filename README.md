***LinkedIn Automation Proof-of-Concept (Go + Rod)***

CLI-based LinkedIn automation proof-of-concept built using Go and the Rod browser automation library.
The project demonstrates modular system design, browser automation, basic stealth techniques, and SQLite-based state persistence.


**Demo Video:**
https://drive.google.com/file/d/1p6GcHDxQ9iScupHQEMCiVsFN2kN3-ZO3/view?usp=sharing


**project demonstrates**

End-to-end browser automation using Go

Clean modular architecture

External configuration using YAML

Responsible automation with limits and persistence

Handling real-world platform restrictions (LinkedIn checkpoints)

**Project Structure**


```text
linkedin-automation/
├── cmd/
│   └── app/
│       └── main.go          # Application entry point
│
├── internal/
│   ├── auth/                # LinkedIn login & session handling
│   ├── browser/             # Browser initialization using Rod
│   ├── config/              # Configuration loader (YAML)
│   ├── search/              # Profile search logic
│   ├── connect/             # Connection request logic
│   ├── messaging/           # Follow-up messaging
│   ├── stealth/             # Human-like delays & scrolling
│   ├── storage/             # SQLite persistence
│   └── logger/              # Centralized logging
│
├── config.yaml              # Application configuration file
├── linkedin_poc.db          # SQLite database (auto-generated)
├── CHECKPOINT_GUIDE.md      # LinkedIn checkpoint help
├── TROUBLESHOOTING.md       # Common issues & fixes
├── go.mod                   # Go module file
├── go.sum                   # Dependency checksums
└── README.md


Running the Application
go run ./cmd/app
