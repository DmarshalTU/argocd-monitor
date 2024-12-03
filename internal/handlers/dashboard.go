package handlers

import (
	"html/template"
	"log"
	"net/http"

	"argocd-monitor/internal/models"
	"argocd-monitor/internal/services"
)

type DashboardHandler struct {
	argocdService *services.ArgocdService
	template      *template.Template
}

func NewDashboardHandler(argocdService *services.ArgocdService) *DashboardHandler {
	tmpl := template.Must(template.ParseFiles("web/templates/dashboard.html"))
	return &DashboardHandler{
		argocdService: argocdService,
		template:      tmpl,
	}
}

func (h *DashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	apps, err := h.argocdService.GetApplications()
	if err != nil {
		log.Printf("Error getting applications: %v", err)
		http.Error(w, "Failed to get applications", http.StatusInternalServerError)
		return
	}

	allSynced := true
	for _, app := range apps {
		if app.SyncStatus != "Synced" {
			allSynced = false
			break
		}
	}

	data := models.DashboardData{
		Applications: apps,
		AllSynced:    allSynced,
	}

	if err := h.template.Execute(w, data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}