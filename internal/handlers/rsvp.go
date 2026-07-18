package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/vikhandev/events-platform/internal/db"
	"github.com/vikhandev/events-platform/internal/models"
)

type RSVPRequest struct {
	Status string `json:"status" binding:"required,oneof=GOING INTERESTED CANT_GO NOT_RESPONDED"`
}

type RSVPResponse struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

// RSVPEvent handler
func RSVPEvent(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)

	eventID := c.Param("id")
	parsedEventID, err := uuid.Parse(eventID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}

	var req RSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify event exists
	var event models.Event
	if err := db.DB.First(&event, "id = ?", parsedEventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// Check or create RSVP
	var rsvp models.RSVP
	now := time.Now()

	if err := db.DB.Where("event_id = ? AND user_id = ?", parsedEventID, userID).
		First(&rsvp).Error; err != nil {
		// Create new RSVP
		rsvp = models.RSVP{
			ID:          uuid.New(),
			UserID:      userID,
			EventID:     parsedEventID,
			Status:      req.Status,
			RespondedAt: &now,
			CreatedAt:   time.Now(),
		}
		if err := db.DB.Create(&rsvp).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create RSVP"})
			return
		}
	} else {
		// Update existing RSVP
		rsvp.Status = req.Status
		rsvp.RespondedAt = &now
		if err := db.DB.Save(&rsvp).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update RSVP"})
			return
		}
	}

	c.JSON(http.StatusOK, RSVPResponse{
		ID:      rsvp.ID.String(),
		UserID:  rsvp.UserID.String(),
		EventID: rsvp.EventID.String(),
		Status:  rsvp.Status,
	})
}

// GetAttendees handler
func GetAttendees(c *gin.Context) {
	eventID := c.Param("id")

	var rsvps []models.RSVP
	if err := db.DB.
		Preload("User").
		Where("event_id = ? AND status = ?", eventID, "GOING").
		Find(&rsvps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch attendees"})
		return
	}

	var attendees []map[string]interface{}
	for _, rsvp := range rsvps {
		if rsvp.User != nil {
			attendees = append(attendees, map[string]interface{}{
				"id":     rsvp.User.ID.String(),
				"name":   rsvp.User.Name,
				"avatar": rsvp.User.AvatarURL,
			})
		}
	}

	var count int64
	db.DB.Model(&models.RSVP{}).
		Where("event_id = ? AND status = ?", eventID, "GOING").
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"attendees": attendees,
		"count":     count,
	})
}
