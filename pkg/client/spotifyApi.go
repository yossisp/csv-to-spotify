package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/yossisp/csv-to-spotify/pkg/csv"
)

const (
	spotifyAccessTokenRoute    = "https://accounts.spotify.com/api/token"
	apiBaseURL                 = "https://api.spotify.com/v1"
	lookupTrackRoute           = "/search"
	playlistRoute              = "/users/{user_id}/playlists"
	createdPlaylistDescription = "Created by csv-to-spotify"
	addItemsToPlaylistRoute    = "/playlists/{playlist_id}/tracks"
)

/*
https://developer.spotify.com/documentation/web-api/reference/search/search/
TODO: set user country via `market` param dynamically
*/
func (provider *SpotifyProvider) lookupTrack(track csv.TrackInput) *SearchResult {
	const funcName = "lookupTrack"
	market := conf.Market
	lookupURL := fmt.Sprintf("%s%s", apiBaseURL, lookupTrackRoute)
	parsedURL, err := url.Parse(lookupURL)
	query, _ := url.ParseQuery(parsedURL.RawQuery)
	query.Add("q", fmt.Sprintf("artist:%s track:%s", track.Artist, track.Track))
	query.Add("type", "track")
	query.Add("limit", "1")
	query.Add("market", market)
	parsedURL.RawQuery = query.Encode()
	log.Println("making request to: ", parsedURL)
	req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		logger("%s: NewRequest: %v", funcName, err)
		return nil
	}

	response, err := provider.request(req)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	payload := SearchResult{}
	err = json.NewDecoder(response.Body).Decode(&payload)
	if err != nil {
		logger("%s: json.NewDecoder: %v", funcName, err)
		return nil
	}
	payload.IsFound = payload.Tracks.Total > 0
	return &payload
}

/*
	uses Client Credentials Flow (https://developer.spotify.com/documentation/general/guides/authorization-guide/)
*/
func (provider *SpotifyProvider) setAccessTokenClientFlow() {
	const funcName = "setAccessTokenClientFlow"
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, spotifyAccessTokenRoute, strings.NewReader(data.Encode()))
	if err != nil {
		logger("%s: NewRequest: %v", funcName, err)
		return
	}
	req.Header.Set("Authorization", "Basic "+conf.SpotifySecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := provider.client.Do(req)
	if err != nil {
		logger("%s: response: %v", funcName, err)
		return
	}
	defer response.Body.Close()
	result := accessToken{}
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		logger("%s: json.NewDecoder: %v", funcName, err)
		return
	}
	provider.accessToken = result.Token
}

/*
	uses Authorization Code Flow (https://developer.spotify.com/documentation/general/guides/authorization-guide/)
*/
func (provider *SpotifyProvider) setAccessToken() error {
	const funcName = "setAccessToken"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", provider.refreshToken)

	req, err := http.NewRequest(http.MethodPost, spotifyAccessTokenRoute, strings.NewReader(data.Encode()))
	if err != nil {
		logger("%s: NewRequest: %v", funcName, err)
		return err
	}
	req.Header.Set("Authorization", "Basic "+conf.SpotifySecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := provider.client.Do(req)
	if err != nil {
		logger("%s: response: %v", funcName, err)
		return err
	}
	defer response.Body.Close()
	result := accessToken{}
	json.NewDecoder(response.Body).Decode(&result)
	provider.accessToken = result.Token
	// log.Println("setAccessToken: ", provider.accessToken)
	return nil
}

func (provider *SpotifyProvider) sendJSONPayload(payload map[string]interface{}, apiRoute string) (response *http.Response, err error) {
	const funcName = "sendJSONPayload"
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logger("%s: json.Marshal: %v", funcName, err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, apiRoute, bytes.NewBuffer(payloadJSON))
	if err != nil {
		logger("%s: NewRequest: %v", funcName, err)
		return
	}
	response, err = provider.request(req)
	return
}

// getUserPlaylists get the list of user playlists
func (provider *SpotifyProvider) getUserPlaylists() *userPlaylists {
	const funcName = "getUserPlaylists"
	path := strings.Replace(playlistRoute, "{user_id}", provider.userID, 1)
	route := fmt.Sprintf("%s%s", apiBaseURL, path)
	req, err := http.NewRequest(http.MethodGet, route, nil)
	if err != nil {
		logger("%s: NewRequest: %v", funcName, err)
		return nil
	}
	response, err := provider.request(req)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	var payload userPlaylists
	err = json.NewDecoder(response.Body).Decode(&payload)
	if err != nil {
		logger("%s: json.NewDecoder: %v", funcName, err)
		return nil
	}
	return &payload
}

// CreatePlaylist creates new playlist
func (provider *SpotifyProvider) CreatePlaylist(playlistName string) (hasCreated bool) {
	const funcName = "CreatePlaylist"
	hasCreated = false
	logger("%s: desired playlist name: %s", funcName, playlistName)
	userPlaylists := provider.getUserPlaylists()
	if userPlaylists != nil {
		for _, playlist := range userPlaylists.Items {
			if playlist.Name == playlistName {
				logger("%s: playlist with name: %s already exists", funcName, playlistName)
				provider.playlistID = playlist.ID
				return
			}
		}
	}

	path := strings.Replace(playlistRoute, "{user_id}", provider.userID, 1)
	route := fmt.Sprintf("%s%s", apiBaseURL, path)
	body := map[string]interface{}{
		"name":        playlistName,
		"public":      false,
		"description": createdPlaylistDescription,
	}
	response, err := provider.sendJSONPayload(body, route)
	if err != nil {
		return
	}
	defer response.Body.Close()
	result := createdPlaylist{}
	json.NewDecoder(response.Body).Decode(&result)
	log.Println("created playlist id: ", result)
	provider.playlistID = result.ID
	hasCreated = true
	return
}

// AddItemsToPlaylist adds tracks to a playlist
// TODO: when there are more than 10.000 items in the playlist, returns error 403 Forbidden.
func (provider *SpotifyProvider) AddItemsToPlaylist() (hasAdded bool) {
	const funcName = "AddItemsToPlaylist"
	tracks := provider.LookedUpTracks
	hasAdded = false
	path := strings.Replace(addItemsToPlaylistRoute, "{playlist_id}", provider.playlistID, 1)
	route := fmt.Sprintf("%s%s", apiBaseURL, path)

	spotifyURIs := make([]string, len(tracks))
	for i, track := range tracks {
		spotifyURIs[i] = track.Tracks.Items[0].URI
	}
	body := map[string]interface{}{
		"uris": spotifyURIs,
	}
	response, err := provider.sendJSONPayload(body, route)
	if err != nil {
		return
	}
	defer response.Body.Close()
	logger("%s: added %d items to playlist id %s", funcName, len(tracks), provider.playlistID)
	hasAdded = true
	return
}
