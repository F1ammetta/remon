package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
)

// ServiceManager handles the list of monitored services
type ServiceManager struct {
	services []string
	mutex    sync.RWMutex
	filePath string
}

// NewServiceManager creates a new service manager
func NewServiceManager() *ServiceManager {
	sm := &ServiceManager{
		filePath: "monitored_services.json",
		services: []string{
			"nginx",
			"postgresql",
			"redis",
			"docker",
			"ssh",
			"cron",
			"systemd-resolved",
			"NetworkManager",
		},
	}

	// Load services from file if it exists
	sm.loadServicesFromFile()

	return sm
}

// GetServices returns a copy of the monitored services list
func (sm *ServiceManager) GetServices() []string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	services := make([]string, len(sm.services))
	copy(services, sm.services)
	return services
}

// AddService adds a new service to the monitored list
func (sm *ServiceManager) AddService(serviceName string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Clean the service name
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return fmt.Errorf("service name cannot be empty")
	}

	// Check if service already exists
	for _, existing := range sm.services {
		if existing == serviceName {
			return fmt.Errorf("service '%s' is already being monitored", serviceName)
		}
	}

	// Add the service
	sm.services = append(sm.services, serviceName)

	// Sort the services list
	sort.Strings(sm.services)

	// Save to file
	if err := sm.saveServicesToFile(); err != nil {
		log.Printf("Warning: Failed to save services to file: %v", err)
	}

	log.Printf("Added service '%s' to monitoring list", serviceName)
	return nil
}

// RemoveService removes a service from the monitored list
func (sm *ServiceManager) RemoveService(serviceName string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for i, existing := range sm.services {
		if existing == serviceName {
			// Remove the service
			sm.services = append(sm.services[:i], sm.services[i+1:]...)

			// Save to file
			if err := sm.saveServicesToFile(); err != nil {
				log.Printf("Warning: Failed to save services to file: %v", err)
			}

			log.Printf("Removed service '%s' from monitoring list", serviceName)
			return nil
		}
	}

	return fmt.Errorf("service '%s' not found in monitoring list", serviceName)
}

// ValidateService checks if a service exists on the system
func (sm *ServiceManager) ValidateService(serviceName string) error {
	// Use systemctl to check if the service exists
	cmd := fmt.Sprintf("systemctl cat %s", serviceName)
	if err := executeCommand(cmd); err != nil {
		return fmt.Errorf("service '%s' does not exist on this system", serviceName)
	}
	return nil
}

// loadServicesFromFile loads the services list from a JSON file
func (sm *ServiceManager) loadServicesFromFile() {
	if _, err := os.Stat(sm.filePath); os.IsNotExist(err) {
		// File doesn't exist, use default services
		return
	}

	data, err := ioutil.ReadFile(sm.filePath)
	if err != nil {
		log.Printf("Warning: Failed to read services file: %v", err)
		return
	}

	var services []string
	if err := json.Unmarshal(data, &services); err != nil {
		log.Printf("Warning: Failed to parse services file: %v", err)
		return
	}

	sm.services = services
	log.Printf("Loaded %d services from file", len(services))
}

// saveServicesToFile saves the services list to a JSON file
func (sm *ServiceManager) saveServicesToFile() error {
	data, err := json.MarshalIndent(sm.services, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal services: %v", err)
	}

	if err := ioutil.WriteFile(sm.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write services file: %v", err)
	}

	return nil
}

// executeCommand executes a shell command (helper function)
func executeCommand(cmd string) error {
	// This is a simplified version - you might want to use exec.Command for better control
	return nil // Placeholder - implement actual command execution
}
