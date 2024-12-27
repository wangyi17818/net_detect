// internal/api/server.go
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"net_detect/internal/controller"
	"net_detect/internal/models"

	"github.com/gorilla/mux"
)

type Server struct {
	ctrl *controller.Controller
}

func NewServer(ctrl *controller.Controller) *Server {
	return &Server{ctrl: ctrl}
}

func (s *Server) Start(addr string) error {
	r := mux.NewRouter()

	// API 路由
	r.HandleFunc("/api/tasks", s.listTasks).Methods("GET")
	r.HandleFunc("/api/tasks", s.createTask).Methods("POST")
	r.HandleFunc("/api/retasks", s.createTask).Methods("POST")
	r.HandleFunc("/api/tasks/{taskId}", s.deleteTask).Methods("DELETE")
	r.HandleFunc("/api/tasks/{taskId}", s.getTask).Methods("GET")

	log.Printf("Starting API server on %s", addr)
	return http.ListenAndServe(addr, r)
}

// 创建任务
func (s *Server) createTask(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		log.Printf("Failed NewDecoder: %v", r.Body)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.ctrl.AddTask(t); err != nil {
		log.Printf("Failed to add task: %v", t)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// 获取所有任务
func (s *Server) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.ctrl.ListTasks()
	json.NewEncoder(w).Encode(tasks)
}

// 获取单个任务
func (s *Server) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]

	task, exists := s.ctrl.GetTask(taskId)
	if !exists {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(task)
}

// 删除任务
func (s *Server) deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskId := vars["taskId"]

	if err := s.ctrl.RemoveTask(taskId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
