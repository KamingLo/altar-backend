package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ClassRequest struct {
	NamaKelas   string `json:"nama_kelas" binding:"required"`
	Jurusan     string `json:"jurusan" binding:"required"`
	JumlahSiswa int    `json:"jumlah_siswa" binding:"required"`
}

func CreateClass(c *gin.Context) {
	var req ClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	class := models.Kelas{
		NamaKelas:   req.NamaKelas,
		Jurusan:     req.Jurusan,
		JumlahSiswa: req.JumlahSiswa,
	}

	if err := services.CreateClass(&class); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to create class", err)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Class created successfully", class)
}

func GetAllClasses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")

	data, total, err := services.GetAllClasses(page, limit, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch classes", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Classes fetched successfully", gin.H{
		"items":      data,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_page": (total + int64(limit) - 1) / int64(limit),
	})
}

func GetClassByID(c *gin.Context) {
	id := c.Param("id")
	class, err := services.GetClassByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Class not found", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Class fetched successfully", class)
}

func UpdateClass(c *gin.Context) {
	id := c.Param("id")
	var req ClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid request payload", err)
		return
	}

	updateData := map[string]interface{}{
		"nama_kelas":   req.NamaKelas,
		"jurusan":      req.Jurusan,
		"jumlah_siswa": req.JumlahSiswa,
	}

	class, err := services.UpdateClass(id, updateData)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to update class", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Class updated successfully", class)
}

func DeleteClass(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteClass(id); err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to delete class", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Class deleted successfully", nil)
}
