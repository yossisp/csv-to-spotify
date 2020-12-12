package runner

import (
	"github.com/yossisp/csv-to-spotify/pkg/utils"

	"github.com/yossisp/csv-to-spotify/pkg/csv"
	"github.com/yossisp/csv-to-spotify/pkg/kafkahelper"

	"github.com/yossisp/csv-to-spotify/pkg/client"
	"github.com/yossisp/csv-to-spotify/pkg/db"
)

var (
	producer = kafkahelper.NewProducer()
	logger   = utils.NewLogger("runner")
)

func init() {
	go producer.LogDeliveredMessages()
}

// Runner starts playlist copy job
type Runner struct {
	tracksAdded    int
	tracksNotAdded int
	csvFile        string
	fileName       string
	user           db.SpotifyUser
}

// CSVPayload contains csv file data
type CSVPayload struct {
	UserID   *string `json:"userId"`
	CSVFile  *string `json:"csvFile"`
	FileName *string `json:"uploadFileName"`
}

// NewRunner returns a runner
func NewRunner(input CSVPayload, user *db.SpotifyUser) *Runner {
	return &Runner{
		csvFile:  *input.CSVFile,
		fileName: *input.FileName,
		user:     *user,
	}
}

// Run starts playlist copy job
func (runner *Runner) Run() {
	const funcName = "Run"
	var (
		isTrackFound bool
		isSuccess    bool
	)
	tracksProgress := client.TracksLookupProgress{
		IsFound: make(chan bool),
		Quit:    make(chan bool),
	}
	tracks, err := csv.GetInputTracks(runner.csvFile)
	if err != nil {
		producer.ProduceMessage(runner.user.UserID, kafkahelper.CSVFileError)
		return
	}

	spotifyProvider := client.NewSpotifyProvider()
	spotifyProvider.SetUserData(&runner.user)
	spotifyProvider.CreatePlaylist(runner.fileName)
	go spotifyProvider.GetSearchResults(tracksProgress, tracks)

	for {
		select {
		case isTrackFound = <-tracksProgress.IsFound:
			if isTrackFound {
				runner.tracksAdded++
			} else {
				runner.tracksNotAdded++
			}
			producer.ProduceMessage(runner.user.UserID, kafkahelper.TrackProgress, runner.tracksAdded, runner.tracksNotAdded)
		case isSuccess = <-tracksProgress.Quit:
			if isSuccess {
				spotifyProvider.AddItemsToPlaylist()
				producer.ProduceMessage(runner.user.UserID, kafkahelper.JobFinished)
			}
			close(tracksProgress.Quit)
			close(tracksProgress.IsFound)
			return
		}
	}
}
