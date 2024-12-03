package models

type Application struct {
	Name        string `json:"name"`
	SyncStatus  string `json:"sync_status"`
	Health      string `json:"health_status"`
	Namespace   string `json:"namespace"`
	Project     string `json:"project"`
}

type DashboardData struct {
	Applications []Application
	AllSynced    bool
}