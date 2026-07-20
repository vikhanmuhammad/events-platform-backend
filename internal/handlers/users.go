package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/vikhanmuhammad/project-trainee/internal/db"
	"github.com/vikhanmuhammad/project-trainee/internal/models"
)

// GetUserProfile handler
func GetUserProfile(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"name":          user.Name,
		"avatar_url":    user.AvatarURL,
		"bio":           user.Bio,
		"location_name": user.LocationName,
		"latitude":      user.Latitude,
		"longitude":     user.Longitude,
		"interests":     user.Interests,
	})
}

// UpdateUserProfile handler
func UpdateUserProfile(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Only allow updating certain fields
	allowedFields := map[string]bool{
		"name": true, "avatar_url": true, "bio": true,
		"location_name": true, "latitude": true, "longitude": true,
		"interests": true,
	}

	for key, value := range updates {
		if !allowedFields[key] {
			continue
		}
		db.DB.Model(&user).Update(key, value)
	}

	user.UpdatedAt = time.Now()
	db.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

// GetUserUpcomingEvents handler
func GetUserUpcomingEvents(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)

	var rsvps []models.RSVP
	if err := db.DB.
		Preload("Event").
		Where("user_id = ?", userID).
		Find(&rsvps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}

	var events []map[string]interface{}
	now := time.Now()

	for _, rsvp := range rsvps {
		if rsvp.Event != nil && rsvp.Event.StartTime.After(now) {
			var attendeeCount int64
			db.DB.Model(&models.RSVP{}).
				Where("event_id = ? AND status = ?", rsvp.Event.ID, "GOING").
				Count(&attendeeCount)

			events = append(events, map[string]interface{}{
				"id":             rsvp.Event.ID.String(),
				"title":          rsvp.Event.Title,
				"description":    rsvp.Event.Description,
				"category":       rsvp.Event.Category,
				"start_time":     rsvp.Event.StartTime,
				"location_name":  rsvp.Event.LocationName,
				"latitude":       rsvp.Event.Latitude,
				"longitude":      rsvp.Event.Longitude,
				"creator_id":     rsvp.Event.CreatorID.String(),
				"attendee_count": attendeeCount,
				"your_status":    rsvp.Status,
				"created_at":     rsvp.Event.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

// GetUserPastEvents handler
func GetUserPastEvents(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)

	var rsvps []models.RSVP
	if err := db.DB.
		Preload("Event").
		Where("user_id = ?", userID).
		Find(&rsvps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}

	var events []map[string]interface{}
	now := time.Now()

	for _, rsvp := range rsvps {
		if rsvp.Event != nil && rsvp.Event.StartTime.Before(now) {
			var attendeeCount int64
			db.DB.Model(&models.RSVP{}).
				Where("event_id = ? AND status = ?", rsvp.Event.ID, "GOING").
				Count(&attendeeCount)

			events = append(events, map[string]interface{}{
				"id":             rsvp.Event.ID.String(),
				"title":          rsvp.Event.Title,
				"description":    rsvp.Event.Description,
				"category":       rsvp.Event.Category,
				"start_time":     rsvp.Event.StartTime,
				"location_name":  rsvp.Event.LocationName,
				"latitude":       rsvp.Event.Latitude,
				"longitude":      rsvp.Event.Longitude,
				"creator_id":     rsvp.Event.CreatorID.String(),
				"attendee_count": attendeeCount,
				"your_status":    rsvp.Status,
				"created_at":     rsvp.Event.CreatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
