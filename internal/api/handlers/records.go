package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/s-piazzano/cosmoria/internal/auth"
	"github.com/s-piazzano/cosmoria/internal/records"
)

type RecordsHandler struct {
	Service *records.Service
}

type createRecordRequest struct {
	Data map[string]any `json:"data"`
}

type updateRecordRequest struct {
	Data map[string]any `json:"data"`
}

func (h *RecordsHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	collectionID := r.PathValue("cid")
	if tenantID == "" || collectionID == "" {
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
	writeJSON(w, http.StatusCreated, record)
}

func (h *RecordsHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	collectionID := r.PathValue("cid")
	if tenantID == "" || collectionID == "" {
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

func (h *RecordsHandler) Get(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	recordID := r.PathValue("rid")
	if tenantID == "" || recordID == "" {
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

func (h *RecordsHandler) Update(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	recordID := r.PathValue("rid")
	if tenantID == "" || recordID == "" {
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
	writeJSON(w, http.StatusOK, record)
}

func (h *RecordsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetAuth(r.Context())
	if claims == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	tenantID := r.PathValue("tid")
	recordID := r.PathValue("rid")
	if tenantID == "" || recordID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_params"})
		return
	}

	if err := h.Service.DeleteRecord(r.Context(), recordID, claims.ProjectID, tenantID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "delete_failed"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
