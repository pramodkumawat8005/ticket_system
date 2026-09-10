package main

import (
    "net/http"
    "strings"
)

type createTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}


func (app *App) handleCreateTicket(w http.ResponseWriter, r *http.Request) {
	var req createTicketRequest
	if err := decodeJSONBody(r, &req); 
	err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	userID := userIDFromContext(r)
	ticket := app.store.CreateTicket(userID, req.Title, req.Description)
	writeJSON(w, http.StatusCreated, ticket)
}


func (app *App) handleListTickets(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromContext(r)
	tickets := app.store.ListTicketsForUser(userID)
	writeJSON(w, http.StatusOK, tickets)
}


func (app *App) handleGetTicket(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := userIDFromContext(r)

	ticket, ok := app.store.GetTicket(id)
	if !ok {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if ticket.UserID != userID {
		writeError(w, http.StatusForbidden, "you do not have access to this ticket")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}


func (app *App) handleUpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	userID := userIDFromContext(r)

	var req updateStatusRequest
	if err := decodeJSONBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !IsValidStatus(req.Status) {
		writeError(w, http.StatusBadRequest, "status must be one of: open, in_progress, closed")
		return
	}

	ticket, ok := app.store.GetTicket(id)
	if !ok {
		writeError(w, http.StatusNotFound, "ticket not found")
		return
	}
	if ticket.UserID != userID {
		writeError(w, http.StatusForbidden, "you do not have access to this ticket")
		return
	}

	newStatus := TicketStatus(req.Status)
	if !CanTransition(ticket.Status, newStatus) {
		writeError(w, http.StatusBadRequest, "invalid status transition from "+string(ticket.Status)+" to "+req.Status)
		return
	}

	updated, _ := app.store.UpdateTicketStatus(id, newStatus)
	writeJSON(w, http.StatusOK, updated)
}

