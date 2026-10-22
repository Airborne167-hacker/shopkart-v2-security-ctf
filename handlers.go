package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

var tmpl *template.Template

func loadTemplates() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}

type Product struct {
	ID          int
	Name        string
	Price       string
	Tagline     string
	Description string
	Emoji       string
}

type Review struct {
	ID        int
	ProductID int
	Author    string
	Content   template.HTML // deliberately unescaped - this IS the XSS bug
	CreatedAt string
}

// ---------- storefront ----------

func homeHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	rows, err := db.Query(`SELECT id, name, price, tagline, description, emoji FROM products ORDER BY id`)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	defer rows.Close()
	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Tagline, &p.Description, &p.Emoji); err == nil {
			products = append(products, p)
		}
	}
	data := map[string]any{
		"Products": products,
		"Session":  sess,
	}
	if err := tmpl.ExecuteTemplate(w, "home", data); err != nil {
		log.Println("template error:", err)
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	_ = tmpl.ExecuteTemplate(w, "about", map[string]any{"Session": sess})
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	idStr := strings.TrimPrefix(r.URL.Path, "/product/")
	var p Product
	err := db.QueryRow(`SELECT id, name, price, tagline, description, emoji FROM products WHERE id = ?`, idStr).
		Scan(&p.ID, &p.Name, &p.Price, &p.Tagline, &p.Description, &p.Emoji)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		author := r.FormValue("author")
		content := r.FormValue("content")
		if author == "" {
			author = "Anonymous shopper"
		}
		if strings.TrimSpace(content) != "" {
			// VULNERABLE ON PURPOSE: review content is stored and later
			// rendered as raw HTML (see below) with no sanitisation and no
			// output-encoding. This is a textbook Stored XSS bug, seeded
			// deliberately for the "Poisoned Review" challenge.
			_, err := db.Exec(
				`INSERT INTO reviews (product_id, author, content, created_at) VALUES (?, ?, ?, ?)`,
				p.ID, author, content, "just now",
			)
			if err != nil {
				log.Println("review insert error:", err)
			}
		}
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}

	rows, err := db.Query(`SELECT id, product_id, author, content, created_at FROM reviews WHERE product_id = ? ORDER BY id DESC`, p.ID)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	defer rows.Close()
	var reviews []Review
	for rows.Next() {
		var rv Review
		var content string
		if err := rows.Scan(&rv.ID, &rv.ProductID, &rv.Author, &content, &rv.CreatedAt); err == nil {
			rv.Content = template.HTML(content) // raw, unescaped - intentional bug
			reviews = append(reviews, rv)
		}
	}

	data := map[string]any{
		"Product": p,
		"Reviews": reviews,
		"Session": sess,
		// A fake "secret" exposed to page JS, standing in for something an
		// attacker would actually want in a real XSS attack (a session
		// token, an auth cookie, etc). Grabbing this via injected JS is the
		// whole point of the "Poisoned Review" challenge.
		"Secret": challengeByID("f3").Flag,
	}
	_ = tmpl.ExecuteTemplate(w, "product", data)
}

// ---------- login / account (SQL injection + broken auth live here) ----------

func loginHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		// VULNERABLE ON PURPOSE: the query is built with raw string
		// concatenation instead of a parameterised query / prepared
		// statement. This is a textbook SQL Injection bug, seeded
		// deliberately for the "Front Door" challenge. Never build SQL like
		// this in a real application - always use placeholders (see every
		// OTHER query in this codebase for the correct pattern).
		query := fmt.Sprintf(
			"SELECT username, role, display_name FROM users WHERE username = '%s' AND password = '%s'",
			username, password,
		)

		row := db.QueryRow(query)
		var uname, role, display string
		err := row.Scan(&uname, &role, &display)
		if err != nil {
			data := map[string]any{"Session": sess, "Error": "Invalid username or password."}
			_ = tmpl.ExecuteTemplate(w, "login", data)
			return
		}

		sessionsMu.Lock()
		sess.LoggedInAs = uname
		sess.Role = role
		sess.DisplayName = display
		sessionsMu.Unlock()

		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}

	_ = tmpl.ExecuteTemplate(w, "login", map[string]any{"Session": sess})
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	sessionsMu.Lock()
	sess.LoggedInAs = ""
	sess.Role = ""
	sess.DisplayName = ""
	sessionsMu.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func accountHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	if sess.LoggedInAs == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	var flag *Challenge
	switch sess.Role {
	case "admin":
		flag = challengeByID("f1")
	case "staff":
		flag = challengeByID("f2")
	}
	_ = tmpl.ExecuteTemplate(w, "account", map[string]any{"Session": sess, "Flag": flag})
}

// ---------- deliberately exposed "backend" surfaces ----------

// apiUsersHandler is an unauthenticated endpoint that dumps every user
// row, including the internal "notes" field. VULNERABLE ON PURPOSE:
// Sensitive Data Exposure, seeded for the "Leaky API" challenge.
func apiUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT username, role, display_name, notes FROM users`)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	defer rows.Close()

	type userOut struct {
		Username    string `json:"username"`
		Role        string `json:"role"`
		DisplayName string `json:"display_name"`
		Notes       string `json:"internal_notes,omitempty"`
	}
	var out []userOut
	for rows.Next() {
		var u userOut
		if err := rows.Scan(&u.Username, &u.Role, &u.DisplayName, &u.Notes); err == nil {
			out = append(out, u)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// envFileHandler serves a fake .env file from the web root. VULNERABLE ON
// PURPOSE: Security Misconfiguration, seeded for the "Backdoor Left Open"
// challenge.
func envFileHandler(w http.ResponseWriter, r *http.Request) {
	body := `# ShopKart production environment
APP_ENV=production
APP_DEBUG=false
DB_HOST=127.0.0.1
DB_NAME=shopkart_prod
DB_USER=shopkart_svc
DB_PASS=CorrectHorseBatteryStaple99!
MAIL_DRIVER=smtp
MAIL_HOST=smtp.shopkart.example
# TODO(before launch): rotate this - it's been in here since staging
DEBUG_BACKDOOR_TOKEN=` + challengeByID("f5").Flag + `
`
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(body))
}

// ---------- Mission Control (the CTF layer) ----------

func missionHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	sessionsMu.Lock()
	capturedCopy := make(map[string]bool, len(sess.Captured))
	for k, v := range sess.Captured {
		capturedCopy[k] = v
	}
	sessionsMu.Unlock()

	data := map[string]any{
		"Session":     sess,
		"Challenges":  challenges,
		"Captured":    sess.capturedCount(),
		"Total":       len(challenges),
		"CapturedMap": capturedCopy,
	}
	_ = tmpl.ExecuteTemplate(w, "mission", data)
}

type hintRequest struct {
	ID string `json:"id"`
}
type hintResponse struct {
	Hints []string `json:"hints"`
	Done  bool     `json:"done"`
}

func hintHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	var req hintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	ch := challengeByID(req.ID)
	if ch == nil {
		http.Error(w, "unknown challenge", 404)
		return
	}
	sessionsMu.Lock()
	shown := sess.HintsShown[req.ID]
	if shown < len(ch.Hints) {
		shown++
		sess.HintsShown[req.ID] = shown
	}
	sessionsMu.Unlock()

	resp := hintResponse{Hints: ch.Hints[:shown], Done: shown >= len(ch.Hints)}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type submitRequest struct {
	ID   string `json:"id"`
	Flag string `json:"flag"`
}
type submitResponse struct {
	Correct   bool   `json:"correct"`
	Message   string `json:"message"`
	Reveal    string `json:"reveal,omitempty"`
	Captured  int    `json:"captured"`
	Total     int    `json:"total"`
	AllDone   bool   `json:"all_done"`
	Challenge string `json:"challenge_id"`
}

func submitFlagHandler(w http.ResponseWriter, r *http.Request) {
	sess := getOrCreateSession(w, r)
	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", 400)
		return
	}
	ch := challengeByID(req.ID)
	if ch == nil {
		http.Error(w, "unknown challenge", 404)
		return
	}

	submitted := strings.TrimSpace(req.Flag)
	resp := submitResponse{Challenge: req.ID}

	if submitted == ch.Flag {
		sessionsMu.Lock()
		alreadyHad := sess.Captured[req.ID]
		sess.Captured[req.ID] = true
		captured := sess.capturedCount()
		sessionsMu.Unlock()

		resp.Correct = true
		resp.Captured = captured
		resp.Total = len(challenges)
		resp.AllDone = captured == len(challenges)
		if alreadyHad {
			resp.Message = "Already logged in the case file - nice work all the same."
		} else {
			resp.Message = "Correct! Evidence logged."
			resp.Reveal = ch.Reveal
		}
	} else {
		resp.Correct = false
		resp.Message = "That's not it. Check the hints if you're stuck."
		sessionsMu.Lock()
		resp.Captured = sess.capturedCount()
		resp.Total = len(challenges)
		sessionsMu.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func progressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	sess := getOrCreateSession(w, r)
	sessionsMu.Lock()
	defer sessionsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"captured":    sess.Captured,
		"hints_shown": sess.HintsShown,
		"total":       len(challenges),
		"logged_in_as": sess.LoggedInAs,
		"role":         sess.Role,
		"server_time":  time.Now().Format(time.RFC3339),
	})
}

func resetProgressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	sess := getOrCreateSession(w, r)
	sessionsMu.Lock()
	sess.Captured = map[string]bool{}
	sess.HintsShown = map[string]int{}
	sessionsMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":   "ok",
		"message":  "Progress reset to zero",
		"captured": sess.Captured,
	})
}
