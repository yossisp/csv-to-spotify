package csv

import (
	"encoding/csv"
	"strings"

	"github.com/yossisp/csv-to-spotify/pkg/utils"
)

var (
	logger = utils.NewLogger("csv")
)

// TrackInput type
type TrackInput struct {
	Artist string
	Track  string
}

type trackInputChannelData struct {
	track TrackInput
	index int
}

// GetInputTracks converts CSV records to an array of TrackInput
func GetInputTracks(csvFile string) (tracks []TrackInput, err error) {
	const funcName = "GetInputTracks"
	csvReader := csv.NewReader(strings.NewReader(csvFile))
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		logger("%s: csvReader.ReadAll: %v", funcName, err)
		return nil, err
	}

	// disregard header
	if len(records) > 1 {
		records = records[1:]
		tracks = make([]TrackInput, len(records))
		trackChannel := make(chan trackInputChannelData)
		for index, record := range records {
			index, record := index, record
			go func() {
				trackChannel <- trackInputChannelData{track: TrackInput{Track: record[0], Artist: record[1]}, index: index}
			}()
		}

		for range records {
			record := <-trackChannel
			tracks[record.index] = record.track
		}
	}
	return
}
