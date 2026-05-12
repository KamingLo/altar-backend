package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateSessionRequest struct {
	IDKelas       string  `json:"id_kelas" binding:"required"`
	IDMk          string  `json:"id_mk" binding:"required"`
	IDSemester    string  `json:"id_semester" binding:"required"`
	IDRuangan     string  `json:"id_ruangan" binding:"required"`
	IDAsdos1      *string `json:"id_asdos1"`
	IDAsdos2      *string `json:"id_asdos2"`
	IDDosen       *string `json:"id_dosen"`
	KelasMulai    string  `json:"kelas_mulai" binding:"required"`
	KelasBerakhir string  `json:"kelas_berakhir" binding:"required"`
}

func CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload or missing required fields", err)
		return
	}

	// Handle empty string input as nil for optional fields
	if req.IDDosen != nil && *req.IDDosen == "" {
		req.IDDosen = nil
	}
	if req.IDAsdos1 != nil && *req.IDAsdos1 == "" {
		req.IDAsdos1 = nil
	}
	if req.IDAsdos2 != nil && *req.IDAsdos2 == "" {
		req.IDAsdos2 = nil
	}

	// Time Parsing (RFC3339)
	startTime, err := time.Parse(time.RFC3339, req.KelasMulai)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid 'kelas_mulai' format. Use ISO8601/RFC3339", err)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.KelasBerakhir)
	if err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid 'kelas_berakhir' format. Use ISO8601/RFC3339", err)
		return
	}

	// Logic Validation
	if endTime.Before(startTime) || endTime.Equal(startTime) {
		utils.SendError(c, http.StatusBadRequest, "Class end time must be after start time", nil)
		return
	}

	session := models.JadwalUtama{
		IDKelas:       req.IDKelas,
		IDMk:          req.IDMk,
		IDSemester:    req.IDSemester,
		IDRuangan:     req.IDRuangan,
		IDAsdos1:      req.IDAsdos1,
		IDAsdos2:      req.IDAsdos2,
		IDDosen:       req.IDDosen,
		KelasMulai:    startTime,
		KelasBerakhir: endTime,
	}

	if err := services.CreateSession(&session); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Failed to create session", err.Error())
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Session created successfully", session)
}
