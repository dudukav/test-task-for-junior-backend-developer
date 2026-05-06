package tasktemplate

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	tasktemplatedomain "example.com/taskservice/internal/domain/task_template"
	tasktemplateusecase "example.com/taskservice/internal/usecase/task_template"
)

type TaskHandler struct {
	usecase tasktemplateusecase.Usecase
}

func NewTaskTemplateHandler(usecase tasktemplateusecase.Usecase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req taskTemplateMutationDTO
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	created, err := h.usecase.Create(r.Context(), tasktemplateusecase.CreateInput{
		Title:            req.Title,
		Description:      req.Description,
		AssignedID:       req.AssignedID,
		RecurrenceType:   tasktemplatedomain.RecurrenceType(req.RecurrenceType),
		RecurrenceConfig: req.RecurrenceConfig,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		Status:           req.Status,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newTaskTemplateDTO(created))
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	task, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskTemplateDTO(task))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var req taskTemplateMutationDTO
	if err = decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	updated, err := h.usecase.Update(r.Context(), id, tasktemplateusecase.UpdateInput{
		Title:            &req.Title,
		Description:      &req.Description,
		AssignedID:       &req.AssignedID,
		RecurrenceType:   (*tasktemplatedomain.RecurrenceType)(&req.RecurrenceType),
		RecurrenceConfig: &req.RecurrenceConfig,
		StartDate:        &req.StartDate,
		EndDate:          req.EndDate,
		Status:           &req.Status,
	})
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newTaskTemplateDTO(updated))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getIDFromRequest(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err = h.usecase.Delete(r.Context(), id); err != nil {
		writeUsecaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.usecase.List(r.Context())
	if err != nil {
		writeUsecaseError(w, err)
		return
	}

	response := make([]taskTemplateDTO, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskTemplateDTO(tasks[i]))
	}

	writeJSON(w, http.StatusOK, response)
}

func getIDFromRequest(r *http.Request) (uuid.UUID, error) {
	rawID := mux.Vars(r)["id"]
	if rawID == "" {
		return uuid.Nil, errors.New("missing task id")
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		return uuid.Nil, errors.New("invalid task id")
	}

	if id == uuid.Nil {
		return uuid.Nil, errors.New("invalid task id")
	}

	return id, nil
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("%w: failed to decode request", err)
	}

	return nil
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tasktemplatedomain.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, tasktemplateusecase.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{
		"error": err.Error(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}
