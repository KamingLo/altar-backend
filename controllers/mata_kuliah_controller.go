package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CourseRequest struct {
	NamaMK string `json:"nama_mk" binding:"required"`
	SKS    int    `json:"sks" binding:"required"`
}

func CreateCourse(c *gin.Context) {
	var req CourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	course := models.MataKuliah{
		NamaMK: req.NamaMK,
		SKS:    req.SKS,
	}

	if err := services.CreateCourse(&course); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create course", err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Course created successfully", course)
}

func GetAllCourses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	data, total, err := services.GetAllCourses(page, limit, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch courses", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Courses fetched successfully", gin.H{
		"items":      data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (total + int64(limit) - 1) / int64(limit),
	})
}

func GetCourseByID(c *gin.Context) {
	id := c.Param("id")
	course, err := services.GetCourseByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Course not found", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Course fetched successfully", course)
}

func UpdateCourse(c *gin.Context) {
	id := c.Param("id")
	var req CourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	updateData := map[string]interface{}{
		"nama_mk": req.NamaMK,
		"sks":     req.SKS,
	}

	course, err := services.UpdateCourse(id, updateData)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to update course", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Course updated successfully", course)
}

func DeleteCourse(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteCourse(id); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete course", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Course deleted successfully", nil)
}
