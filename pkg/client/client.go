/*
Package client communicates with Spotify
*/
package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/yossisp/csv-to-spotify/pkg/db"
	"github.com/yossisp/csv-to-spotify/pkg/utils"

	"github.com/yossisp/csv-to-spotify/pkg/config"
	"github.com/yossisp/csv-to-spotify/pkg/csv"
)

var (
	conf   config.Config = config.NewConfig()
	logger               = utils.NewLogger("client")
)

// NewSpotifyProvider provides spotify state
func NewSpotifyProvider() *SpotifyProvider {
	return &SpotifyProvider{
		maxRetries:       7,
		actualRetries:    0,
		trackIsFoundChan: make(chan bool),
		client: &http.Client{
			Timeout: time.Duration(time.Duration(conf.ClientTimeout) * time.Second),
		},
	}
}

func (provider *SpotifyProvider) lookupTracksBatch(batch []csv.TrackInput, isLastBatch bool, tracksProgress TracksLookupProgress) {
	searchChan := make(chan SearchResult)
	for _, inputTrack := range batch {
		inputTrack := inputTrack
		go func() {
			track := provider.lookupTrack(inputTrack)
			if track != nil {
				searchChan <- *track
			}
		}()
	}

	for range batch {
		track := <-searchChan
		if track.IsFound {
			log.Println("track found")
			log.Println(track)
			provider.LookedUpTracks = append(provider.LookedUpTracks, track)
		}
		tracksProgress.IsFound <- track.IsFound
	}
	if isLastBatch {
		tracksProgress.Quit <- true
	}
}

// GetSearchResults gets search results
func (provider *SpotifyProvider) GetSearchResults(tracksProgress TracksLookupProgress, inputTracks []csv.TrackInput) {
	var (
		lowerBound         int
		upperBound         int
		batch              []csv.TrackInput
		tracksSearchedNum  int = 0
		resultsNumPerBatch int = 3
	)
	trackLookupInterval, err := strconv.Atoi(conf.TrackLookupInterval)
	if err != nil {
		tracksProgress.Quit <- false
		return
	}
	playlistTracks := inputTracks[0:len(inputTracks)]
	batchesNum := int(math.Ceil(float64(len(playlistTracks)) / float64(resultsNumPerBatch)))
	log.Println("len(playlistTracks)", len(playlistTracks), "batchesNum", batchesNum)
	// access token may have expired so it should be set once
	// here instead of re-setting it `len(inputTracks)` times
	err = provider.setAccessToken()
	if err != nil {
		tracksProgress.Quit <- false
		return
	}

	for i := 0; i < batchesNum; i++ {
		lowerBound = i * resultsNumPerBatch
		upperBound = lowerBound + resultsNumPerBatch
		if upperBound > len(playlistTracks) {
			upperBound = lowerBound + (len(playlistTracks) % resultsNumPerBatch)
		}

		batch = playlistTracks[lowerBound:upperBound]
		log.Println("batch ", batch)
		log.Println("lowerBound ", lowerBound, "upperBound ", upperBound)
		tracksSearchedNum += len(batch)
		isLastBatch := i == batchesNum-1
		go provider.lookupTracksBatch(batch, isLastBatch, tracksProgress)
		utils.SleepFor(trackLookupInterval)
	}
}

func (track SearchResult) String() string {
	const funcName = "client.String"
	res, err := json.MarshalIndent(track, "", "    ")
	if err != nil {
		logger("%s json.MarshalIndent: %v", funcName, err)
	}
	return "SearchResult: " + string(res)
}

func (provider *SpotifyProvider) request(req *http.Request) (response *http.Response, err error) {
	const funcName = "request"
	nextRequestDelaySec := 5
	url := req.URL.String()

	// in case the first req fails
	// body needs to be cloned for the next request
	// https://github.com/golang/go/issues/36095
	clonedReq := req.Clone(req.Context())
	if req.Body != nil {
		clonedReq.Body, err = req.GetBody()
		if err != nil {
			logger("%s: url: %s clonedReq: %v", funcName, url, err)
			return nil, err
		}
	}
	defer func() {
		if clonedReq.Body != nil {
			clonedReq.Body.Close()
		}
	}()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.accessToken))
	response, err = provider.client.Do(req)
	if err != nil {
		logger("request: response: %v", err)
		return nil, err
	}

	if provider.maxRetries == provider.actualRetries {
		err = fmt.Errorf("%s: too many retries for url: %s", funcName, url)
		logger("%s: %v", funcName, err)
		response.Body.Close()
		return nil, err
	}
	logger("%s: url: %s statusCode: %d", funcName, url, response.StatusCode)
	if response.StatusCode >= http.StatusBadRequest && response.StatusCode < http.StatusInternalServerError && provider.maxRetries > provider.actualRetries {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logger("%s: url: %s ioutil.ReadAll: %v", funcName, url, err)
		} else {
			logger("%s: url: %s response body: %s", funcName, url, string(bodyBytes))
		}
		response.Body.Close()
		provider.actualRetries++
		utils.SleepFor(nextRequestDelaySec)
		err = provider.setAccessToken()
		if err != nil {
			return nil, err
		}
		response, err = provider.request(clonedReq)
	}

	return response, err
}

// SetUserData sets user token
func (provider *SpotifyProvider) SetUserData(user *db.SpotifyUser) {
	provider.refreshToken = user.RefreshToken
	provider.accessToken = user.AccessToken
	provider.userID = user.UserID
	// log.Println("userID", provider.userID, "accessToken", provider.accessToken)
}
