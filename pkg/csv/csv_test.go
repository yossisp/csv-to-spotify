package csv

import (
	"testing"
)

func TestGetInputTracksPositive(t *testing.T) {
	csvFile := `Name,Artist,Composer,Album,Grouping,Work,Movement Number,Movement Count,Movement Name,Genre,Size,Time,Disc Number,Disc Count,Track Number,Track Count,Year,Date Modified,Date Added,Bit Rate,Sample Rate,Volume Adjustment,Kind,Equalizer,Comments,Plays,Last Played,Skips,Last Skipped,My Rating,Location
	Darling Pretty,Mark Knopfler,Mark Knopfler,Golden Heart,,,,,,Folk,10719229,267,1,,1,,1996,"01/05/2013, 22:56","21/01/2016, 21:57",320,44100,,Internet audio stream,,,22,"16/05/2020, 9:27",11,"18/01/2020, 14:03",,
	What It Is,Mark Knopfler,,Sailing to Philadelphia,,,,,,,11836732,295,1,,1,,,"03/10/2012, 3:59","21/01/2016, 21:57",320,44100,,Internet audio stream,,,19,"16/05/2020, 9:32",4,"15/05/2020, 10:04",,
	Sailing to Philadelphia,Mark Knopfler,,Sailing to Philadelphia,,,,,,,13152259,328,1,,2,,,"03/10/2012, 3:58","21/01/2016, 21:57",256,44100,,Internet audio stream,,,16,"27/04/2020, 14:41",1,"27/01/2017, 1:27",,`
	tracks, err := GetInputTracks(csvFile)
	if err != nil {
		t.Errorf("TestGetInputTracksPositive: %v", err)
	}

	if len(tracks) != 3 {
		t.Errorf("TestGetInputTracksPositive: len(tracks) != 3")
	}
}
