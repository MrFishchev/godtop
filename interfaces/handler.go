package interfaces

import (
	"fmt"
	"godtop/application"
	"godtop/domain"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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
func Error(ctx *gin.Context, code int, err error, msg string) {
	e := &ErrorResponse{
		Message: msg,
		Error:   err,
	}

	logDebug("%s", e.ToString())
	ctx.JSON(code, e)
}

func Ok(ctx *gin.Context, src ...interface{}) {
	ctx.JSON(http.StatusOK, src)
}

//Handler docker service
type Handler struct {
	DockerService domain.DockerService
	HostService   domain.HostService
}

//Routes returns the initialized router
func (h Handler) Routes() *gin.Engine {
	r := gin.Default()
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
	return h.Routes().Run(fmt.Sprintf(":%d", port))
}

func (h Handler) getHostInfo(ctx *gin.Context) {
	interactor := application.HostInteractor{
		Service: h.HostService,
	}

	info := interactor.GetInfo(ctx)

	Ok(ctx, info)
}

func (h Handler) getContainer(ctx *gin.Context) {
	nameOrId := ctx.Param("nameOrId")

	interactor := application.ContainerInteractor{
		Service: h.DockerService,
	}

	container, err := interactor.Get(ctx, nameOrId)
	if err != nil {
		Error(ctx, http.StatusNotFound, err, err.Error())
		return
	}

	Ok(ctx, container)
}

func (h Handler) getRunningContainers(ctx *gin.Context) {
	h.getContainers(ctx, false)
}

func (h Handler) getAllContainers(ctx *gin.Context) {
	h.getContainers(ctx, true)
}

func (h Handler) getContainers(ctx *gin.Context, all bool) {
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
		Error(ctx, http.StatusNotFound, err, err.Error())
		return
	}
	type payload struct {
		Containers *[]domain.Container `json:"containers"`
	}
	Ok(ctx, payload{Containers: containers})
}

func (h Handler) getContainerStats(ctx *gin.Context) {
	nameOrId := ctx.Param("nameOrId")

	interactor := application.ContainerInteractor{
		Service: h.DockerService,
	}

	stats, err := interactor.GetStats(ctx, nameOrId, false)
	if err != nil {
		Error(ctx, http.StatusNotFound, err, err.Error())
		return
	}

	type payload struct {
		Stats *domain.ContainerStats `json:"stats"`
	}

	Ok(ctx, payload{Stats: stats})
}

func (h Handler) getVolumes(ctx *gin.Context) {
	interactor := application.VolumeInteractor{
		Service: h.DockerService,
	}

	volumes, err := interactor.GetAll(ctx)
	if err != nil {
		Error(ctx, http.StatusNotFound, err, err.Error())
		return
	}

	type payload struct {
		Volumes *[]domain.Volume `json:"volumes"`
	}

	Ok(ctx, payload{Volumes: volumes})
}
