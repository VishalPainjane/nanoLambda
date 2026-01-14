package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/nikhi/nanolambda/pkg/docker"
	"github.com/nikhi/nanolambda/pkg/proxy"
	"github.com/nikhi/nanolambda/pkg/reaper"
	"github.com/nikhi/nanolambda/pkg/registry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"function", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
}

// App holds the application state
type App struct {
	Docker   *docker.Manager
	Registry *registry.Manager
	Reaper   *reaper.Manager
	Router   *mux.Router
}

func main() {
	app := &App{}
	var err error

	// 1. Initialize Docker Manager
	app.Docker, err = docker.NewManager()
	if err != nil {
		log.Fatalf("Error initializing Docker manager: %v", err)
	}

	// 2. Initialize Registry Manager
	// Ensure the data directory exists
	os.MkdirAll("./data", 0755)
	app.Registry, err = registry.NewManager("./data/nanolambda.db")
	if err != nil {
		log.Fatalf("Error initializing Registry: %v", err)
	}
	defer app.Registry.Close()

	// 3. Initialize Reaper (Scale-to-zero)
	app.Reaper = reaper.NewManager(app.Docker)
	// Start reaper in background
	go app.Reaper.Start(context.Background())

	// 4. Initialize Router
	app.Router = mux.NewRouter()
	
	// 5. Define Routes
	app.Router.Handle("/metrics", promhttp.Handler())
	app.Router.HandleFunc("/admin/health", app.HealthCheckHandler).Methods("GET")
	app.Router.HandleFunc("/function/{name}", app.InvokeHandler).Methods("POST")
	
	// Admin Routes
	app.Router.HandleFunc("/admin/warmup", app.WarmupHandler).Methods("POST")
	
	// 6. Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Gateway running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, app.Router))
}

// HealthCheckHandler returns simple status
func (app *App) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "component": "gateway"})
}

// InvokeHandler handles function invocation
func (app *App) InvokeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	funcName := vars["name"]

	httpRequestsTotal.WithLabelValues(funcName, "invoked").Inc()

	// 1. Check if function is already running (Hot Start)
	addr, running := app.Reaper.GetContainer(funcName)
	
	if running {
		// Update last accessed time
		app.Reaper.Touch(funcName)
	} else {
		// Cold Start Logic
		// fmt.Printf("Cold start for %s...\n", funcName)
		
		// Fetch function metadata
		fn, err := app.Registry.GetFunction(funcName)
		if err != nil {
			http.Error(w, fmt.Sprintf("Function '%s' not found", funcName), http.StatusNotFound)
			return
		}

		// Start Container
		var id string
		addr, id, err = app.Docker.StartContainer(r.Context(), fn.ImageTag, fn.Name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
			return
		}

		// Wait for Container to be Ready
		ready := false
		for i := 0; i < 20; i++ {
			resp, err := http.Get(fmt.Sprintf("http://%s/health", addr))
			if err == nil && resp.StatusCode == 200 {
				ready = true
				resp.Body.Close()
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		if !ready {
			// Clean up if it failed to start properly
			app.Docker.StopContainer(r.Context(), id)
			http.Error(w, "Container timed out starting", http.StatusGatewayTimeout)
			return
		}

		// Register with Reaper
		app.Reaper.Register(funcName, id, addr, fn.Timeout)
	}

	// 2. Proxy Request
	p := proxy.NewReverseProxy(addr)
	p.ServeHTTP(w, r)
}

// WarmupHandler handles pre-warming requests from AI
func (app *App) WarmupHandler(w http.ResponseWriter, r *http.Request) {
	// Simple implementation: Just trigger a "start" without invocation
	var req struct {
		Function string `json:"function"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if already running
	_, running := app.Reaper.GetContainer(req.Function)
	if running {
		app.Reaper.Touch(req.Function) // Extend life
		json.NewEncoder(w).Encode(map[string]string{"status": "already_running"})
		return
	}

	// Start it
	fn, err := app.Registry.GetFunction(req.Function)
	if err != nil {
		http.Error(w, "Function not found", http.StatusNotFound)
		return
	}

	addr, id, err := app.Docker.StartContainer(r.Context(), fn.ImageTag, fn.Name)
	if err != nil {
		http.Error(w, "Failed to start", http.StatusInternalServerError)
		return
	}

	// Wait for ready (simplified)
	time.Sleep(200 * time.Millisecond) // Give it a tiny headstart, real healthcheck is better but blocking here is ok for admin api

	// Use a longer timeout for warmup (e.g. 5 minutes) to ensure it's ready for the predicted spike
	const WarmupTimeout = 300
	app.Reaper.Register(req.Function, id, addr, WarmupTimeout)
	
	json.NewEncoder(w).Encode(map[string]string{"status": "warmed_up"})
}