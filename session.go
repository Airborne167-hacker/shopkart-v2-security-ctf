package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const sessionCookieName = "sk_session"

// PlayerState tracks one player's progress through Mission Control, and
// (separately) which site account they're currently "logged in" as within
// the ShopKart storefront story - these are deliberately kept apart so that
// capturing flags never depends on real authentication state.
type PlayerState struct {
	Captured    map[string]bool // challenge ID -> captured
	HintsShown  map[string]int  // challenge ID -> number of hints revealed
	LoggedInAs  string          // username, empty if not logged in
	DisplayName string
	Role        string
}

var (
	sessionsMu sync.Mutex
	sessions   = map[string]*PlayerState{}
)

func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// getOrCreateSession reads the session cookie (or creates one) and returns
// the player's state, creating a fresh one on first visit.
func getOrCreateSession(w http.ResponseWriter, r *http.Request) *PlayerState {
	var id string
	if c, err := r.Cookie(sessionCookieName); err == nil && c.Value != "" {
		id = c.Value
	} else {
		id = newSessionID()
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    id,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Expires:  time.Now().Add(7 * 24 * time.Hour),
		})
	}

	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	st, ok := sessions[id]
	if !ok {
		st = &PlayerState{
			Captured:   map[string]bool{},
			HintsShown: map[string]int{},
		}
		sessions[id] = st
	}
	return st
}

func (p *PlayerState) capturedCount() int {
	n := 0
	for _, v := range p.Captured {
		if v {
			n++
		}
	}
	return n
}
