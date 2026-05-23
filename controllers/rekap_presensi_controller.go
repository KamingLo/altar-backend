package controllers

import (
	"altar/services"
	"altar/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetRekapPresensiMe(c *gin.Context) {
	asdosID := c.GetString("id_asisten")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.SendError(c, http.StatusBadRequest, "start_date and end_date are required (YYYY-MM-DD)", nil)
		return
	}

	startDate, err1 := time.Parse("2006-01-02", startDateStr)
	endDate, err2 := time.Parse("2006-01-02", endDateStr)

	if err1 != nil || err2 != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", nil)
		return
	}

	res, err := services.GetRekapPresensi(asdosID, startDate, endDate)
	if err != nil {
		utils.SendError(c, http.StatusInternalServerError, "Failed to calculate attendance recap", err)
		return
	}

	utils.SendSuccess(c, http.StatusOK, "Attendance recap fetched successfully", res)
}

func GetRekapPresensi(c *gin.Context) {
	asdosID := c.Query("asdos_id")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		utils.SendError(c, http.StatusBadRequest, "start_date and end_date are required (YYYY-MM-DD)", nil)
		return
	}

	startDate, err1 := time.Parse("2006-01-02", startDateStr)
	endDate, err2 := time.Parse("2006-01-02", endDateStr)

	if err1 != nil || err2 != nil {
		utils.SendError(c, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD", nil)
		return
	}

	if asdosID == "" {
		res, err := services.GetAllRekapPresensi(startDate, endDate)
		if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Failed to calculate all attendance recaps", err)
			return
		}
		utils.SendSuccess(c, http.StatusOK, "All attendance recaps fetched successfully", res)
	} else {
		res, err := services.GetRekapPresensi(asdosID, startDate, endDate)
		if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "Failed to calculate attendance recap", err)
			return
		}
		utils.SendSuccess(c, http.StatusOK, "Attendance recap fetched successfully", res)
	}
}
