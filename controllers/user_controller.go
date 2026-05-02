package controllers

import (
	"altar/models"
	"altar/services"
	"altar/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// --- Asisten Dosen Handlers ---

func CreateAsdos(c *gin.Context) {
	var input struct {
		models.User
		NIM         string `json:"nim" binding:"required"`
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Data is not complete", err)
		return
	}

	if err := services.CreateAsdos(&input.User, input.NIM, input.PhoneNumber); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Asisten Dosen created successfully", nil)
}

func GetAllAsdos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	search := c.Query("search")

	asdos, err := services.GetAllAsdos(page, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch data", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", asdos)
}

func GetAsdosByID(c *gin.Context) {
	id := c.Param("id")
	asdos, err := services.GetAsdosByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Asisten Dosen not found", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", asdos)
}

func UpdateAsdos(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NIM         string `json:"nim" binding:"required"`
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := services.UpdateAsdos(id, input.NIM, input.PhoneNumber); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Asisten Dosen updated successfully", nil)
}

func DeleteAsdos(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteAsdos(id); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Asisten Dosen deleted successfully", nil)
}

// --- Koordinator Handlers ---

func CreateKoordinator(c *gin.Context) {
	var input struct {
		models.User
		NIP string `json:"nip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Data is not complete", err)
		return
	}

	if err := services.CreateKoordinator(&input.User, input.NIP); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusCreated, "Koordinator created successfully", nil)
}

func GetAllKoordinator(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	search := c.Query("search")

	koor, err := services.GetAllKoordinator(page, search)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to fetch data", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", koor)
}

func GetKoordinatorByID(c *gin.Context) {
	id := c.Param("id")
	koor, err := services.GetKoordinatorByID(id)
	if err != nil {
		utils.SendError(c, http.StatusNotFound, "Koordinator not found", err)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Success", koor)
}

func UpdateKoordinator(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		NIP string `json:"nip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	if err := services.UpdateKoordinator(id, input.NIP); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Koordinator updated successfully", nil)
}

func DeleteKoordinator(c *gin.Context) {
	id := c.Param("id")
	if err := services.DeleteKoordinator(id); err != nil {
		utils.SendError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	utils.SendSuccess(c, http.StatusOK, "Koordinator deleted successfully", nil)
}
