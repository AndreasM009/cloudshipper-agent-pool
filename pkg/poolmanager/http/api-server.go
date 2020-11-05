package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/repositories"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/domain"

	"github.com/andreasM009/cloudshipper-agent-pool/pkg/poolmanager/services"
	"github.com/gorilla/mux"
)

// APIServer Http server
type APIServer struct {
	port              string
	router            *mux.Router
	deploymentService *services.DeploymentService
	server            *http.Server
	done              chan bool
	deplomentRepo     repositories.DeploymentRepository
}

// NewHTTPServer new instance
func NewHTTPServer(port string, deploymentService *services.DeploymentService, deploymentRepo repositories.DeploymentRepository) *APIServer {
	server := &APIServer{
		port:              port,
		router:            mux.NewRouter(),
		deploymentService: deploymentService,
		deplomentRepo:     deploymentRepo,
		server: &http.Server{
			Addr: fmt.Sprintf(":%s", port),
		},
		done: make(chan bool, 1),
	}

	server.server.Handler = server.router

	server.router.HandleFunc("/tenants/{id}/deployments", server.onPostDeployment).
		Methods(http.MethodPost, http.MethodOptions)

	server.router.HandleFunc("/tenants/{id}/deployments", server.onGetDeployment).
		Methods(http.MethodGet, http.MethodOptions)

	server.router.Use(mux.CORSMethodMiddleware(server.router))
	return server
}

// ListentAndServeAsync listen and serve async
func (server *APIServer) ListentAndServeAsync(ctx context.Context) {
	go func() {
		log.Println("Starting API Server")
		if err := server.server.ListenAndServe(); err != nil {
			log.Println(err)
			return
		}
	}()

	// watch for cancel
	go func() {
		<-ctx.Done()
		srvctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		// wait defined seconds to shutdown server
		server.server.Shutdown(srvctx)
		log.Println("API Server stopped")
		server.done <- true
	}()
}

// Done stopped?
func (server *APIServer) Done() <-chan bool {
	return server.done
}

func (server *APIServer) onPostDeployment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	params := mux.Vars(r)
	id := params["id"]

	reqobj := struct {
		DefinitionID   string `json:"definitionId"`
		DeploymentName string `json:"deploymentName"`
		Yaml           string `json:"yaml"`
		Parrameters    string `json:"parameters"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&reqobj); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	yamlb, err := base64.StdEncoding.DecodeString(reqobj.Yaml)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	paramb, err := base64.StdEncoding.DecodeString(reqobj.Parrameters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	dep := domain.Deployment{
		DefinitionID:   reqobj.DefinitionID,
		DeploymentName: reqobj.DeploymentName,
		TenantID:       id,
	}

	if err := server.deploymentService.EnqueueDeployment(&dep, yamlb, paramb); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(&dep)
	w.WriteHeader(http.StatusCreated)
}

func (server *APIServer) onGetDeployment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}

	params := mux.Vars(r)
	id := params["id"]

	res := []*domain.Deployment{}
	deps, err := server.deplomentRepo.GetAll(id)
	if err == nil {
		res = deps
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	json.NewEncoder(w).Encode(res)
	w.WriteHeader(http.StatusOK)
}
