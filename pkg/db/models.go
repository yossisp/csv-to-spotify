package db

import "time"

// SpotifyUser contains spotify user data
type SpotifyUser struct {
	RefreshToken string    `json:"refreshToken" bson:"refreshToken"`
	AccessToken  string    `json:"accessToken" bson:"accessToken"`
	UserID       string    `json:"userId" bson:"userId"`
	UpdatedAt    time.Time `bson:"updatedAt"`
}
