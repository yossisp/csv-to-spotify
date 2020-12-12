package client

import "net/http"

// SpotifyProvider holds auth state
type SpotifyProvider struct {
	client           *http.Client
	accessToken      string
	refreshToken     string
	maxRetries       int
	actualRetries    int
	userID           string
	LookedUpTracks   []SearchResult
	trackIsFoundChan chan bool
	playlistID       string
}

// SearchResult contains track metadata
type SearchResult struct {
	Tracks struct {
		Items []trackMetaData `json:"items"`
		Total int             `json:"total"`
	} `json:"tracks"`
	IsFound bool
}

// TracksLookupProgress struct
type TracksLookupProgress struct {
	IsFound chan bool
	Quit    chan bool
}

type accessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int    `json:"expires_in"`
}

type createdPlaylist struct {
	ID string `json:"id"`
}

type userPlaylists struct {
	Items []struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"items"`
}

type trackMetaData struct {
	ID   string `json:"id"`
	URI  string `json:"uri"`
	Name string `json:"name"`
}
