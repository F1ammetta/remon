package main

import (
	"github.com/F1ammetta/remon/handlers"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize the service manager
	systemdManager := handlers.NewSystemdManager()

	// Create router
	r := mux.NewRouter()

	// Main page
	r.HandleFunc("/", systemdManager.IndexHandler).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/services", systemdManager.GetServicesHandler).Methods("GET")
	api.HandleFunc("/services/add", systemdManager.AddServiceHandler).Methods("POST")
	api.HandleFunc("/services/{name}/details", systemdManager.GetServiceDetailsHandler).Methods("GET")
	api.HandleFunc("/services/{name}/status", systemdManager.GetServiceStatusHandler).Methods("GET")
	api.HandleFunc("/services/{name}/start", systemdManager.StartServiceHandler).Methods("POST")
	api.HandleFunc("/services/{name}/stop", systemdManager.StopServiceHandler).Methods("POST")
	api.HandleFunc("/services/{name}/restart", systemdManager.RestartServiceHandler).Methods("POST")
	api.HandleFunc("/services/{name}/logs", systemdManager.GetLogsHandler).Methods("GET")

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
