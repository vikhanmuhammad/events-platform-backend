package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/vikhanmuhammad/project-trainee/internal/db"
	"github.com/vikhanmuhammad/project-trainee/internal/models"
)

// Run with: go run ./cmd/seed   (from the backend/ directory)
// Safe to re-run: if the first seed user already exists, seeding is skipped.
func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	if err := db.Init(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	var existing models.User
	if err := db.DB.Where("email = ?", seedUsers[0].email).First(&existing).Error; err == nil {
		log.Println("Seed data already present (found", seedUsers[0].email, ") - skipping.")
		return
	}

	users := seedUserRows()
	events := seedEventRows(users)
	seedRSVPs(users, events)
	seedComments(users, events)
	seedNotifications(users, events)

	fmt.Printf(
		"Seed complete: %d users, %d events, RSVPs, comments, and notifications.\n",
		len(users), len(events),
	)
}

type userSeed struct {
	email        string
	name         string
	bio          string
	locationName string
	latitude     float64
	longitude    float64
	interests    []string
}

var seedUsers = []userSeed{
	{"budi.santoso@example.com", "Budi Santoso", "Software engineer who loves community meetups.", "Jakarta Selatan", -6.2615, 106.7810, []string{"Tech", "Music"}},
	{"siti.nurhaliza@example.com", "Siti Nurhaliza", "Weekend runner and yoga enthusiast.", "Jakarta Pusat", -6.1862, 106.8347, []string{"Sports", "Social"}},
	{"andi.wijaya@example.com", "Andi Wijaya", "Coffee addict, occasional painter.", "Bandung", -6.9175, 107.6191, []string{"Art", "Food"}},
	{"dewi.lestari@example.com", "Dewi Lestari", "Bookworm and language exchange host.", "Jakarta Barat", -6.1683, 106.7588, []string{"Social", "Art"}},
	{"rizky.pratama@example.com", "Rizky Pratama", "Backend developer, futsal every weekend.", "Jakarta Timur", -6.2250, 106.9004, []string{"Tech", "Sports"}},
	{"maya.sari@example.com", "Maya Sari", "Photographer chasing golden hour.", "Yogyakarta", -7.7956, 110.3695, []string{"Art", "Music"}},
	{"fajar.hidayat@example.com", "Fajar Hidayat", "Street food explorer.", "Surabaya", -7.2575, 112.7521, []string{"Food", "Social"}},
	{"putri.ayu@example.com", "Putri Ayu", "Indie musician and jazz lover.", "Jakarta Selatan", -6.2897, 106.7997, []string{"Music", "Art"}},
	{"agus.setiawan@example.com", "Agus Setiawan", "Marathon trainee, data engineer by day.", "Jakarta Utara", -6.1214, 106.8804, []string{"Sports", "Tech"}},
	{"nadia.rahma@example.com", "Nadia Rahma", "Volunteer organizer for beach cleanups.", "Jakarta Selatan", -6.2441, 106.8006, []string{"Social", "Sports"}},
	{"yoga.pranata@example.com", "Yoga Pranata", "Startup founder, workshop junkie.", "Bandung", -6.8951, 107.6107, []string{"Tech", "Food"}},
	{"lina.marlina@example.com", "Lina Marlina", "Watercolor hobbyist and cat mom.", "Jakarta Pusat", -6.1751, 106.8650, []string{"Art", "Social"}},
}

func seedUserRows() []models.User {
	hashed, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash seed password: %v", err)
	}

	var users []models.User
	for _, u := range seedUsers {
		user := models.User{
			ID:           uuid.New(),
			Email:        u.email,
			PasswordHash: string(hashed),
			Name:         u.name,
			Bio:          u.bio,
			LocationName: u.locationName,
			Latitude:     u.latitude,
			Longitude:    u.longitude,
			Interests:    u.interests,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := db.DB.Create(&user).Error; err != nil {
			log.Fatalf("failed to seed user %s: %v", u.email, err)
		}
		users = append(users, user)
	}
	return users
}

type eventSeed struct {
	title        string
	description  string
	category     string
	dayOffset    int // relative to now; negative = past
	locationName string
	latitude     float64
	longitude    float64
	maxCapacity  *int
}

func intPtr(v int) *int { return &v }

var seedEvents = []eventSeed{
	{"Golang Meetup Jakarta", "Monthly meetup for Go developers to share what they've been building.", "Tech", 5, "Kopi Kenangan Hub, Jakarta Selatan", -6.2608, 106.7815, intPtr(60)},
	{"React Conference Indonesia 2026", "A full day of talks on React, Next.js, and the modern frontend ecosystem.", "Tech", 20, "Balai Kartini, Jakarta", -6.2213, 106.8286, intPtr(300)},
	{"AI & Machine Learning Workshop", "Hands-on workshop covering practical ML applications for product teams.", "Tech", 10, "Conclave Coworking, Jakarta", -6.2088, 106.8230, intPtr(50)},
	{"Startup Weekend Jakarta", "54 hours to build a startup from scratch with mentors and investors.", "Tech", -15, "GoWork Plaza, Jakarta", -6.2244, 106.8090, nil},
	{"Docker & Kubernetes Bootcamp", "Deep dive into containers and orchestration for backend engineers.", "Tech", 30, "Dago Creative Hub, Bandung", -6.8925, 107.6120, intPtr(40)},

	{"Weekend Futsal Tournament", "Casual 5-a-side futsal tournament, all skill levels welcome.", "Sports", 3, "GOR Bulungan, Jakarta Selatan", -6.2415, 106.8087, intPtr(80)},
	{"Jakarta Marathon Community Run", "Group training run ahead of the city marathon.", "Sports", 45, "Gelora Bung Karno, Jakarta", -6.2189, 106.8021, nil},
	{"Badminton Club Meetup", "Weekly badminton session, bring your own racket or borrow one.", "Sports", -7, "GOR Ciracas, Jakarta Timur", -6.3312, 106.8794, intPtr(24)},
	{"Yoga in the Park", "Sunrise yoga session for all levels, mats provided.", "Sports", 2, "Taman Tegallega, Bandung", -6.9280, 107.6009, intPtr(30)},

	{"Indie Music Night", "Local indie bands performing live at an intimate venue.", "Music", 7, "Rossi Musik, Jakarta Selatan", -6.2352, 106.7910, intPtr(150)},
	{"Jazz Under The Stars", "Outdoor jazz performance featuring local musicians.", "Music", 14, "Museum Fatahillah Courtyard, Jakarta", -6.1352, 106.8133, nil},
	{"Acoustic Sunday Sessions", "Chill acoustic sets, open mic slots available.", "Music", -3, "Kopi Toko Djawa, Jakarta", -6.2440, 106.7972, intPtr(40)},

	{"Contemporary Art Exhibition", "Exhibition featuring emerging Indonesian contemporary artists.", "Art", 21, "Museum MACAN, Jakarta Barat", -6.1727, 106.7913, nil},
	{"Watercolor Painting Workshop", "Beginner-friendly watercolor workshop, materials included.", "Art", 9, "Taman Budaya Yogyakarta", -7.8014, 110.3644, intPtr(20)},
	{"Street Photography Walk", "Guided walk through old town capturing street life.", "Art", -10, "Kota Tua, Jakarta", -6.1352, 106.8133, intPtr(15)},

	{"Street Food Festival", "A celebration of Jakarta's best street food vendors in one place.", "Food", 6, "Lapangan Banteng, Jakarta", -6.1706, 106.8330, nil},
	{"Coffee Cupping Session", "Learn to taste and evaluate specialty coffee like a pro.", "Food", 4, "Anomali Coffee, Jakarta Selatan", -6.2495, 106.7986, intPtr(15)},
	{"Cooking Class: Nusantara Cuisine", "Hands-on class cooking classic Indonesian dishes.", "Food", 12, "Dapur Nusantara, Surabaya", -7.2650, 112.7460, intPtr(18)},

	{"Book Club Monthly Meetup", "This month: discussing Indonesian contemporary fiction.", "Social", 8, "Reading Room Cafe, Jakarta", -6.1963, 106.8231, intPtr(25)},
	{"Volunteer Beach Cleanup", "Community cleanup event, gloves and bags provided.", "Social", 25, "Ancol Beach, Jakarta Utara", -6.1223, 106.8317, nil},
}

func seedEventRows(users []models.User) []models.Event {
	var events []models.Event
	for i, e := range seedEvents {
		creator := users[i%len(users)]
		event := models.Event{
			ID:           uuid.New(),
			Title:        e.title,
			Description:  e.description,
			Category:     e.category,
			StartTime:    time.Now().AddDate(0, 0, e.dayOffset),
			LocationName: e.locationName,
			Latitude:     e.latitude,
			Longitude:    e.longitude,
			MaxCapacity:  e.maxCapacity,
			Visibility:   "public",
			CreatorID:    creator.ID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := db.DB.Create(&event).Error; err != nil {
			log.Fatalf("failed to seed event %s: %v", e.title, err)
		}
		events = append(events, event)
	}
	return events
}

func seedRSVPs(users []models.User, events []models.Event) {
	statuses := []string{"GOING", "GOING", "GOING", "INTERESTED", "CANT_GO"}

	for _, event := range events {
		// Creator always attends their own event.
		now := time.Now()
		db.DB.Create(&models.RSVP{
			ID:          uuid.New(),
			UserID:      event.CreatorID,
			EventID:     event.ID,
			Status:      "GOING",
			RespondedAt: &now,
			CreatedAt:   now,
		})

		attendeeCount := 3 + rand.Intn(6) // 3-8 additional attendees
		picked := map[uuid.UUID]bool{event.CreatorID: true}

		for i := 0; i < attendeeCount; i++ {
			user := users[rand.Intn(len(users))]
			if picked[user.ID] {
				continue
			}
			picked[user.ID] = true

			respondedAt := time.Now()
			db.DB.Create(&models.RSVP{
				ID:          uuid.New(),
				UserID:      user.ID,
				EventID:     event.ID,
				Status:      statuses[rand.Intn(len(statuses))],
				RespondedAt: &respondedAt,
				CreatedAt:   respondedAt,
			})
		}
	}
}

var sampleComments = []string{
	"Looking forward to this!",
	"Is parking available nearby?",
	"Attended last time, highly recommend.",
	"Can we bring a plus one?",
	"What time should we arrive?",
	"This is exactly what I needed this month.",
	"Anyone want to carpool?",
	"Count me in!",
}

func seedComments(users []models.User, events []models.Event) {
	for _, event := range events {
		commentCount := 1 + rand.Intn(3) // 1-3 comments
		for i := 0; i < commentCount; i++ {
			user := users[rand.Intn(len(users))]
			db.DB.Create(&models.Comment{
				ID:        uuid.New(),
				UserID:    user.ID,
				EventID:   event.ID,
				Content:   sampleComments[rand.Intn(len(sampleComments))],
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			})
		}
	}
}

func seedNotifications(users []models.User, events []models.Event) {
	for i := 0; i < 20; i++ {
		user := users[rand.Intn(len(users))]
		event := events[rand.Intn(len(events))]

		notifTypes := []struct {
			kind    string
			message string
		}{
			{"comment", fmt.Sprintf("Someone commented on \"%s\"", event.Title)},
			{"reminder", fmt.Sprintf("\"%s\" is coming up soon", event.Title)},
			{"rsvp", fmt.Sprintf("Someone new is going to \"%s\"", event.Title)},
		}
		pick := notifTypes[rand.Intn(len(notifTypes))]
		eventID := event.ID

		db.DB.Create(&models.Notification{
			ID:        uuid.New(),
			UserID:    user.ID,
			EventID:   &eventID,
			Type:      pick.kind,
			Message:   pick.message,
			IsRead:    rand.Intn(2) == 0,
			CreatedAt: time.Now(),
		})
	}
}
