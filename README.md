# Crochet by Juliette

A modern, containerized Go web application for showcasing and managing handmade crochet items. Built with Go, SQLite, and a touch of frontend magic.

## Features

-   **Public Shop:** Beautiful responsive grid layout with "Hero" section and "Glassmorphism" design.
-   **Order System:** Customers can request orders with quantities and notes.
-   **Order Tracking:** Magic Link system for customers to view their order status (Ordered -> Shipped -> Delivered) without passwords.
-   **Admin Dashboard:** Secure area to manage items (CRUD) and update order statuses.
-   **Notifications:** Toast notifications for user feedback.
-   **Security:** CSRF protection, secure sessions, and bcrypt password hashing.

## Tech Stack

-   **Backend:** Go (Golang) 1.23+
-   **Database:** SQLite (Pure Go, CGO-free via `modernc.org/sqlite`)
-   **Frontend:** HTML5, CSS3, Vanilla JS (No heavy frameworks)
-   **Build Tool:** Task (Taskfile)
-   **Containerization:** Docker / Podman

## Prerequisites

-   Go 1.23+
-   [Task](https://taskfile.dev/) (Optional, but recommended for build scripts)
-   Podman or Docker

## Setup (Local Development)

1.  **Initialize Dependencies:**
    ```bash
    go mod tidy
    ```

2.  **Create an Admin User:**
    Use the CLI tool to create your first admin account.
    ```bash
    go run cmd/cli/main.go add-user -username admin -password mysecretpassword
    ```

3.  **Run the Server:**
    ```bash
    # Default port is 8585
    go run cmd/server/main.go
    ```

4.  **Access the App:**
    -   **Shop:** [http://localhost:8585](http://localhost:8585)
    -   **Admin Login:** [http://localhost:8585/login](http://localhost:8585/login)

## Configuration

The application is configured via environment variables.

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` | HTTP Port to listen on | `8585` |
| `DB_PATH` | Path to SQLite database file | `./crochet.db` |
| `CSRF_KEY` | 32-byte base64 string for CSRF protection | *(Randomly generated on start if unset)* |
| `SESSION_KEY` | 32-byte base64 string for session encryption | *(Randomly generated on start if unset)* |
| `COOKIE_SECURE`| Set to `true` if running behind HTTPS | `false` |
| `COOKIE_DOMAIN`| Domain for cookies (e.g., `example.com`) | *(empty)* |

**Security Note:** For production, you **MUST** set `CSRF_KEY` and `SESSION_KEY` to persist sessions across restarts.

## Build & Deployment

This project uses `Taskfile` for automation.

### Build Binaries
```bash
task build
```
Creates `crochet-server` and `crochet-cli` in the root.

### Build Container Image (Podman)
```bash
task image
```

### Deploy to Kubernetes
See [K8S_CLOUDFLARE_GUIDE.md](K8S_CLOUDFLARE_GUIDE.md) (local doc) for details on deploying with Flux and Cloudflare Tunnels.

Basic usage:
```bash
kubectl apply -f k8s/manifests.yaml
```

## Project Structure

-   `cmd/`: Entry points (`server` for web app, `cli` for management tools).
-   `internal/`: Private application code.
    -   `handlers/`: HTTP controllers.
    -   `store/`: Database access layer.
    -   `models/`: Data structures.
-   `templates/`: HTML templates.
-   `static/`: Assets (CSS, JS, Images).
-   `migrations/`: SQL schema migrations.