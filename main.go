package main

import (
	"log"
	"net/http"
	"os"
	"time"
)


type App struct {
	store     *Store
	jwtSecret []byte
	tokenTTL  time.Duration
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
		log.Println("WARNING: JWT_SECRET not set, using an insecure default. Set JWT_SECRET in production.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	app := &App{
		store:     NewStore(),
		jwtSecret: []byte(secret),
		tokenTTL:  24 * time.Minute,
	}

	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("POST /auth/register", app.handleRegister)
	mux.HandleFunc("POST /auth/login", app.handleLogin)

	// Protected endpoints (require a valid JWT)
	mux.HandleFunc("POST /tickets", app.requireAuth(app.handleCreateTicket))
	mux.HandleFunc("GET /tickets", app.requireAuth(app.handleListTickets))
	mux.HandleFunc("GET /tickets/{id}", app.requireAuth(app.handleGetTicket))
	mux.HandleFunc("PATCH /tickets/{id}/status", app.requireAuth(app.handleUpdateTicketStatus))

	addr := ":" + port
	log.Printf("ticket-system listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
