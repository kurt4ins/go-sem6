package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kurt4ins/taskmanager/internal/repo"
)

type TaskHandler struct {
	repo repo.TaskRepository
}

func NewTaskHandler(repo repo.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.PathValue("userId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var data repo.Task
	if err := readJSON(w, r, &data); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	data.UserId = userId

	task, err := h.repo.Create(r.Context(), data)
	if err != nil {
		fmt.Println(err.Error())
		WriteError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	WriteJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetById(w http.ResponseWriter, r *http.Request) {
	strId := r.PathValue("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	task, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		fmt.Println(err.Error())
		WriteError(w, http.StatusInternalServerError, "failed to fetch task")
		return
	}
	if task == nil {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("task with id %d doesn't exist", id))
		return
	}

	WriteJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	var data repo.Task
	if err := readJSON(w, r, &data); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	data.Id = id

	task, err := h.repo.Update(r.Context(), data)
	if err != nil {
		fmt.Println(err.Error())
		WriteError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	if task == nil {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("task with id %d doesn't exist", id))
		return
	}

	WriteJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	var fields repo.PatchTask
	if err := readJSON(w, r, &fields); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := h.repo.Patch(r.Context(), id, fields)
	if err != nil {
		fmt.Println(err.Error())
		WriteError(w, http.StatusInternalServerError, "failed to patch task")
		return
	}
	if task == nil {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("task with id %d doesn't exist", id))
		return
	}

	WriteJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err.Error() == "TaskRepo.Delete: not found" {
			WriteError(w, http.StatusNotFound, fmt.Sprintf("task with id %d doesn't exist", id))
			return
		}
		fmt.Println(err.Error())
		WriteError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.PathValue("userId"))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	limit, offset := 20, 0

	if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
		l, err := strconv.Atoi(strLimit)
		if err != nil || l <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = l
	}

	if strOffset := r.URL.Query().Get("offset"); strOffset != "" {
		o, err := strconv.Atoi(strOffset)
		if err != nil || o < 0 {
			WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = o
	}

	tasks, err := h.repo.ListByUser(r.Context(), userId, limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		WriteError(w, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}

	if tasks == nil {
		tasks = []repo.Task{}
	}

	WriteJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := 20, 0

	if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
		l, err := strconv.Atoi(strLimit)
		if err != nil || l <= 0 {
			WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = l
	}

	if strOffset := r.URL.Query().Get("offset"); strOffset != "" {
		o, err := strconv.Atoi(strOffset)
		if err != nil || o < 0 {
			WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = o
	}

	tasks, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		WriteError(w, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}

	if tasks == nil {
		tasks = []repo.Task{}
	}

	WriteJSON(w, http.StatusOK, tasks)
}
