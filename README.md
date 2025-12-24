#LinkedIn Automation Proof-of-Concept (Go + Rod)

CLI-based LinkedIn automation proof-of-concept built using Go and the Rod browser automation library.
The project demonstrates modular system design, browser automation, basic stealth techniques, and SQLite-based state persistence.

Disclaimer
This project is created only for educational and evaluation purposes.
Do not use it to violate LinkedIn’s Terms of Service or platform policies.

Demo Video:
https://drive.google.com/file/d/1p6GcHDxQ9iScupHQEMCiVsFN2kN3-ZO3/view?usp=sharing


##project demonstrates

End-to-end browser automation using Go

Clean modular architecture

External configuration using YAML

Responsible automation with limits and persistence

Handling real-world platform restrictions (LinkedIn checkpoints)

##Project Structure
linkedin-automation/
├── cmd/
│   └── app/
│       └── main.go          # Application entry point
│
├── internal/
│   ├── auth/                # LinkedIn login & session handling
│   ├── browser/             # Rod browser setup (headless/headful)
│   ├── config/              # Config loader and validation
│   ├── connect/             # Connection request workflow
│   ├── logger/              # Centralized structured logging
│   ├── messaging/           # Follow-up messaging logic
│   ├── search/              # Keyword-based profile search
│   ├── stealth/             # Human-like delays, typing, scrolling
│   └── storage/             # SQLite persistence layer
│
├── config.yaml              # YAML configuration file
├── linkedin_poc.db          # SQLite database (auto-generated)
├── CHECKPOINT_GUIDE.md      # Handling LinkedIn checkpoints
├── TROUBLESHOOTING.md       # Common issues and fixes
├── go.mod                   # Go module dependencies
├── go.sum                   # Dependency checksums
└── README.md


Running the Application
go run ./cmd/app
