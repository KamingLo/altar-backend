package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SemesterRequest struct {
	TahunAjaran  string `json:"tahun_ajaran" binding:"required"`
	TipeSemester string `json:"tipe_semester" binding:"required"`
}

func CreateSemester(c *gin.Context) {
	var req SemesterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	semester := models.Semester{
		TahunAjaran:  req.TahunAjaran,
		TipeSemester: req.TipeSemester,
	}

	if err := services.CreateSemester(&semester); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create semester", err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Semester created successfully", semester)
}

func GetAllSemesters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	data, total, err := services.GetAllSemesters(page, limit, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch semesters", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Semesters fetched successfully", gin.H{
		"items":      data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (total + int64(limit) - 1) / int64(limit),
	})
}

func GetSemesterByID(c *gin.Context) {
	id := c.Param("id")
	semester, err := services.GetSemesterByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Semester not found", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Semester fetched successfully", semester)
}

func UpdateSemester(c *gin.Context) {
	id := c.Param("id")
	var req SemesterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	updateData := map[string]interface{}{
		"tahun_ajaran":  req.TahunAjaran,
		"tipe_semester": req.TipeSemester,
	}

	semester, err := services.UpdateSemester(id, updateData)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to update semester", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Semester updated successfully", semester)
}

func DeleteSemester(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteSemester(id); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete semester", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Semester deleted successfully", nil)
}
