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
	IDSession      string `json:"id_session"       binding:"required"`
	IDRuangan      string `json:"id_ruangan"       binding:"required"`
	SubstituteDate string `json:"substitute_date"  binding:"required"` // YYYY-MM-DD
	SlotOption     int    `json:"slot_option"      binding:"required,min=1,max=7"`
	Reason         string `json:"reason"           binding:"required"`
}

// ─────────────────────────────────────────────
// Request DTO: UpdateSubstituteStatusRequest
// Used for PATCH /substitute-sessions/:id/status
// ─────────────────────────────────────────────

type UpdateSubstituteStatusRequest struct {
	Status string `json:"status" binding:"required"`
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
		SubstituteDate: req.SubstituteDate,
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
		Status: models.SubstituteSessionStatus(req.Status),
	}

	resp, err := services.UpdateSubstituteStatus(id, input)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to update substitute session status", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Substitute session status updated successfully", resp)
}
