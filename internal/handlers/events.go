package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/vikhandev/events-platform/internal/db"
	"github.com/vikhandev/events-platform/internal/models"
)

type CreateEventRequest struct {
	Title        string    `json:"title" binding:"required,min=5"`
	Description  string    `json:"description" binding:"required"`
	Category     string    `json:"category" binding:"required"`
	StartTime    time.Time `json:"start_time" binding:"required"`
	LocationName string    `json:"location_name" binding:"required"`
	Latitude     float64   `json:"latitude" binding:"required"`
	Longitude    float64   `json:"longitude" binding:"required"`
	MaxCapacity  *int      `json:"max_capacity"`
	ImageURL     string    `json:"image_url"`
	Visibility   string    `json:"visibility"`
}

type EventResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Category      string  `json:"category"`
	StartTime     time.Time `json:"start_time"`
	LocationName  string  `json:"location_name"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	MaxCapacity   *int    `json:"max_capacity"`
	ImageURL      string  `json:"image_url"`
	CreatorID     string  `json:"creator_id"`
	AttendeeCount int64   `json:"attendee_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// CreateEvent handler
func CreateEvent(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.StartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time must be in the future"})
		return
	}

	event := models.Event{
		ID:           uuid.New(),
		Title:        req.Title,
		Description:  req.Description,
		Category:     req.Category,
		StartTime:    req.StartTime,
		LocationName: req.LocationName,
		Latitude:     req.Latitude,
		Longitude:    req.Longitude,
		MaxCapacity:  req.MaxCapacity,
		ImageURL:     req.ImageURL,
		Visibility:   req.Visibility,
		CreatorID:    userID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := db.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	// Auto-RSVP creator
	rsvp := models.RSVP{
		ID:        uuid.New(),
		UserID:    userID,
		EventID:   event.ID,
		Status:    "GOING",
		CreatedAt: time.Now(),
	}
	db.DB.Create(&rsvp)

	c.JSON(http.StatusCreated, EventResponse{
		ID:           event.ID.String(),
		Title:        event.Title,
		Description:  event.Description,
		Category:     event.Category,
		StartTime:    event.StartTime,
		LocationName: event.LocationName,
		Latitude:     event.Latitude,
		Longitude:    event.Longitude,
		CreatorID:    event.CreatorID.String(),
		AttendeeCount: 1,
		CreatedAt:    event.CreatedAt,
	})
}

// ListEvents with geolocation
func ListEvents(c *gin.Context) {
	category := c.Query("category")
	distance := c.DefaultQuery("distance", "25")
	latitude := c.Query("latitude")
	longitude := c.Query("longitude")
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)
	distanceFloat, _ := strconv.ParseFloat(distance, 64)

	query := db.DB

	// Filter by category
	if category != "" {
		query = query.Where("category = ?", category)
	}

	// Geolocation filter using PostGIS
	if latitude != "" && longitude != "" {
		lat, _ := strconv.ParseFloat(latitude, 64)
		lon, _ := strconv.ParseFloat(longitude, 64)
		distanceMeters := distanceFloat * 1000

		// PostGIS query - find events within distance
		query = query.Where(
			"ST_DistanceSphere(ST_Point(longitude, latitude), ST_Point(?, ?)) <= ?",
			lon, lat, distanceMeters,
		)
	}

	var total int64
	query.Model(&models.Event{}).Count(&total)

	var events []models.Event
	if err := query.
		Order("start_time ASC").
		Limit(limitInt).
		Offset(offsetInt).
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch events"})
		return
	}

	// Convert to response
	var eventResponses []EventResponse
	for _, event := range events {
		var attendeeCount int64
		db.DB.Model(&models.RSVP{}).
			Where("event_id = ? AND status = ?", event.ID, "GOING").
			Count(&attendeeCount)

		eventResponses = append(eventResponses, EventResponse{
			ID:            event.ID.String(),
			Title:         event.Title,
			Description:   event.Description,
			Category:      event.Category,
			StartTime:     event.StartTime,
			LocationName:  event.LocationName,
			Latitude:      event.Latitude,
			Longitude:     event.Longitude,
			CreatorID:     event.CreatorID.String(),
			AttendeeCount: attendeeCount,
			CreatedAt:     event.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"events": eventResponses,
		"total":  total,
		"limit":  limitInt,
		"offset": offsetInt,
	})
}

// GetEventDetail
func GetEventDetail(c *gin.Context) {
	eventID := c.Param("id")

	var event models.Event
	if err := db.DB.First(&event, "id = ?", eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	var attendeeCount int64
	db.DB.Model(&models.RSVP{}).
		Where("event_id = ? AND status = ?", event.ID, "GOING").
		Count(&attendeeCount)

	c.JSON(http.StatusOK, EventResponse{
		ID:            event.ID.String(),
		Title:         event.Title,
		Description:   event.Description,
		Category:      event.Category,
		StartTime:     event.StartTime,
		LocationName:  event.LocationName,
		Latitude:      event.Latitude,
		Longitude:     event.Longitude(),
		CreatorID:     event.CreatorID.String(),
		AttendeeCount: attendeeCount,
		CreatedAt:     event.CreatedAt,
	})
}

// UpdateEvent
func UpdateEvent(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)
	eventID := c.Param("id")

	var event models.Event
	if err := db.DB.First(&event, "id = ?", eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	if event.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only creator can update"})
		return
	}

	var req CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.Title = req.Title
	event.Description = req.Description
	event.Category = req.Category
	event.StartTime = req.StartTime
	event.LocationName = req.LocationName
	event.Latitude = req.Latitude
	event.Longitude = req.Longitude
	event.UpdatedAt = time.Now()

	if err := db.DB.Save(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event updated"})
}

// DeleteEvent
func DeleteEvent(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)
	eventID := c.Param("id")

	var event models.Event
	if err := db.DB.First(&event, "id = ?", eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	if event.CreatorID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only creator can delete"})
		return
	}

	if err := db.DB.Delete(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event deleted"})
}