package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/vikhanmuhammad/project-trainee/internal/db"
	"github.com/vikhanmuhammad/project-trainee/internal/handlers"
	"github.com/vikhanmuhammad/project-trainee/internal/middleware"
)

func main() {
	// Load .env
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to database
	if err := db.Init(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{os.Getenv("FRONTEND_URL"), "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// Routes
	api := router.Group("/api")
	{
		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/signup", handlers.SignUp)
			auth.POST("/login", handlers.Login)
			auth.GET("/me", middleware.AuthRequired(), handlers.GetCurrentUser)
		}

		// Events
		events := api.Group("/events")
		{
			events.GET("", handlers.ListEvents)
			events.GET("/:id", handlers.GetEventDetail)
			events.POST("", middleware.AuthRequired(), handlers.CreateEvent)
			events.PUT("/:id", middleware.AuthRequired(), handlers.UpdateEvent)
			events.DELETE("/:id", middleware.AuthRequired(), handlers.DeleteEvent)
		}

		// RSVPs
		// RSVP
		events.GET("/:id/attendees", handlers.GetAttendees)
		events.POST("/:id/rsvp", middleware.AuthRequired(), handlers.RSVPEvent)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
