package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/s-piazzano/cosmoria/internal/adminauth"
	"github.com/s-piazzano/cosmoria/internal/collections"
)

type CollectionsHandler struct {
	Service *collections.Service
}

type createCollectionRequest struct {
	Name   string             `json:"name"`
	Schema collections.Schema `json:"schema"`
}

type updateCollectionSchemaRequest struct {
	Schema collections.Schema `json:"schema"`
}

func (h *CollectionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_project_id"})
		return
	}

	var req createCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name_required"})
		return
	}

	c, err := h.Service.CreateCollection(r.Context(), projectID, req.Name, req.Schema)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "create_failed"})
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (h *CollectionsHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_project_id"})
		return
	}

	list, err := h.Service.ListCollections(r.Context(), projectID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if list == nil {
		list = []collections.Collection{}
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *CollectionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	collectionID := r.PathValue("cid")
	if projectID == "" || collectionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	c, err := h.Service.GetCollection(r.Context(), collectionID, projectID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "collection_not_found"})
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionsHandler) UpdateSchema(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	collectionID := r.PathValue("cid")
	if projectID == "" || collectionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	var req updateCollectionSchemaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	c, err := h.Service.UpdateCollectionSchema(r.Context(), collectionID, projectID, req.Schema)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "update_failed"})
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *CollectionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := adminauth.GetAdminAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	projectID := r.PathValue("pid")
	collectionID := r.PathValue("cid")
	if projectID == "" || collectionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	if err := h.Service.DeleteCollection(r.Context(), collectionID, projectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
