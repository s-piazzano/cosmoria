package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/realtime"
	"github.com/s-piazzano/cosmoria/internal/records"
)

type RecordsHandler struct {
	Service   *records.Service
	Publisher *realtime.Publisher
}

type createRecordRequest struct {
	Data map[string]any `json:"data"`
}

type updateRecordRequest struct {
	Data map[string]any `json:"data"`
}

// @Summary Create a record
// @Security BearerAuth
// @Description Create a new record in a collection. Data is validated against the collection schema.
// @Tags Records
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param cid path string true "Collection ID"
// @Param body body createRecordRequest true "Record data"
// @Success 201 {object} records.Record
// @Failure 400 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid}/collections/{cid}/records [post]
func (h *RecordsHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := strPtr(r.PathValue("tid"))
	collectionID := r.PathValue("cid")
	if collectionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	var req createRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Data == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "data_required"})
		return
	}

	record, err := h.Service.CreateRecord(r.Context(), claims.ProjectID, tenantID, collectionID, req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if h.Publisher != nil {
		payload, _ := json.Marshal(map[string]string{"collection_id": collectionID})
		h.Publisher.Publish(&realtime.Event{
			ProjectID:  claims.ProjectID,
			TenantID:   safeStr(tenantID),
			Resource:   "records",
			Action:     "create",
			ResourceID: record.ID,
			Payload:    payload,
		})
	}

	writeJSON(w, http.StatusCreated, record)
}

// @Summary List records
// @Security BearerAuth
// @Description List records in a collection with cursor-based pagination.
// @Tags Records
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param cid path string true "Collection ID"
// @Param cursor query string false "Pagination cursor (record ID)"
// @Param limit query int false "Page size (default 50, max 100)"
// @Success 200 {object} map[string]any "data: [Record], next_cursor: string"
// @Failure 400 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid}/collections/{cid}/records [get]
func (h *RecordsHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := strPtr(r.PathValue("tid"))
	collectionID := r.PathValue("cid")
	if collectionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	cursor := r.URL.Query().Get("cursor")
	limit := 50
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	recs, nextCursor, err := h.Service.ListRecords(r.Context(), claims.ProjectID, tenantID, collectionID, cursor, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "list_failed"})
		return
	}
	if recs == nil {
		recs = []records.Record{}
	}

	resp := map[string]any{
		"data": recs,
	}
	if nextCursor != "" {
		resp["next_cursor"] = nextCursor
	}

	writeJSON(w, http.StatusOK, resp)
}

// @Summary Get a record
// @Security BearerAuth
// @Description Get a single record by ID.
// @Tags Records
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param cid path string true "Collection ID"
// @Param rid path string true "Record ID"
// @Success 200 {object} records.Record
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid} [get]
func (h *RecordsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := strPtr(r.PathValue("tid"))
	recordID := r.PathValue("rid")
	if recordID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	record, err := h.Service.GetRecord(r.Context(), recordID, claims.ProjectID, tenantID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "record_not_found"})
		return
	}
	writeJSON(w, http.StatusOK, record)
}

// @Summary Update a record
// @Security BearerAuth
// @Description Replace a record's data (full JSONB replacement). Validated against the schema.
// @Tags Records
// @Accept json
// @Produce json
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param cid path string true "Collection ID"
// @Param rid path string true "Record ID"
// @Param body body updateRecordRequest true "Updated data"
// @Success 200 {object} records.Record
// @Failure 400 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid} [put]
func (h *RecordsHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := strPtr(r.PathValue("tid"))
	collectionID := r.PathValue("cid")
	recordID := r.PathValue("rid")
	if collectionID == "" || recordID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	var req updateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}
	if req.Data == nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "data_required"})
		return
	}

	record, err := h.Service.UpdateRecord(r.Context(), recordID, claims.ProjectID, tenantID, req.Data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if h.Publisher != nil {
		payload, _ := json.Marshal(map[string]string{"collection_id": collectionID})
		h.Publisher.Publish(&realtime.Event{
			ProjectID:  claims.ProjectID,
			TenantID:   safeStr(tenantID),
			Resource:   "records",
			Action:     "update",
			ResourceID: record.ID,
			Payload:    payload,
		})
	}

	writeJSON(w, http.StatusOK, record)
}

// @Summary Delete a record
// @Security BearerAuth
// @Description Delete a record by ID.
// @Tags Records
// @Param pid path string true "Project ID"
// @Param tid path string true "Tenant ID"
// @Param cid path string true "Collection ID"
// @Param rid path string true "Record ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/projects/{pid}/tenants/{tid}/collections/{cid}/records/{rid} [delete]
func (h *RecordsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := strPtr(r.PathValue("tid"))
	collectionID := r.PathValue("cid")
	recordID := r.PathValue("rid")
	if collectionID == "" || recordID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	if err := h.Service.DeleteRecord(r.Context(), recordID, claims.ProjectID, tenantID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}

	if h.Publisher != nil {
		payload, _ := json.Marshal(map[string]string{"collection_id": collectionID})
		h.Publisher.Publish(&realtime.Event{
			ProjectID:  claims.ProjectID,
			TenantID:   safeStr(tenantID),
			Resource:   "records",
			Action:     "delete",
			ResourceID: recordID,
			Payload:    payload,
		})
	}

	w.WriteHeader(http.StatusNoContent)
}
