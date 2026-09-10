package main

import (
	"time"
	
)


type TicketStatus string

const (
	StatusOpen       TicketStatus = "open"
	StatusInProgress TicketStatus = "in_progress"
	StatusClosed     TicketStatus = "closed"
)

var validNextStatuses = map[TicketStatus]map[TicketStatus]bool{
	StatusOpen: {
		StatusInProgress: true,
		StatusClosed:     true,
	},
	StatusInProgress: {
		StatusClosed: true,
	},
	StatusClosed: {},
}


func IsValidStatus(s string) bool {
	switch TicketStatus(s) {
	case StatusOpen, StatusInProgress, StatusClosed:
		return true
	}
	return false
}


func CanTransition(from, to TicketStatus) bool {
	if from == to {
		return false 
	}
	next, ok := validNextStatuses[from]
	if !ok {
		return false
	}
	return next[to]
}


type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // never serialized in API responses
	CreatedAt    time.Time `json:"created_at"`
}


type Ticket struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Status      TicketStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
