package client

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/yossisp/csv-to-spotify/pkg/csv"
)

func TestClient(t *testing.T) {
	provider := NewSpotifyProvider()
	provider.refreshToken = conf.TestRefreshToken
	const testRequestPositive = "TestRequestNoAccessTokenPositive"
	t.Run(testRequestPositive, func(t *testing.T) {
		const funcName = testRequestPositive
		track := csv.TrackInput{
			Artist: "The Beatles",
			Track:  "Yesterday",
		}
		lookupURL := fmt.Sprintf("%s%s", apiBaseURL, lookupTrackRoute)
		parsedURL, err := url.Parse(lookupURL)
		query, _ := url.ParseQuery(parsedURL.RawQuery)
		query.Add("q", fmt.Sprintf("artist:%s track:%s", track.Artist, track.Track))
		query.Add("type", "track")
		query.Add("limit", "1")
		query.Add("market", "IL")
		parsedURL.RawQuery = query.Encode()
		req, err := http.NewRequest(http.MethodGet, parsedURL.String(), nil)
		if err != nil {
			t.Errorf("%s: NewRequest: %v", funcName, err)
		}
		response, err := provider.request(req)
		if err != nil {
			t.Errorf("%s: request: %v", funcName, err)
		}
		defer response.Body.Close()
	})

	const testLookupTrackPositive = "TestLookupTrackPositive"
	t.Run(testLookupTrackPositive, func(t *testing.T) {
		inputTrack := csv.TrackInput{
			Artist: "The Beatles",
			Track:  "Yesterday",
		}
		track := provider.lookupTrack(inputTrack)
		if track == nil {
			t.Errorf("%s: couldn't find track", testLookupTrackPositive)
		}
	})
}
