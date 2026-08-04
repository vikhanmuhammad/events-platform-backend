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
	getUserEventsByTime(c, true)
}

// GetUserPastEvents handler
func GetUserPastEvents(c *gin.Context) {
	getUserEventsByTime(c, false)
}

// getUserEventsByTime fetches the user's RSVP'd events split by whether
// they're upcoming or past, batching attendee counts into a single grouped
// query instead of one COUNT query per event.
func getUserEventsByTime(c *gin.Context, upcoming bool) {
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

	now := time.Now()
	var relevant []models.RSVP
	var eventIDs []uuid.UUID
	for _, rsvp := range rsvps {
		if rsvp.Event == nil {
			continue
		}
		isUpcoming := rsvp.Event.StartTime.After(now)
		if isUpcoming != upcoming {
			continue
		}
		relevant = append(relevant, rsvp)
		eventIDs = append(eventIDs, rsvp.Event.ID)
	}

	attendeeCounts := make(map[uuid.UUID]int64)
	if len(eventIDs) > 0 {
		var rows []struct {
			EventID uuid.UUID
			Count   int64
		}
		db.DB.Model(&models.RSVP{}).
			Select("event_id, count(*) as count").
			Where("event_id IN ? AND status = ?", eventIDs, "GOING").
			Group("event_id").
			Scan(&rows)
		for _, row := range rows {
			attendeeCounts[row.EventID] = row.Count
		}
	}

	var events []map[string]interface{}
	for _, rsvp := range relevant {
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
			"attendee_count": attendeeCounts[rsvp.Event.ID],
			"your_status":    rsvp.Status,
			"created_at":     rsvp.Event.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}
