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
	// EmailVerifiedAt is nil until the user confirms their email address.
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
}

type Cafe struct {
	ID                int        `json:"id"`
	Name              string     `json:"name"`
	Address           string     `json:"address"`
	Description       string     `json:"description"`
	AvgRating         float64    `json:"avg_rating"`
	ReviewCount       int        `json:"review_count"`
	ReviewerCount     int        `json:"reviewer_count"`
	Closed            bool       `json:"closed"`
	ClosedAt          *time.Time `json:"closed_at"`
	BackgroundImage   string     `json:"background_image_url"`
	BackgroundImageID *int       `json:"background_image_id"`
	// BackgroundSource is "review_image", "uploaded", or "" when there is none.
	BackgroundSource string    `json:"background_source"`
	CreatedAt        time.Time `json:"created_at"`
	Tags             []Tag     `json:"tags"`
}

type List struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	IsPublic       bool      `json:"is_public"`
	CafeCount      int       `json:"cafe_count"`
	ContainsCafe   bool      `json:"contains_cafe"`
	ContainsUser   bool      `json:"contains_user"`
	ContainsReview bool      `json:"contains_review"`
	CreatedAt      time.Time `json:"created_at"`
}

type ListDetail struct {
	List
	Cafes   []Cafe        `json:"cafes"`
	Users   []UserSummary `json:"users"`
	Reviews []Review      `json:"reviews"`
}

type UserSummary struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
}

type Tag struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Scope    string `json:"scope"`
}

type Review struct {
	ID         int           `json:"id"`
	CafeID     int           `json:"cafe_id"`
	UserID     int           `json:"user_id"`
	AuthorName string        `json:"author_name"`
	Rating     int           `json:"rating"`
	Body       string        `json:"body"`
	VisitedOn  string        `json:"visited_on"`
	CafeName   string        `json:"cafe_name,omitempty"`
	Images     []ReviewImage `json:"images"`
	Tags       []Tag         `json:"tags"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

type ReviewImage struct {
	ID        int       `json:"id"`
	ReviewID  int       `json:"review_id"`
	URL       string    `json:"url"`
	Tags      []Tag     `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}
