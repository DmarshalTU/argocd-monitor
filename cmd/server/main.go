package main

import (
	"log"
	"net/http"

	"argocd-monitor/internal/handlers"
	"argocd-monitor/internal/services"
)

func main() {
	argocdService := services.NewArgocdService()
	dashboardHandler := handlers.NewDashboardHandler(argocdService)

	http.HandleFunc("/", dashboardHandler.HandleDashboard)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}