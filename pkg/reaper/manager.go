package reaper

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nikhi/nanolambda/pkg/docker"
)

// containerinfo tracks the state of a running function container
type ContainerInfo struct {
	ID           string
	Address      string
	LastAccessed time.Time
	Timeout      time.Duration
}

// manager handles the lifecycle of containers (idle cleanup)
type Manager struct {
	docker     *docker.Manager
	mu         sync.RWMutex
	containers map[string]*ContainerInfo // map[functionname]*containerinfo
}

// newmanager creates a new reaper manager
func NewManager(d *docker.Manager) *Manager {
	return &Manager{
		docker:     d,
		containers: make(map[string]*ContainerInfo),
	}
}

// getcontainer returns the address of a running container if it exists
func (m *Manager) GetContainer(name string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.containers[name]
	if !exists {
		return "", false
	}
	return info.Address, true
}

// register adds a new container to the tracker
func (m *Manager) Register(name, id, address string, timeoutSeconds int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second // default to 10s if not specified
	}

	m.containers[name] = &ContainerInfo{
		ID:           id,
		Address:      address,
		LastAccessed: time.Now(),
		Timeout:      timeout,
	}
	fmt.Printf("[reaper] registered container for %s (id: %s, timeout: %s)\n", name, id[:12], timeout)
}

// touch updates the last accessed time for a container
func (m *Manager) Touch(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if info, exists := m.containers[name]; exists {
		info.LastAccessed = time.Now()
	}
}

// start runs the background cleanup loop
func (m *Manager) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Println("[reaper] background job started")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[reaper] stopping background job")
			return
		case <-ticker.C:
			m.cleanup()
		}
	}
}

func (m *Manager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for name, info := range m.containers {
		if now.Sub(info.LastAccessed) > info.Timeout {
			fmt.Printf("[reaper] container for %s idle for %v. stopping...\n", name, now.Sub(info.LastAccessed))
			
			// stop container in docker
			// we use a background context because cleanup shouldn't be cancelled by request context
			if err := m.docker.StopContainer(context.Background(), info.ID); err != nil {
				log.Printf("error stopping container %s: %v", info.ID, err)
			}

			// remove from map
			delete(m.containers, name)
		}
	}
}