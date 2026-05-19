package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────
// Request DTO: CreateSubstituteSessionRequest
// Used for POST /substitute-sessions
// ─────────────────────────────────────────────

type CreateSubstituteSessionRequest struct {
	IDSession      string  `json:"id_session"            binding:"required"`
	IDRuangan      string  `json:"id_ruangan"            binding:"required"`
	IDDosen        *string `json:"id_dosen"`
	IDAsdos1       *string `json:"id_asdos1"`
	IDAsdos2       *string `json:"id_asdos2"`
	SubstituteDate string  `json:"substitute_date"       binding:"required"` // YYYY-MM-DD
	OriginalDate   string  `json:"original_date"         binding:"required"` // YYYY-MM-DD
	SlotOption     int     `json:"slot_option"           binding:"required,min=1,max=7"`
	Reason         string  `json:"reason"                binding:"required"`
}

// ─────────────────────────────────────────────
// Request DTO: UpdateSubstituteStatusRequest
// Used for PATCH /substitute-sessions/:id/status
// ─────────────────────────────────────────────

type UpdateSubstituteStatusRequest struct {
	Status            string  `json:"status" binding:"required"`
	CoordinatorReason *string `json:"coordinator_reason"`
}

// ─────────────────────────────────────────────
// POST /substitute-sessions
// Role: Asdos — submit a replacement class request
// ─────────────────────────────────────────────

func CreateSubstituteSession(c *gin.Context) {
	var req CreateSubstituteSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload or missing required fields", err)
		return
	}

	
	input := &services.SubstituteSessionInput{
		IDSession:      req.IDSession,
		IDRuangan:      req.IDRuangan,
		IDDosen:        req.IDDosen,
		IDAsdos1:       req.IDAsdos1,
		IDAsdos2:       req.IDAsdos2,
		SubstituteDate: req.SubstituteDate,
		OriginalDate:   req.OriginalDate,
		SlotOption:     req.SlotOption,
		Reason:         req.Reason,
	}

	resp, err := services.CreateSubstituteSession(input)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to create substitute session", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Substitute session submitted successfully and is pending approval", resp)
}

// ─────────────────────────────────────────────
// GET /substitute-sessions?status=PENDING
// Role: Koordinator — list all requests, filterable by status
// ─────────────────────────────────────────────

func GetAllSubstituteSessions(c *gin.Context) {
	statusFilter := c.Query("status")

	data, err := services.GetAllSubstituteSessions(statusFilter)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch substitute sessions", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Substitute sessions fetched successfully", gin.H{
		"items": data,
		"total": len(data),
	})
}

// ─────────────────────────────────────────────
// GET /substitute-sessions/:id
// Role: Read details of a substitute session
// ─────────────────────────────────────────────

func GetSubstituteSessionByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := services.GetSubstituteSessionByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Substitute session not found", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Substitute session fetched successfully", resp)
}

// ─────────────────────────────────────────────
// DELETE /substitute-sessions/:id
// Role: Soft delete a substitute session (only if PENDING)
// ─────────────────────────────────────────────

func DeleteSubstituteSession(c *gin.Context) {
	id := c.Param("id")

	if err := services.DeleteSubstituteSession(id); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to delete substitute session", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Substitute session deleted successfully", nil)
}

// ─────────────────────────────────────────────
// PATCH /substitute-sessions/:id/status
// Role: Koordinator — approve or reject a pending request
// ─────────────────────────────────────────────

func UpdateSubstituteStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateSubstituteStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload or missing required fields", err)
		return
	}

	input := &services.UpdateSubstituteStatusInput{
		Status:            models.SubstituteSessionStatus(req.Status),
		CoordinatorReason: req.CoordinatorReason,
	}

	resp, err := services.UpdateSubstituteStatus(id, input)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to update substitute session status", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Substitute session status updated successfully", resp)
}

// ─────────────────────────────────────────────
// GET /jadwal/sessions?start_date=...&end_date=...&id_semester=...
// Role: Any authenticated user (global view of all sessions)
// Returns all sessions across all asdos for the given semester and date range.
// ─────────────────────────────────────────────

func GetScheduleTimeline(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	idSemester := c.Query("id_semester")

	if startDate == "" || endDate == "" || idSemester == "" {
		utils.SendError(c, http.StatusBadRequest,
			"Missing required query parameters",
			"'start_date', 'end_date', and 'id_semester' are required (format: YYYY-MM-DD)")
		return
	}

	// asdosID = "" means global: no filtering by a specific teacher
	data, err := services.GetTimelineJadwal(startDate, endDate, idSemester, "")
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to generate session timeline", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Session timeline fetched successfully", gin.H{
		"start_date":  startDate,
		"end_date":    endDate,
		"id_semester": idSemester,
		"total":       len(data),
		"items":       data,
	})
}

// ─────────────────────────────────────────────
// GET /jadwal/my-sessions?start_date=...&end_date=...&id_semester=...
// Role: Asdos only (personal view — filtered to the requesting asdos)
// Extracts asdos_id from Gin context (set by IsAsdosMiddleware after JWT validation).
// ─────────────────────────────────────────────

func GetMyScheduleTimeline(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	idSemester := c.Query("id_semester")

	if startDate == "" || endDate == "" || idSemester == "" {
		utils.SendError(c, http.StatusBadRequest,
			"Missing required query parameters",
			"'start_date', 'end_date', and 'id_semester' are required (format: YYYY-MM-DD)")
		return
	}

	// Extract the asdos ID injected by IsAsdosMiddleware (originates from JWT claims)
	asdosIDRaw, exists := c.Get("id_asisten")
	if !exists || asdosIDRaw == nil {
		utils.SendError(c, http.StatusUnauthorized, "Asdos identity not found in token", nil)
		return
	}
	asdosID, ok := asdosIDRaw.(string)
	if !ok || asdosID == "" {
		utils.SendError(c, http.StatusUnauthorized, "Invalid asdos identity in token", nil)
		return
	}

	data, err := services.GetTimelineJadwal(startDate, endDate, idSemester, asdosID)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to generate personal session timeline", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Personal session timeline fetched successfully", gin.H{
		"start_date":  startDate,
		"end_date":    endDate,
		"id_semester": idSemester,
		"asdos_id":    asdosID,
		"total":       len(data),
		"items":       data,
	})
}
