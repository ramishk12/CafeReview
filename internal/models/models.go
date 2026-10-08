package models

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"display_name"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Cafe struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Address     string    `json:"address"`
	Description string    `json:"description"`
	AvgRating   float64   `json:"avg_rating"`
	ReviewCount int       `json:"review_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Review struct {
	ID         int           `json:"id"`
	CafeID     int           `json:"cafe_id"`
	UserID     int           `json:"user_id"`
	AuthorName string        `json:"author_name"`
	Rating     int           `json:"rating"`
	Body       string        `json:"body"`
	Images     []ReviewImage `json:"images"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type ReviewImage struct {
	ID        int       `json:"id"`
	ReviewID  int       `json:"review_id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
