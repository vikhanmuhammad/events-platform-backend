package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string         `gorm:"index;uniqueIndex" json:"email"`
	PasswordHash string         `gorm:"-" json:"-"`
	Name         string         `json:"name"`
	AvatarURL    string         `json:"avatar_url"`
	Bio          string         `json:"bio"`
	LocationName string         `json:"location_name"`
	Latitude     float64        `json:"latitude"`
	Longitude    float64        `json:"longitude"`
	Interests    pq.StringArray `gorm:"type:text[]" json:"interests"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`

	Events   []Event   `gorm:"foreignKey:CreatorID" json:"-"`
	RSVPs    []RSVP    `gorm:"foreignKey:UserID" json:"-"`
	Comments []Comment `gorm:"foreignKey:UserID" json:"-"`
}

type Event struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	Title             string     `gorm:"index" json:"title"`
	Description       string     `json:"description"`
	Category          string     `gorm:"index" json:"category"`
	StartTime         time.Time  `gorm:"index" json:"start_time"`
	EndTime           *time.Time `json:"end_time"`
	LocationName      string     `json:"location_name"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
	MaxCapacity       *int       `json:"max_capacity"`
	ImageURL          string     `json:"image_url"`
	Visibility        string     `json:"visibility"`
	CreatorID         uuid.UUID  `gorm:"index" json:"creator_id"`
	Creator           *User      `json:"creator,omitempty"`
	IsRecurring       bool       `json:"is_recurring"`
	RecurrencePattern string     `json:"recurrence_pattern"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	RSVPs    []RSVP    `gorm:"foreignKey:EventID" json:"-"`
	Comments []Comment `gorm:"foreignKey:EventID" json:"-"`
}

type RSVP struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID  `gorm:"index" json:"user_id"`
	User        *User      `json:"user,omitempty"`
	EventID     uuid.UUID  `gorm:"index" json:"event_id"`
	Event       *Event     `json:"-"`
	Status      string     `json:"status"`
	RespondedAt *time.Time `json:"responded_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"index" json:"user_id"`
	User      *User     `json:"user,omitempty"`
	EventID   uuid.UUID `gorm:"index" json:"event_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Notification struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"index" json:"user_id"`
	User      *User      `json:"-"`
	EventID   *uuid.UUID `json:"event_id"`
	Event     *Event     `json:"-"`
	Type      string     `json:"type"`
	Message   string     `json:"message"`
	IsRead    bool       `json:"is_read"`
	CreatedAt time.Time  `json:"created_at"`
}
