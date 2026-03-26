package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/kurt4ins/taskmanager/internal/indexer"
	"github.com/kurt4ins/taskmanager/internal/middleware"
	"github.com/kurt4ins/taskmanager/internal/repo"
	"github.com/kurt4ins/taskmanager/internal/utils"
)

type TaskHandler struct {
	repo    repo.TaskRepository
	indexer *indexer.Indexer
}

func NewTaskHandler(repo repo.TaskRepository, idx *indexer.Indexer) *TaskHandler {
	return &TaskHandler{repo: repo, indexer: idx}
}

func checkOwner(w http.ResponseWriter, r *http.Request, task *repo.Task) bool {
	userId, ok := middleware.UserIdFromContext(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	if task.UserId != userId {
		utils.WriteError(w, http.StatusForbidden, "forbidden")
		return false
	}

	return true
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.PathValue("userId"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var data repo.Task
	if err := utils.ReadJSON(w, r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	data.UserId = userId

	task, err := h.repo.Create(r.Context(), data)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to create task")
		return
	}

	if h.indexer != nil {
		ok := h.indexer.Submit(indexer.IndexJob{
			TaskId:      task.Id,
			Title:       task.Title,
			Description: task.Description,
		})

		if !ok {
			utils.WriteError(w, http.StatusServiceUnavailable, "indexer queue is full, try again later")
			return
		}
	}

	utils.WriteJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetById(w http.ResponseWriter, r *http.Request) {
	strId := r.PathValue("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	task, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch task")
		return
	}
	if task == nil {
		utils.WriteError(w, http.StatusNotFound, fmt.Sprintf("task with id %d doesn't exist", id))
		return
	}

	utils.WriteJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	var data repo.Task
	if err := utils.ReadJSON(w, r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	data.Id = id

	exists, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exists == nil {
		utils.WriteError(w, http.StatusNotFound, "task doesn't exist")
		return
	}

	if !checkOwner(w, r, exists) {
		return
	}

	task, err := h.repo.Update(r.Context(), data)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to update task")
		return
	}

	utils.WriteJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	var data repo.PatchTask
	if err := utils.ReadJSON(w, r, &data); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	exists, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exists == nil {
		utils.WriteError(w, http.StatusNotFound, "task doesn't exist")
		return
	}

	if !checkOwner(w, r, exists) {
		return
	}

	task, err := h.repo.Patch(r.Context(), id, data)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to patch task")
		return
	}

	utils.WriteJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid id provided")
		return
	}

	exists, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if exists == nil {
		utils.WriteError(w, http.StatusNotFound, "task doesn't exist")
		return
	}

	if !checkOwner(w, r, exists) {
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userId, err := strconv.Atoi(r.PathValue("userId"))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	limit, offset := 20, 0

	if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
		l, err := strconv.Atoi(strLimit)
		if err != nil || l <= 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = l
	}

	if strOffset := r.URL.Query().Get("offset"); strOffset != "" {
		o, err := strconv.Atoi(strOffset)
		if err != nil || o < 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = o
	}

	tasks, err := h.repo.ListByUser(r.Context(), userId, limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}

	if tasks == nil {
		tasks = []repo.Task{}
	}

	utils.WriteJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset := 20, 0

	if strLimit := r.URL.Query().Get("limit"); strLimit != "" {
		l, err := strconv.Atoi(strLimit)
		if err != nil || l <= 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = l
	}

	if strOffset := r.URL.Query().Get("offset"); strOffset != "" {
		o, err := strconv.Atoi(strOffset)
		if err != nil || o < 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = o
	}

	tasks, err := h.repo.List(r.Context(), limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		utils.WriteError(w, http.StatusInternalServerError, "failed to fetch tasks")
		return
	}

	if tasks == nil {
		tasks = []repo.Task{}
	}

	utils.WriteJSON(w, http.StatusOK, tasks)
}
