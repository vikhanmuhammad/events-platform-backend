package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vikhanmuhammad/project-trainee/internal/models"
)

var DB *gorm.DB

func Init() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable PostGIS extension
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS postgis").Error; err != nil {
		log.Println("Warning: PostGIS extension not enabled (may need manual setup)")
	}

	log.Println("✅ Database connected successfully")
	return nil
}

func Migrate() error {
	return DB.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.RSVP{},
		&models.Comment{},
		&models.Notification{},
	)
}

func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
