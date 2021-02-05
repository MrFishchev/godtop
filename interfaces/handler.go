package interfaces

import (
	"encoding/json"
	"fmt"
	"godtop/application"
	"godtop/domain"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func logDebug(format string, args ...interface{}) {
	if env := os.Getenv("GO_SERVER_DEBUG"); len(env) != 0 {
		log.Printf("[DEBUG] "+format+"\n", args...)
	}
}

//ErrorResponse is error respose template
type ErrorResponse struct {
	Message string `json:"reason"`
	Error   error  `json:"-"`
}

func (e *ErrorResponse) ToString() string {
	return fmt.Sprintf("reason: %s, error: %s", e.Message, e.Error)
}

//Error is wrapped Respond when error
func Error(w http.ResponseWriter, code int, err error, msg string) {
	e := &ErrorResponse{
		Message: msg,
		Error:   err,
	}

	logDebug("%s", e.ToString())
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	Respond(w, code, e)
}

func Ok(w http.ResponseWriter, code int, src ...interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	Respond(w, code, src)
}

//Respond writes response to ResponseWriter
func Respond(w http.ResponseWriter, code int, src interface{}) {
	var body []byte
	var err error

	switch s := src.(type) {
	case []byte:
		if !json.Valid(s) {
			Error(w, http.StatusInternalServerError, err, "Invalid json")
			return
		}
		body = s
	case string:
		body = []byte(s)
	case *ErrorResponse, ErrorResponse:
		//if cannot parse json of ErrorResponse, avoid inifinite loop
		if body, err = json.Marshal(src); err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("{\"reason\":\"Failed to parse json\"}"))
			return
		}
	default:
		if body, err = json.Marshal(src); err != nil {
			Error(w, http.StatusInternalServerError, err, "Failed to parse json")
			return
		}
	}

	w.WriteHeader(code)
	w.Write(body)
}

//Handler docker service
type Handler struct {
	DockerService domain.DockerService
	HostService   domain.HostService
}

//Routes returns the initialized router
func (h Handler) Routes() *httprouter.Router {
	r := httprouter.New()
	r.GET("/containers", h.getRunningContainers)
	r.GET("/containers/all", h.getAllContainers)
	r.GET("/container/:nameOrId", h.getContainer)
	r.GET("/container/:nameOrId/stats", h.getContainerStats)
	r.GET("/volumes", h.getVolumes)
	r.GET("/host", h.getHostInfo)
	return r
}

//RunServer starts server on a specific port
func (h Handler) RunServer(port int) error {
	log.Printf("Server running at http://localhost:%d/", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), h.Routes())
}

func (h Handler) getHostInfo(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	interactor := application.HostInteractor{
		Service: h.HostService,
	}

	info := interactor.GetInfo(ctx)

	Ok(w, http.StatusOK, info)
}

func (h Handler) getContainer(w http.ResponseWriter, r *http.Request, args httprouter.Params) {
	nameOrId := args.ByName("nameOrId")
	ctx := r.Context()

	interactor := application.ContainerInteractor{
		Service: h.DockerService,
	}

	container, err := interactor.Get(ctx, nameOrId)
	if err != nil {
		Error(w, http.StatusNotFound, err, err.Error())
		return
	}

	Ok(w, http.StatusOK, container)
}

func (h Handler) getRunningContainers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	h.getContainers(&w, r, false)
}

func (h Handler) getAllContainers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	h.getContainers(&w, r, true)
}

func (h Handler) getContainers(w *http.ResponseWriter, r *http.Request, all bool) {

	ctx := r.Context()

	interactor := application.ContainerInteractor{
		Service: h.DockerService,
	}
	var containers *[]domain.Container
	var err error
	if all {
		containers, err = interactor.GetAll(ctx)

	} else {
		containers, err = interactor.GetRunning(ctx)
	}

	if err != nil {
		Error(*w, http.StatusNotFound, err, err.Error())
		return
	}
	type payload struct {
		Containers *[]domain.Container `json:"containers"`
	}
	Ok(*w, http.StatusOK, payload{Containers: containers})
}

func (h Handler) getContainerStats(w http.ResponseWriter, r *http.Request, args httprouter.Params) {
	nameOrId := args.ByName("nameOrId")
	ctx := r.Context()

	interactor := application.ContainerInteractor{
		Service: h.DockerService,
	}

	stats, err := interactor.GetStats(ctx, nameOrId, false)
	if err != nil {
		Error(w, http.StatusNotFound, err, err.Error())
		return
	}

	type payload struct {
		Stats *domain.ContainerStats `json:"stats"`
	}

	Ok(w, http.StatusOK, payload{Stats: stats})
}

func (h Handler) getVolumes(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	ctx := r.Context()

	interactor := application.VolumeInteractor{
		Service: h.DockerService,
	}

	volumes, err := interactor.GetAll(ctx)
	if err != nil {
		Error(w, http.StatusNotFound, err, err.Error())
		return
	}

	type payload struct {
		Volumes *[]domain.Volume `json:"volumes"`
	}

	Ok(w, http.StatusOK, payload{Volumes: volumes})
}
