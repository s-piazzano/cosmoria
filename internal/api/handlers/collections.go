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

// @Summary Create a collection
// @Security AdminBearerAuth
// @Description Create a new dynamic collection with a schema.
// @Tags Collections
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param body body createCollectionRequest true "Collection definition"
// @Success 201 {object} collections.Collection
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/projects/{pid}/collections [post]
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

// @Summary List collections
// @Security AdminBearerAuth
// @Description List all collections for a project.
// @Tags Collections
// @Produce json
// @Param pid path string true "Project ID"
// @Success 200 {array} collections.Collection
// @Failure 500 {object} map[string]string
// @Router /api/admin/projects/{pid}/collections [get]
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

// @Summary Get a collection
// @Security AdminBearerAuth
// @Description Get a collection definition by ID.
// @Tags Collections
// @Produce json
// @Param pid path string true "Project ID"
// @Param cid path string true "Collection ID"
// @Success 200 {object} collections.Collection
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/admin/projects/{pid}/collections/{cid} [get]
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

// @Summary Update collection schema
// @Security AdminBearerAuth
// @Description Update the schema of an existing collection.
// @Tags Collections
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param cid path string true "Collection ID"
// @Param body body updateCollectionSchemaRequest true "Updated schema"
// @Success 200 {object} collections.Collection
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/projects/{pid}/collections/{cid} [put]
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

// @Summary Delete a collection
// @Security AdminBearerAuth
// @Description Delete a collection and all its records.
// @Tags Collections
// @Param pid path string true "Project ID"
// @Param cid path string true "Collection ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/projects/{pid}/collections/{cid} [delete]
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
