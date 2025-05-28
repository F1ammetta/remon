package handlers

import (
	"bytes"
	"fmt"
	"github.com/F1ammetta/remon/templates"
	"github.com/F1ammetta/remon/types"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// SystemdManager handles systemd operations
type SystemdManager struct {
	serviceManager *ServiceManager
}

// NewSystemdManager creates a new systemd manager
func NewSystemdManager() *SystemdManager {
	return &SystemdManager{
		serviceManager: NewServiceManager(),
	}
}

// IndexHandler renders the main page
func (sm *SystemdManager) IndexHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.Index("server-01.example.com", true)
	component.Render(r.Context(), w)
}

// GetServicesHandler returns the services list as HTML
func (sm *SystemdManager) GetServicesHandler(w http.ResponseWriter, r *http.Request) {
	services, err := sm.getServiceStatuses()
	if err != nil {
		http.Error(w, "Failed to get service statuses", http.StatusInternalServerError)
		log.Printf("Error getting service statuses: %v", err)
		return
	}

	component := templates.ServicesList(services)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// GetServiceDetailsHandler returns detailed view of a specific service
func (sm *SystemdManager) GetServiceDetailsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	service, err := sm.getServiceStatus(serviceName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get service status: %v", err), http.StatusInternalServerError)
		return
	}

	component := templates.ServiceDetails(service)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// StartServiceHandler starts a systemd service
func (sm *SystemdManager) StartServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	if err := sm.executeSystemctlCommand("start", serviceName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start service: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Service %s started successfully", serviceName)

	// Return updated service details
	sm.GetServiceDetailsHandler(w, r)
}

// StopServiceHandler stops a systemd service
func (sm *SystemdManager) StopServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	if err := sm.executeSystemctlCommand("stop", serviceName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop service: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Service %s stopped successfully", serviceName)

	// Return updated service details
	sm.GetServiceDetailsHandler(w, r)
}

// RestartServiceHandler restarts a systemd service
func (sm *SystemdManager) RestartServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	if err := sm.executeSystemctlCommand("restart", serviceName); err != nil {
		http.Error(w, fmt.Sprintf("Failed to restart service: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Service %s restarted successfully", serviceName)

	// Return updated service details
	sm.GetServiceDetailsHandler(w, r)
}

// GetServiceStatusHandler returns just the status for a service (for refresh)
func (sm *SystemdManager) GetServiceStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	service, err := sm.getServiceStatus(serviceName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get service status: %v", err), http.StatusInternalServerError)
		return
	}

	component := templates.ServiceDetails(service)
	w.Header().Set("Content-Type", "text/html")
	component.Render(r.Context(), w)
}

// GetLogsHandler returns service logs
func (sm *SystemdManager) GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["name"]

	// Get query parameters
	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "100"
	}

	logs, err := sm.getServiceLogs(serviceName, lines)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get logs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(logs))
}

// getServiceStatuses retrieves status for all monitored services
func (sm *SystemdManager) getServiceStatuses() ([]types.Service, error) {
	var services []types.Service

	monitoredServices := sm.serviceManager.GetServices()

	for _, serviceName := range monitoredServices {
		service, err := sm.getServiceStatus(serviceName)
		if err != nil {
			log.Printf("Error getting status for service %s: %v", serviceName, err)
			// Create a service with error state
			service = types.Service{
				Name:        serviceName,
				LoadState:   "error",
				ActiveState: "unknown",
				SubState:    "unknown",
				Since:       time.Time{}, // Zero time
				Description: fmt.Sprintf("Error: %v", err),
			}
		}
		services = append(services, service)
	}

	return services, nil
}

// getServiceStatus retrieves status for a single service
func (sm *SystemdManager) getServiceStatus(serviceName string) (types.Service, error) {
	cmd := exec.Command("systemctl", "show", serviceName, "--no-page")
	output, err := cmd.Output()
	if err != nil {
		return types.Service{}, fmt.Errorf("failed to get service status: %v", err)
	}

	service := types.Service{Name: serviceName}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key, value := parts[0], parts[1]
		switch key {
		case "LoadState":
			service.LoadState = value
		case "ActiveState":
			service.ActiveState = value
		case "SubState":
			service.SubState = value
		case "ActiveEnterTimestamp":
			service.Since = sm.parseTimestamp(value, serviceName)
		case "Description":
			service.Description = value
		}
	}

	return service, nil
}

// parseTimestamp parses various timestamp formats from systemctl
func (sm *SystemdManager) parseTimestamp(value, serviceName string) time.Time {
	if value == "" || value == "0" {
		return time.Time{} // Zero time for never started services
	}

	// Try parsing as Unix timestamp in microseconds
	if timestamp, err := strconv.ParseInt(value, 10, 64); err == nil {
		if timestamp > 0 {
			// Convert microseconds to seconds
			return time.Unix(timestamp/1000000, (timestamp%1000000)*1000)
		}
	}

	// Try parsing as RFC3339 format (common systemd format)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}

	// Try parsing as RFC3339Nano format
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t
	}

	// Try parsing common systemd date formats
	formats := []string{
		"Mon 2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05-07:00",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t
		}
	}

	// If all parsing fails, try to get the timestamp using a different approach
	return sm.getServiceTimestampAlternative(serviceName)
}

// getServiceTimestampAlternative gets service timestamp using systemctl status
func (sm *SystemdManager) getServiceTimestampAlternative(serviceName string) time.Time {
	cmd := exec.Command("systemctl", "status", serviceName, "--no-pager", "-l")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Failed to get alternative timestamp for %s: %v", serviceName, err)
		return time.Time{}
	}

	// Parse the status output to find the "Active:" line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Active:") {
			// Look for "since" in the line
			if sinceIndex := strings.Index(line, "since "); sinceIndex != -1 {
				timeStr := line[sinceIndex+6:] // Skip "since "

				// Try to parse the time string
				formats := []string{
					"Mon 2006-01-02 15:04:05 MST",
					"2006-01-02 15:04:05",
					"Mon Jan 2 15:04:05 2006",
					"Jan 2 15:04:05",
				}

				for _, format := range formats {
					if t, err := time.Parse(format, timeStr); err == nil {
						return t
					}
				}
			}
		}
	}

	log.Printf("Could not parse timestamp for service %s", serviceName)
	return time.Time{}
}

// executeSystemctlCommand executes a systemctl command
func (sm *SystemdManager) executeSystemctlCommand(action, serviceName string) error {
	cmd := exec.Command("sudo", "systemctl", action, serviceName)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemctl %s %s failed: %v, stderr: %s", action, serviceName, err, stderr.String())
	}

	return nil
}

// getServiceLogs retrieves logs for a service
func (sm *SystemdManager) getServiceLogs(serviceName, lines string) (string, error) {
	cmd := exec.Command("journalctl", "-u", serviceName, "-n", lines, "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v", err)
	}

	// Clean up the output
	logs := string(output)
	if logs == "" {
		logs = "No logs available for this service."
	}

	return logs, nil
}

// AddServiceHandler adds a new service to monitor
func (sm *SystemdManager) AddServiceHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	serviceName := strings.TrimSpace(r.FormValue("service_name"))
	validate := r.FormValue("validate") == "on"

	if serviceName == "" {
		http.Error(w, "Service name is required", http.StatusBadRequest)
		return
	}

	// Validate service name format
	if !isValidServiceName(serviceName) {
		http.Error(w, "Invalid service name format", http.StatusBadRequest)
		return
	}

	// Validate service exists if requested
	if validate {
		if err := sm.serviceManager.ValidateService(serviceName); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// Add the service
	if err := sm.serviceManager.AddService(serviceName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service added successfully"))
}

// isValidServiceName checks if a service name is valid
func isValidServiceName(name string) bool {
	if len(name) == 0 || len(name) > 100 {
		return false
	}

	// Allow letters, numbers, hyphens, underscores, and dots
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_' || char == '.') {
			return false
		}
	}

	return true
}
