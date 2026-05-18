package controllers

import (
	"altar/services"
	"altar/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────
// Request DTO
// ─────────────────────────────────────────────

type SessionRequest struct {
	IDKelas    string  `json:"id_kelas"    binding:"required"`
	IDMk       string  `json:"id_mk"       binding:"required"`
	IDSemester string  `json:"id_semester" binding:"required"`
	IDRuangan  string  `json:"id_ruangan"  binding:"required"`
	IDAsdos1   *string `json:"id_asdos1"`
	IDAsdos2   *string `json:"id_asdos2"`
	IDDosen    *string `json:"id_dosen"`
	DayOption  int     `json:"opsi_hari"   binding:"required,min=1,max=6"`
	SlotOption int     `json:"opsi_jam"    binding:"required,min=1,max=7"`
}

// ─────────────────────────────────────────────
// Internal: toSessionInput
// Maps the HTTP request DTO to the service input struct.
// ─────────────────────────────────────────────

func toSessionInput(req *SessionRequest) *services.SessionInput {
	return &services.SessionInput{
		IDKelas:    req.IDKelas,
		IDMk:       req.IDMk,
		IDSemester: req.IDSemester,
		IDRuangan:  req.IDRuangan,
		IDAsdos1:   req.IDAsdos1,
		IDAsdos2:   req.IDAsdos2,
		IDDosen:    req.IDDosen,
		DayOption:  req.DayOption,
		SlotOption: req.SlotOption,
	}
}

// ─────────────────────────────────────────────
// POST /sessions
// ─────────────────────────────────────────────

func CreateSession(c *gin.Context) {
	var req SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload or missing required fields", err)
		return
	}

	resp, err := services.CreateSession(toSessionInput(&req))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to create session", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Session created successfully", resp)
}

// ─────────────────────────────────────────────
// GET /sessions?page=1&limit=10&id_semester=...
// ─────────────────────────────────────────────

func GetAllSessions(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	semesterID := c.Query("id_semester")

	if semesterID == "" {
		utils.SendError(c, http.StatusBadRequest, "Missing parameter", "'id_semester' is required")
		return
	}

	// Default to a 7-day timeline starting today if dates are not provided
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	}

	data, err := services.GetTimelineJadwal(startDate, endDate, semesterID, "")
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch sessions timeline", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Sessions fetched successfully", gin.H{
		"start_date":  startDate,
		"end_date":    endDate,
		"id_semester": semesterID,
		"total":       len(data),
		"items":       data,
	})
}

// ─────────────────────────────────────────────
// GET /sessions/:id
// ─────────────────────────────────────────────

func GetSessionByID(c *gin.Context) {
	id := c.Param("id")

	resp, err := services.GetSessionByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Session not found", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Session fetched successfully", resp)
}

// ─────────────────────────────────────────────
// PATCH /sessions/:id
// ─────────────────────────────────────────────

func UpdateSession(c *gin.Context) {
	id := c.Param("id")

	var req SessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload or missing required fields", err)
		return
	}

	resp, err := services.UpdateSession(id, toSessionInput(&req))
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to update session", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Session updated successfully", resp)
}

// ─────────────────────────────────────────────
// DELETE /sessions/:id
// ─────────────────────────────────────────────

func DeleteSession(c *gin.Context) {
	id := c.Param("id")

	if err := services.DeleteSession(id); err != nil {
		utils.SendError(c, http.StatusNotFound, "Failed to delete session", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Session deleted successfully", nil)
}

// ─────────────────────────────────────────────
// GET /sessions/me
// ─────────────────────────────────────────────

func GetMySession(c *gin.Context) {
	idAsistenRaw, exists := c.Get("id_asisten")
	if !exists || idAsistenRaw == nil {
		utils.SendError(c, http.StatusForbidden, "Forbidden", "id_asisten not found in context")
		return
	}
	idAsisten, ok := idAsistenRaw.(string)
	if !ok || idAsisten == "" {
		utils.SendError(c, http.StatusForbidden, "Forbidden", "invalid id_asisten type")
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		dateStr = time.Now().Format("2006-01-02")
	}

	resp, err := services.GetDailyAssistantSessions(dateStr, idAsisten)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch daily sessions", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Daily sessions fetched successfully", resp)
}

