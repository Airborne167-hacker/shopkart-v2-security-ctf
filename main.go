package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	initDB()
	loadTemplates()

	mux := http.NewServeMux()

	// storefront
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/about", aboutHandler)
	mux.HandleFunc("/product/", productHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/account", accountHandler)

	// deliberately vulnerable / exposed surfaces
	mux.HandleFunc("/api/users", apiUsersHandler)
	mux.HandleFunc("/.env", envFileHandler)

	// Mission Control (the CTF layer)
	mux.HandleFunc("/mission", missionHandler)
	mux.HandleFunc("/api/hint", hintHandler)
	mux.HandleFunc("/api/submit-flag", submitFlagHandler)
	mux.HandleFunc("/api/progress", progressHandler)
	mux.HandleFunc("/api/reset", resetProgressHandler)

	// static assets
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	port := os.Getenv("PORT")
	if port == "" {
		port = "10001"
	}
	addr := ":" + port
	log.Printf("ShopKart CTF running: http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
