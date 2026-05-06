package transporthttp

import (
	"net/http"

	"github.com/gorilla/mux"

	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	httptaskhandlers "example.com/taskservice/internal/transport/http/handlers/task"
	httptemplatehandlers "example.com/taskservice/internal/transport/http/handlers/task_template"
)

func NewRouter(taskHandler *httptaskhandlers.Handler,
	taskTemplateHandler *httptemplatehandlers.TaskHandler,
	docsHandler *swaggerdocs.Handler) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/swagger/openapi.json", docsHandler.ServeSpec).Methods(http.MethodGet)
	router.HandleFunc("/swagger/", docsHandler.ServeUI).Methods(http.MethodGet)
	router.HandleFunc("/swagger", docsHandler.RedirectToUI).Methods(http.MethodGet)

	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/tasks", taskHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/tasks", taskHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9a-fA-F-]+}", taskHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/tasks/{id:[0-9a-fA-F-]+}", taskHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/tasks/{id:[0-9a-fA-F-]+}", taskHandler.Delete).Methods(http.MethodDelete)

	api.HandleFunc("/task-templates", taskTemplateHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/task-templates", taskTemplateHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/task-templates/{id:[0-9a-fA-F-]+}", taskTemplateHandler.GetByID).Methods(http.MethodGet)
	api.HandleFunc("/task-templates/{id:[0-9a-fA-F-]+}", taskTemplateHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/task-templates/{id:[0-9a-fA-F-]+}", taskTemplateHandler.Delete).Methods(http.MethodDelete)

	return router
}
