package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/vikhandev/events-platform/internal/db"
	"github.com/vikhandev/events-platform/internal/models"
)

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}

// GetComments handler
func GetComments(c *gin.Context) {
	eventID := c.Param("id")

	var comments []models.Comment
	if err := db.DB.
		Preload("User").
		Where("event_id = ?", eventID).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch comments"})
		return
	}

	var responses []map[string]interface{}
	for _, comment := range comments {
		responses = append(responses, map[string]interface{}{
			"id":      comment.ID.String(),
			"user":    comment.User.Name,
			"content": comment.Content,
			"created": comment.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"comments": responses})
}

// AddComment handler
func AddComment(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)

	eventID := c.Param("id")
	parsedEventID, _ := uuid.Parse(eventID)

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment := models.Comment{
		ID:        uuid.New(),
		UserID:    userID,
		EventID:   parsedEventID,
		Content:   req.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      comment.ID.String(),
		"content": comment.Content,
		"created": comment.CreatedAt,
	})
}

// DeleteComment handler
func DeleteComment(c *gin.Context) {
	userIDStr := c.GetString("userID")
	userID, _ := uuid.Parse(userIDStr)

	commentID := c.Param("commentId")
	parsedCommentID, _ := uuid.Parse(commentID)

	var comment models.Comment
	if err := db.DB.First(&comment, "id = ?", parsedCommentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
		return
	}

	if comment.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "only author can delete"})
		return
	}

	if err := db.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "comment deleted"})
}
