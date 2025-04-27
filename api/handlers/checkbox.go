package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/usman-007/checkbox-backend/internal/services"
)

// CheckboxHandler handles HTTP requests related to checkboxes
type CheckboxHandler struct {
	checkboxService *services.CheckboxService
}

// NewCheckboxHandler creates a new instance of CheckboxHandler
func NewCheckboxHandler(checkboxService *services.CheckboxService) *CheckboxHandler {
	return &CheckboxHandler{
		checkboxService: checkboxService,
	}
}

// GetAllCheckboxes handles GET requests to get all checkboxes
func (h *CheckboxHandler) GetAllCheckboxes(c *gin.Context) {
	checkboxes, err := h.checkboxService.GetAllCheckboxes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get checkboxes: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, checkboxes)
}

// UpdateCheckbox handles PATCH requests to update checkbox state
func (h *CheckboxHandler) UpdateCheckbox(c *gin.Context) {
	// Extract query parameters
	rowStr := c.Query("row")
	columnStr := c.Query("column")
	valueStr := c.Query("value")

	// Validate parameters
	if rowStr == "" || columnStr == "" || valueStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Missing required parameters: row, column, and value are required",
		})
		return
	}

	// Convert row and column to integers
	row, err := strconv.Atoi(rowStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid row parameter: must be an integer",
		})
		return
	}

	column, err := strconv.Atoi(columnStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid column parameter: must be an integer",
		})
		return
	}

	// Parse boolean value
	value := false
	if valueStr == "true" {
		value = true
	} else if valueStr != "false" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid value parameter: must be 'true' or 'false'",
		})
		return
	}

	// Call service to update the checkbox state in Redis
	err = h.checkboxService.UpdateCheckboxState(uint32(row), uint32(column), value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update checkbox state: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Checkbox state updated successfully",
		"data": gin.H{
			"row":    row,
			"column": column,
			"value":  value,
		},
	})
}
