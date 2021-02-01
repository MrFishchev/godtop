package interfaces

import (
	"encoding/json"
	"fmt"
	"godtop/application"
	"godtop/domain"
	"log"
	"net/http"
	"os"

	router "github.com/takashabe/go-router"
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
	Service domain.DockerService
}

//Routes returns the initialized router
func (h Handler) Routes() *router.Router {
	r := router.NewRouter()
	r.Get("/containers", h.getRunningContainers)
	r.Get("/containers/all", h.getAllContainers)
	r.Get("/container/:nameOrId", h.getContainer)
	r.Get("/volumes", h.getVolumes)
	return r
}

//RunServer starts server on a specific port
func (h Handler) RunServer(port int) error {
	log.Printf("Server running at http://localhost:%d/", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), h.Routes())
}

func (h Handler) getContainer(w http.ResponseWriter, r *http.Request, nameOrId string) {
	ctx := r.Context()

	interactor := application.ContainerInteractor{
		Service: h.Service,
	}

	container, err := interactor.Get(ctx, nameOrId)
	if err != nil {
		Error(w, http.StatusNotFound, err, err.Error())
		return
	}

	Ok(w, http.StatusOK, container)
}

func (h Handler) getRunningContainers(w http.ResponseWriter, r *http.Request) {
	h.getContainers(&w, r, false)
}

func (h Handler) getAllContainers(w http.ResponseWriter, r *http.Request) {
	h.getContainers(&w, r, true)
}

func (h Handler) getContainers(w *http.ResponseWriter, r *http.Request, all bool) {
	ctx := r.Context()

	interactor := application.ContainerInteractor{
		Service: h.Service,
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

func (h Handler) getVolumes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	interactor := application.VolumeInteractor{
		Service: h.Service,
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
