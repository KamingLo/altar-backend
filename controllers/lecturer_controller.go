package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type LecturerRequest struct {
	NIP  string `json:"nip" binding:"required"`
	Nama string `json:"nama" binding:"required"`
}

func CreateLecturer(c *gin.Context) {
	var req LecturerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	lecturer := models.Dosen{
		NIP:  req.NIP,
		Nama: req.Nama,
	}

	if err := services.CreateLecturer(&lecturer); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create lecturer", err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Lecturer created successfully", lecturer)
}

func GetAllLecturers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	data, total, err := services.GetAllLecturers(page, limit, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch lecturers", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Lecturers fetched successfully", gin.H{
		"items":      data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (total + int64(limit) - 1) / int64(limit),
	})
}

func GetLecturerByID(c *gin.Context) {
	id := c.Param("id")
	lecturer, err := services.GetLecturerByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Lecturer not found", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Lecturer fetched successfully", lecturer)
}

func UpdateLecturer(c *gin.Context) {
	id := c.Param("id")
	var req LecturerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	updateData := map[string]interface{}{
		"nip":  req.NIP,
		"nama": req.Nama,
	}

	lecturer, err := services.UpdateLecturer(id, updateData)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to update lecturer", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Lecturer updated successfully", lecturer)
}

func DeleteLecturer(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteLecturer(id); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete lecturer", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Lecturer deleted successfully", nil)
}
