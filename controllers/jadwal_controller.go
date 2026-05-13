package controllers

import (
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	semesterID := c.Query("id_semester")

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	data, total, err := services.GetAllSessions(page, limit, semesterID)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch sessions", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Sessions fetched successfully", gin.H{
		"items":      data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (total + int64(limit) - 1) / int64(limit),
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
