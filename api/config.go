package api

import (
	"encoding/json"
	"net/http"
	"src/helpers"
	"src/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetConfig(dbConn *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var config models.Config
		if err := dbConn.First(&config).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				helpers.FormatErrorResponse(c, http.StatusNotFound, "Config not found")
			} else {
				helpers.FormatErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve config")
			}
			return
		}
		helpers.FormatSuccessResponse(c, config)
	}
}

func SetLoginNotCmu(dbConn *gorm.DB, hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := new(struct {
			LoginNotCmu bool `json:"loginNotCmu"`
		})
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := dbConn.Model(&models.Config{}).Where("id = ?", 1).Update("LoginNotCmu", body.LoginNotCmu).Error; err != nil {
			helpers.FormatErrorResponse(c, http.StatusInternalServerError, "Failed to update config")
			return
		}

		message, _ := json.Marshal(map[string]interface{}{
			"event": "setLoginNotCmu",
			"data":  body.LoginNotCmu,
		})
		hub.broadcast <- message

		helpers.FormatSuccessResponse(c, map[string]interface{}{"message": "Config updated successfully"})
	}
}
