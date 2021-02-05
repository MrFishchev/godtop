package interfaces

import (
	"fmt"
	"godtop/application"
	"godtop/domain"
	"log"
	"net/http"
	"os"

	_ "godtop/docs"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
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
	ctx.AbortWithStatusJSON(code, e)
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
func (h Handler) routes(port string) *gin.Engine {
	r := gin.Default()
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api")
	{
		api.GET("/containers", h.getRunningContainers)
		api.GET("/containers/all", h.getAllContainers)
		api.GET("/container/:nameOrId", h.getContainer)
		api.GET("/container/:nameOrId/stats", h.getContainerStats)
		api.GET("/volumes", h.getVolumes)
		api.GET("/host", h.getHostInfo)
	}

	return r
}

//RunServer starts server on a specific port
func (h Handler) RunServer(port int) error {
	log.Printf("Server running at http://localhost:%d/", port)
	portStr := fmt.Sprintf(":%d", port)
	return h.routes(portStr).Run(portStr)
}

//region API Handlers

// getContainer godoc
// @Summary Retrieves container information by its Id or Name
// @Produce json
// @Param nameOrId path string true "container Name or Id"
// @Success 200 {object} domain.Container
// @Router /container/{nameOrId} [get]
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

// getRunningContainers godoc
// @Summary Retrieves running containers
// @Produce json
// @Success 200 {array} domain.Container
// @Router /containers [get]
func (h Handler) getRunningContainers(ctx *gin.Context) {
	h.getContainers(ctx, false)
}

// getAllContainers godoc
// @Summary Retrieves all containers
// @Produce json
// @Success 200 {array} domain.Container
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

// getContainerStats godoc
// @Summary Retrieves statistics of a container
// @Produce json
// @Param nameOrId path string true "container Name or Id"
// @Success 200 {object} domain.ContainerStats
// @Router /container/{nameOrId}/stats [get]
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

// getVolumes godoc
// @Summary Retrieves mounted and created volumes
// @Produce json
// @Success 200 {array} domain.Volume
// @Router /volumes [get]
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

// getHostInfo godoc
// @Summary Retrieves information about host stystem
// @Produce json
// @Success 200 {object} domain.HostInfo
// @Router /host [get]
func (h Handler) getHostInfo(ctx *gin.Context) {
	interactor := application.HostInteractor{
		Service: h.HostService,
	}

	info := interactor.GetInfo(ctx)

	Ok(ctx, info)
}

//endregion
