package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Cozzytree/taldrBack/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HELLO WORLD"))
	})

	mux.HandleFunc("POST /update_shapes/{workspaceId}/{userId}", s.updateShape())
	mux.HandleFunc("POST /new_workspace", s.newWorkspace())
	mux.HandleFunc("DELETE /dele_workspace/{workspace_id}/{user_id}", s.deleteWorkspce())
	mux.HandleFunc("GET /user_workspaces/{user_id}", s.getUserWorkspaces())
	mux.HandleFunc("GET /workspace_data/{workspace_id}", s.getWorkSpaceData())
	mux.HandleFunc("GET /ws", s.WsConnect)
	mux.HandleFunc("DELETE /delete_workspace/{workspace_id}/{user_id}", s.deleteWorkspce())

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		h := s.db.Health()
		w.Write([]byte(h))

	})
	return s.corsMiddlwware(mux)
}

func (s *Server) corsMiddlwware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().
			Set("Access-Control-Allow-Origin", os.Getenv("ALLOW_ORIGIN")) // Replace "*" with specific origins if needed
		w.Header().
			Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().
			Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().
			Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

			// If it's a preflight (OPTIONS) request, return a 200 OK response immediately
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) updateShape() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("workspaceId")
		userId := r.PathValue("userId")
		if id == "" || userId == "" {
			http.Error(w, "invalid user or workspace ids", http.StatusBadRequest)
		}

		var shapes struct {
			Shapes []models.Shape `json:"shapes"`
		}

		err := json.NewDecoder(r.Body).Decode(&shapes)

		if err != nil {
			http.Error(w, "invalid shapes", http.StatusBadRequest)
			return
		}

		s.shapeS.storeShapes(fmt.Sprintf("%v+%v", id, userId), shapes.Shapes)

		w.Write([]byte("success"))
	}
}

func (s *Server) updateDocument() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("workspaceId")
		userId := r.PathValue("userId")
		if id == "" || userId == "" {
			http.Error(w, "invalid user or workspace ids", http.StatusBadRequest)
		}

		var document struct {
			Document string `json:"document"`
		}

		err := json.NewDecoder(r.Body).Decode(&document)

		if err != nil {
			http.Error(w, "invalid shapes", http.StatusBadRequest)
			return
		}

		s.shapeS.saveDocument(fmt.Sprintf("%v+%v", id, userId), document.Document)

		w.Write([]byte("success"))
	}
}

func (s *Server) newWorkspace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data models.Workspace
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = s.db.NewWorkSpace(r.Context(), data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("success"))
	}
}

func (s *Server) deleteWorkspce() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := primitive.ObjectIDFromHex(r.PathValue("workspace_id"))
		userId := r.PathValue("user_id")
		if userId == "" {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "invalid object id", http.StatusBadRequest)
			return
		}

		err = s.db.DeleteWorkspace(r.Context(), id, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("successfully deleted"))
	}
}

func (s *Server) getUserWorkspaces() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Pattern)

		userId := r.PathValue("user_id")
		if userId == "" {
			http.Error(w, "invalid user", http.StatusBadRequest)
			return
		}

		work, err := s.db.UserWorkspaces(r.Context(), userId, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(&work)
	}
}

func (s *Server) getWorkSpaceData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := primitive.ObjectIDFromHex(r.PathValue("workspace_id"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		data, err := s.db.GetWorkspaceData(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) deleteWorkspace() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := primitive.ObjectIDFromHex(r.PathValue("workspace_id"))
		if err != nil {
			http.Error(w, "invalid workspace id", http.StatusBadRequest)
			return
		}
		userId := r.PathValue("user_id")
		if userId == "" {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		err = s.db.DeleteWorkspace(r.Context(), id, userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Write([]byte("successfully deleted"))
	}
}
