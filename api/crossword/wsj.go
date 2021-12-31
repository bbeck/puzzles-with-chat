package crossword

import (
	"fmt"
	"time"
)

// LoadFromWallStreetJournal loads a crossword puzzle from the Wall Street
// Journal for a particular date.
//
// This method downloads a .puz file and loads it into a Puzzle object.  We
// do this in particular on the server side instead of within the client because
// the herbach.dnsalias.com site unfortunately is only HTTP and we can't load
// resources from a non-HTTPS site in the browser.
//
// If the puzzle cannot be loaded or parsed then an error is returned.
func LoadFromWallStreetJournal(date string) (*Puzzle, error) {
	published, err := time.Parse("2006-01-02", date)
	if err != nil {
		err = fmt.Errorf("unable to parse date %s: %+v", date, err)
		return nil, err
	}

	// Download the .puz file from the herbach.dnsalias.com site.
	url := fmt.Sprintf("http://herbach.dnsalias.com/wsj/wsj%02d%02d%02d.puz", published.Year()-2000, published.Month(), published.Day())
	puzzle, err := LoadFromPuzFileURL(url)
	if err != nil {
		return nil, err
	}

	puzzle.Description = fmt.Sprintf("Wall Street Journal puzzle from %s", published.Format("2006-01-02"))

	// Normally .puz files don't have puzzle dates recorded in them, but we
	// happen to know the date for this puzzle, so fill it in.
	puzzle.PublishedDate = published
	puzzle.Publisher = "The Wall Street Journal"

	return puzzle, nil
}

// LoadAvailableWSJDates calculates the set of available dates for crossword
// puzzles from The Wall Street Journal.
func LoadAvailableWSJDates() []time.Time {
	now := time.Now().UTC()

	var dates []time.Time
	for _, s := range AvailableWSJDates {
		date, err := time.Parse("2006-01-02", s)
		if err != nil || date.After(now) {
			continue
		}

		dates = append(dates, date)
	}

	return dates
}

// The following script was used to obtain the WSJ dates.  Puzzles between 2009
// and 2013 have an index page on fleetingimage.com, but are hosted on a
// different site and the .puz files are no longer there.
//
// for yy in {13..21}; do
//   curl --silent "https://www.fleetingimage.com/wij/xyzzy/${yy}-wsj.html" |
//   python3 -c '
// import re, sys
// for date in re.findall("""href=["]wsj([0-9]+).puz["]""", sys.stdin.read()):
//   print("20" + date)
//   '
// done |
// sed -E 's/(....)(..)(..)/"\1-\2-\3",/g' |
// sort -u
var AvailableWSJDates = []string{
	"2013-01-04",
	"2013-01-11",
	"2013-01-18",
	"2013-01-25",
	"2013-02-01",
	"2013-02-08",
	"2013-02-15",
	"2013-02-22",
	"2013-03-01",
	"2013-03-08",
	"2013-03-15",
	"2013-03-22",
	"2013-03-29",
	"2013-04-05",
	"2013-04-12",
	"2013-04-19",
	"2013-04-26",
	"2013-05-03",
	"2013-05-10",
	"2013-05-17",
	"2013-05-24",
	"2013-05-31",
	"2013-06-07",
	"2013-06-14",
	"2013-06-21",
	"2013-06-28",
	"2013-07-05",
	"2013-07-12",
	"2013-07-19",
	"2013-07-26",
	"2013-08-02",
	"2013-08-09",
	"2013-08-16",
	"2013-08-23",
	"2013-08-30",
	"2013-09-06",
	"2013-09-13",
	"2013-09-20",
	"2013-09-27",
	"2013-10-04",
	"2013-10-11",
	"2013-10-18",
	"2013-10-25",
	"2013-11-01",
	"2013-11-08",
	"2013-11-15",
	"2013-11-22",
	"2013-11-29",
	"2013-12-06",
	"2013-12-13",
	"2013-12-20",
	"2013-12-27",
	"2014-01-03",
	"2014-01-10",
	"2014-01-17",
	"2014-01-24",
	"2014-01-31",
	"2014-02-07",
	"2014-02-14",
	"2014-02-21",
	"2014-02-28",
	"2014-03-07",
	"2014-03-14",
	"2014-03-21",
	"2014-03-28",
	"2014-04-04",
	"2014-04-11",
	"2014-04-18",
	"2014-04-25",
	"2014-05-02",
	"2014-05-09",
	"2014-05-16",
	"2014-05-23",
	"2014-05-30",
	"2014-06-06",
	"2014-06-13",
	"2014-06-20",
	"2014-06-27",
	"2014-07-04",
	"2014-07-11",
	"2014-07-18",
	"2014-07-25",
	"2014-08-01",
	"2014-08-08",
	"2014-08-15",
	"2014-08-22",
	"2014-08-29",
	"2014-09-05",
	"2014-09-12",
	"2014-09-19",
	"2014-09-26",
	"2014-10-03",
	"2014-10-10",
	"2014-10-17",
	"2014-10-24",
	"2014-10-31",
	"2014-11-07",
	"2014-11-14",
	"2014-11-21",
	"2014-11-28",
	"2014-12-05",
	"2014-12-12",
	"2014-12-19",
	"2014-12-26",
	"2015-01-02",
	"2015-01-09",
	"2015-01-16",
	"2015-01-23",
	"2015-01-30",
	"2015-02-06",
	"2015-02-13",
	"2015-02-20",
	"2015-02-27",
	"2015-03-06",
	"2015-03-13",
	"2015-03-20",
	"2015-03-27",
	"2015-04-03",
	"2015-04-10",
	"2015-04-17",
	"2015-04-24",
	"2015-05-01",
	"2015-05-08",
	"2015-05-15",
	"2015-05-22",
	"2015-05-29",
	"2015-06-05",
	"2015-06-12",
	"2015-06-19",
	"2015-06-26",
	"2015-07-03",
	"2015-07-10",
	"2015-07-17",
	"2015-07-24",
	"2015-07-31",
	"2015-08-07",
	"2015-08-14",
	"2015-08-21",
	"2015-08-28",
	"2015-09-04",
	"2015-09-11",
	"2015-09-19",
	"2015-09-26",
	"2015-10-03",
	"2015-10-10",
	"2015-10-17",
	"2015-10-24",
	"2015-10-31",
	"2015-11-07",
	"2015-11-14",
	"2015-11-21",
	"2015-11-28",
	"2015-12-05",
	"2015-12-12",
	"2015-12-19",
	"2015-12-26",
	"2016-01-02",
	"2016-01-04",
	"2016-01-05",
	"2016-01-06",
	"2016-01-07",
	"2016-01-08",
	"2016-01-09",
	"2016-01-11",
	"2016-01-12",
	"2016-01-13",
	"2016-01-14",
	"2016-01-15",
	"2016-01-16",
	"2016-01-19",
	"2016-01-20",
	"2016-01-21",
	"2016-01-22",
	"2016-01-23",
	"2016-01-25",
	"2016-01-26",
	"2016-01-27",
	"2016-01-28",
	"2016-01-29",
	"2016-01-30",
	"2016-02-01",
	"2016-02-02",
	"2016-02-03",
	"2016-02-04",
	"2016-02-05",
	"2016-02-06",
	"2016-02-08",
	"2016-02-09",
	"2016-02-10",
	"2016-02-11",
	"2016-02-12",
	"2016-02-13",
	"2016-02-16",
	"2016-02-17",
	"2016-02-18",
	"2016-02-19",
	"2016-02-20",
	"2016-02-22",
	"2016-02-23",
	"2016-02-24",
	"2016-02-25",
	"2016-02-26",
	"2016-02-27",
	"2016-02-29",
	"2016-03-01",
	"2016-03-02",
	"2016-03-03",
	"2016-03-04",
	"2016-03-05",
	"2016-03-07",
	"2016-03-08",
	"2016-03-09",
	"2016-03-10",
	"2016-03-11",
	"2016-03-12",
	"2016-03-14",
	"2016-03-15",
	"2016-03-16",
	"2016-03-17",
	"2016-03-18",
	"2016-03-19",
	"2016-03-21",
	"2016-03-22",
	"2016-03-23",
	"2016-03-24",
	"2016-03-25",
	"2016-03-26",
	"2016-03-28",
	"2016-03-29",
	"2016-03-30",
	"2016-03-31",
	"2016-04-01",
	"2016-04-02",
	"2016-04-04",
	"2016-04-05",
	"2016-04-06",
	"2016-04-07",
	"2016-04-08",
	"2016-04-09",
	"2016-04-11",
	"2016-04-12",
	"2016-04-13",
	"2016-04-14",
	"2016-04-15",
	"2016-04-16",
	"2016-04-18",
	"2016-04-19",
	"2016-04-20",
	"2016-04-21",
	"2016-04-22",
	"2016-04-23",
	"2016-04-25",
	"2016-04-26",
	"2016-04-27",
	"2016-04-28",
	"2016-04-29",
	"2016-04-30",
	"2016-05-02",
	"2016-05-03",
	"2016-05-04",
	"2016-05-05",
	"2016-05-06",
	"2016-05-07",
	"2016-05-09",
	"2016-05-10",
	"2016-05-11",
	"2016-05-12",
	"2016-05-13",
	"2016-05-14",
	"2016-05-16",
	"2016-05-17",
	"2016-05-18",
	"2016-05-19",
	"2016-05-20",
	"2016-05-21",
	"2016-05-23",
	"2016-05-24",
	"2016-05-25",
	"2016-05-26",
	"2016-05-27",
	"2016-05-28",
	"2016-05-31",
	"2016-06-01",
	"2016-06-02",
	"2016-06-03",
	"2016-06-04",
	"2016-06-06",
	"2016-06-07",
	"2016-06-08",
	"2016-06-09",
	"2016-06-10",
	"2016-06-11",
	"2016-06-13",
	"2016-06-14",
	"2016-06-15",
	"2016-06-16",
	"2016-06-17",
	"2016-06-18",
	"2016-06-20",
	"2016-06-21",
	"2016-06-22",
	"2016-06-23",
	"2016-06-24",
	"2016-06-25",
	"2016-06-27",
	"2016-06-28",
	"2016-06-29",
	"2016-06-30",
	"2016-07-01",
	"2016-07-02",
	"2016-07-05",
	"2016-07-06",
	"2016-07-07",
	"2016-07-08",
	"2016-07-09",
	"2016-07-11",
	"2016-07-12",
	"2016-07-13",
	"2016-07-14",
	"2016-07-15",
	"2016-07-16",
	"2016-07-18",
	"2016-07-19",
	"2016-07-20",
	"2016-07-21",
	"2016-07-22",
	"2016-07-23",
	"2016-07-25",
	"2016-07-26",
	"2016-07-27",
	"2016-07-28",
	"2016-07-29",
	"2016-07-30",
	"2016-08-01",
	"2016-08-02",
	"2016-08-03",
	"2016-08-04",
	"2016-08-05",
	"2016-08-06",
	"2016-08-08",
	"2016-08-09",
	"2016-08-10",
	"2016-08-11",
	"2016-08-12",
	"2016-08-13",
	"2016-08-15",
	"2016-08-16",
	"2016-08-17",
	"2016-08-18",
	"2016-08-19",
	"2016-08-20",
	"2016-08-22",
	"2016-08-23",
	"2016-08-24",
	"2016-08-25",
	"2016-08-26",
	"2016-08-27",
	"2016-08-29",
	"2016-08-30",
	"2016-08-31",
	"2016-09-01",
	"2016-09-02",
	"2016-09-03",
	"2016-09-06",
	"2016-09-07",
	"2016-09-08",
	"2016-09-09",
	"2016-09-10",
	"2016-09-12",
	"2016-09-13",
	"2016-09-14",
	"2016-09-15",
	"2016-09-16",
	"2016-09-17",
	"2016-09-19",
	"2016-09-20",
	"2016-09-21",
	"2016-09-22",
	"2016-09-23",
	"2016-09-24",
	"2016-09-26",
	"2016-09-27",
	"2016-09-28",
	"2016-09-29",
	"2016-09-30",
	"2016-10-01",
	"2016-10-03",
	"2016-10-04",
	"2016-10-05",
	"2016-10-06",
	"2016-10-07",
	"2016-10-08",
	"2016-10-10",
	"2016-10-11",
	"2016-10-12",
	"2016-10-13",
	"2016-10-14",
	"2016-10-15",
	"2016-10-17",
	"2016-10-18",
	"2016-10-19",
	"2016-10-20",
	"2016-10-21",
	"2016-10-22",
	"2016-10-24",
	"2016-10-25",
	"2016-10-26",
	"2016-10-27",
	"2016-10-28",
	"2016-10-29",
	"2016-10-31",
	"2016-11-01",
	"2016-11-02",
	"2016-11-03",
	"2016-11-04",
	"2016-11-05",
	"2016-11-07",
	"2016-11-08",
	"2016-11-09",
	"2016-11-10",
	"2016-11-11",
	"2016-11-12",
	"2016-11-14",
	"2016-11-15",
	"2016-11-16",
	"2016-11-17",
	"2016-11-18",
	"2016-11-19",
	"2016-11-21",
	"2016-11-22",
	"2016-11-23",
	"2016-11-25",
	"2016-11-26",
	"2016-11-28",
	"2016-11-29",
	"2016-11-30",
	"2016-12-01",
	"2016-12-02",
	"2016-12-03",
	"2016-12-05",
	"2016-12-06",
	"2016-12-07",
	"2016-12-08",
	"2016-12-09",
	"2016-12-10",
	"2016-12-12",
	"2016-12-13",
	"2016-12-14",
	"2016-12-15",
	"2016-12-16",
	"2016-12-17",
	"2016-12-19",
	"2016-12-20",
	"2016-12-21",
	"2016-12-22",
	"2016-12-23",
	"2016-12-24",
	"2016-12-27",
	"2016-12-28",
	"2016-12-29",
	"2016-12-30",
	"2016-12-31",
	"2017-01-03",
	"2017-01-04",
	"2017-01-05",
	"2017-01-06",
	"2017-01-07",
	"2017-01-09",
	"2017-01-10",
	"2017-01-11",
	"2017-01-12",
	"2017-01-13",
	"2017-01-14",
	"2017-01-17",
	"2017-01-18",
	"2017-01-19",
	"2017-01-20",
	"2017-01-21",
	"2017-01-23",
	"2017-01-24",
	"2017-01-25",
	"2017-01-26",
	"2017-01-27",
	"2017-01-28",
	"2017-01-30",
	"2017-01-31",
	"2017-02-01",
	"2017-02-02",
	"2017-02-03",
	"2017-02-04",
	"2017-02-06",
	"2017-02-07",
	"2017-02-08",
	"2017-02-09",
	"2017-02-10",
	"2017-02-11",
	"2017-02-13",
	"2017-02-14",
	"2017-02-15",
	"2017-02-16",
	"2017-02-17",
	"2017-02-18",
	"2017-02-21",
	"2017-02-22",
	"2017-02-23",
	"2017-02-24",
	"2017-02-25",
	"2017-02-27",
	"2017-02-28",
	"2017-03-01",
	"2017-03-02",
	"2017-03-03",
	"2017-03-04",
	"2017-03-06",
	"2017-03-07",
	"2017-03-08",
	"2017-03-09",
	"2017-03-10",
	"2017-03-11",
	"2017-03-13",
	"2017-03-14",
	"2017-03-15",
	"2017-03-16",
	"2017-03-17",
	"2017-03-18",
	"2017-03-20",
	"2017-03-21",
	"2017-03-22",
	"2017-03-23",
	"2017-03-24",
	"2017-03-25",
	"2017-03-27",
	"2017-03-28",
	"2017-03-29",
	"2017-03-30",
	"2017-03-31",
	"2017-04-01",
	"2017-04-03",
	"2017-04-04",
	"2017-04-05",
	"2017-04-06",
	"2017-04-07",
	"2017-04-08",
	"2017-04-10",
	"2017-04-11",
	"2017-04-12",
	"2017-04-13",
	"2017-04-14",
	"2017-04-15",
	"2017-04-17",
	"2017-04-18",
	"2017-04-19",
	"2017-04-20",
	"2017-04-21",
	"2017-04-22",
	"2017-04-24",
	"2017-04-25",
	"2017-04-26",
	"2017-04-27",
	"2017-04-28",
	"2017-04-29",
	"2017-05-01",
	"2017-05-02",
	"2017-05-03",
	"2017-05-04",
	"2017-05-05",
	"2017-05-06",
	"2017-05-08",
	"2017-05-09",
	"2017-05-10",
	"2017-05-11",
	"2017-05-12",
	"2017-05-13",
	"2017-05-15",
	"2017-05-16",
	"2017-05-17",
	"2017-05-18",
	"2017-05-19",
	"2017-05-20",
	"2017-05-22",
	"2017-05-23",
	"2017-05-24",
	"2017-05-25",
	"2017-05-26",
	"2017-05-27",
	"2017-05-30",
	"2017-05-31",
	"2017-06-01",
	"2017-06-02",
	"2017-06-03",
	"2017-06-05",
	"2017-06-06",
	"2017-06-07",
	"2017-06-08",
	"2017-06-09",
	"2017-06-10",
	"2017-06-12",
	"2017-06-13",
	"2017-06-14",
	"2017-06-15",
	"2017-06-16",
	"2017-06-17",
	"2017-06-19",
	"2017-06-20",
	"2017-06-21",
	"2017-06-22",
	"2017-06-23",
	"2017-06-24",
	"2017-06-26",
	"2017-06-27",
	"2017-06-28",
	"2017-06-29",
	"2017-06-30",
	"2017-07-01",
	"2017-07-03",
	"2017-07-05",
	"2017-07-06",
	"2017-07-07",
	"2017-07-08",
	"2017-07-10",
	"2017-07-11",
	"2017-07-12",
	"2017-07-13",
	"2017-07-14",
	"2017-07-15",
	"2017-07-17",
	"2017-07-18",
	"2017-07-19",
	"2017-07-20",
	"2017-07-21",
	"2017-07-22",
	"2017-07-24",
	"2017-07-25",
	"2017-07-26",
	"2017-07-27",
	"2017-07-28",
	"2017-07-29",
	"2017-07-31",
	"2017-08-01",
	"2017-08-02",
	"2017-08-03",
	"2017-08-04",
	"2017-08-05",
	"2017-08-07",
	"2017-08-08",
	"2017-08-09",
	"2017-08-10",
	"2017-08-11",
	"2017-08-12",
	"2017-08-14",
	"2017-08-15",
	"2017-08-16",
	"2017-08-17",
	"2017-08-18",
	"2017-08-19",
	"2017-08-21",
	"2017-08-22",
	"2017-08-23",
	"2017-08-24",
	"2017-08-25",
	"2017-08-26",
	"2017-08-28",
	"2017-08-29",
	"2017-08-30",
	"2017-08-31",
	"2017-09-01",
	"2017-09-02",
	"2017-09-05",
	"2017-09-06",
	"2017-09-07",
	"2017-09-08",
	"2017-09-09",
	"2017-09-11",
	"2017-09-12",
	"2017-09-13",
	"2017-09-14",
	"2017-09-15",
	"2017-09-16",
	"2017-09-18",
	"2017-09-19",
	"2017-09-20",
	"2017-09-21",
	"2017-09-22",
	"2017-09-23",
	"2017-09-25",
	"2017-09-26",
	"2017-09-27",
	"2017-09-28",
	"2017-09-29",
	"2017-09-30",
	"2017-10-02",
	"2017-10-03",
	"2017-10-04",
	"2017-10-05",
	"2017-10-06",
	"2017-10-07",
	"2017-10-09",
	"2017-10-10",
	"2017-10-11",
	"2017-10-12",
	"2017-10-13",
	"2017-10-14",
	"2017-10-16",
	"2017-10-17",
	"2017-10-18",
	"2017-10-19",
	"2017-10-20",
	"2017-10-21",
	"2017-10-23",
	"2017-10-24",
	"2017-10-25",
	"2017-10-26",
	"2017-10-27",
	"2017-10-28",
	"2017-10-30",
	"2017-10-31",
	"2017-11-01",
	"2017-11-02",
	"2017-11-03",
	"2017-11-04",
	"2017-11-06",
	"2017-11-07",
	"2017-11-08",
	"2017-11-09",
	"2017-11-10",
	"2017-11-11",
	"2017-11-13",
	"2017-11-14",
	"2017-11-15",
	"2017-11-16",
	"2017-11-17",
	"2017-11-18",
	"2017-11-20",
	"2017-11-21",
	"2017-11-22",
	"2017-11-24",
	"2017-11-25",
	"2017-11-27",
	"2017-11-28",
	"2017-11-29",
	"2017-11-30",
	"2017-12-01",
	"2017-12-02",
	"2017-12-04",
	"2017-12-05",
	"2017-12-06",
	"2017-12-07",
	"2017-12-08",
	"2017-12-09",
	"2017-12-11",
	"2017-12-12",
	"2017-12-13",
	"2017-12-14",
	"2017-12-15",
	"2017-12-16",
	"2017-12-18",
	"2017-12-19",
	"2017-12-20",
	"2017-12-21",
	"2017-12-22",
	"2017-12-23",
	"2017-12-26",
	"2017-12-27",
	"2017-12-28",
	"2017-12-29",
	"2017-12-30",
	"2018-01-02",
	"2018-01-03",
	"2018-01-04",
	"2018-01-05",
	"2018-01-06",
	"2018-01-08",
	"2018-01-09",
	"2018-01-10",
	"2018-01-11",
	"2018-01-12",
	"2018-01-13",
	"2018-01-16",
	"2018-01-17",
	"2018-01-18",
	"2018-01-19",
	"2018-01-20",
	"2018-01-22",
	"2018-01-23",
	"2018-01-24",
	"2018-01-25",
	"2018-01-26",
	"2018-01-27",
	"2018-01-29",
	"2018-01-30",
	"2018-01-31",
	"2018-02-01",
	"2018-02-02",
	"2018-02-03",
	"2018-02-05",
	"2018-02-06",
	"2018-02-07",
	"2018-02-08",
	"2018-02-09",
	"2018-02-10",
	"2018-02-12",
	"2018-02-13",
	"2018-02-14",
	"2018-02-15",
	"2018-02-16",
	"2018-02-17",
	"2018-02-20",
	"2018-02-21",
	"2018-02-22",
	"2018-02-23",
	"2018-02-24",
	"2018-02-26",
	"2018-02-27",
	"2018-02-28",
	"2018-03-01",
	"2018-03-02",
	"2018-03-03",
	"2018-03-05",
	"2018-03-06",
	"2018-03-07",
	"2018-03-08",
	"2018-03-09",
	"2018-03-10",
	"2018-03-12",
	"2018-03-13",
	"2018-03-14",
	"2018-03-15",
	"2018-03-16",
	"2018-03-17",
	"2018-03-19",
	"2018-03-20",
	"2018-03-21",
	"2018-03-22",
	"2018-03-23",
	"2018-03-24",
	"2018-03-26",
	"2018-03-27",
	"2018-03-28",
	"2018-03-29",
	"2018-03-30",
	"2018-03-31",
	"2018-04-02",
	"2018-04-03",
	"2018-04-04",
	"2018-04-05",
	"2018-04-06",
	"2018-04-07",
	"2018-04-09",
	"2018-04-10",
	"2018-04-11",
	"2018-04-12",
	"2018-04-13",
	"2018-04-14",
	"2018-04-16",
	"2018-04-17",
	"2018-04-18",
	"2018-04-19",
	"2018-04-20",
	"2018-04-21",
	"2018-04-23",
	"2018-04-24",
	"2018-04-25",
	"2018-04-26",
	"2018-04-27",
	"2018-04-28",
	"2018-04-30",
	"2018-05-01",
	"2018-05-02",
	"2018-05-03",
	"2018-05-04",
	"2018-05-05",
	"2018-05-07",
	"2018-05-08",
	"2018-05-09",
	"2018-05-10",
	"2018-05-11",
	"2018-05-12",
	"2018-05-14",
	"2018-05-15",
	"2018-05-16",
	"2018-05-17",
	"2018-05-18",
	"2018-05-19",
	"2018-05-21",
	"2018-05-22",
	"2018-05-23",
	"2018-05-24",
	"2018-05-25",
	"2018-05-26",
	"2018-05-29",
	"2018-05-30",
	"2018-05-31",
	"2018-06-01",
	"2018-06-02",
	"2018-06-04",
	"2018-06-05",
	"2018-06-06",
	"2018-06-07",
	"2018-06-08",
	"2018-06-09",
	"2018-06-11",
	"2018-06-12",
	"2018-06-13",
	"2018-06-14",
	"2018-06-15",
	"2018-06-16",
	"2018-06-18",
	"2018-06-19",
	"2018-06-20",
	"2018-06-21",
	"2018-06-22",
	"2018-06-23",
	"2018-06-25",
	"2018-06-26",
	"2018-06-27",
	"2018-06-28",
	"2018-06-29",
	"2018-06-30",
	"2018-07-02",
	"2018-07-03",
	"2018-07-05",
	"2018-07-06",
	"2018-07-07",
	"2018-07-09",
	"2018-07-10",
	"2018-07-11",
	"2018-07-12",
	"2018-07-13",
	"2018-07-14",
	"2018-07-16",
	"2018-07-17",
	"2018-07-18",
	"2018-07-19",
	"2018-07-20",
	"2018-07-21",
	"2018-07-23",
	"2018-07-24",
	"2018-07-25",
	"2018-07-26",
	"2018-07-27",
	"2018-07-28",
	"2018-07-30",
	"2018-07-31",
	"2018-08-01",
	"2018-08-02",
	"2018-08-03",
	"2018-08-04",
	"2018-08-06",
	"2018-08-07",
	"2018-08-08",
	"2018-08-09",
	"2018-08-10",
	"2018-08-11",
	"2018-08-13",
	"2018-08-14",
	"2018-08-15",
	"2018-08-16",
	"2018-08-17",
	"2018-08-18",
	"2018-08-20",
	"2018-08-21",
	"2018-08-22",
	"2018-08-23",
	"2018-08-24",
	"2018-08-25",
	"2018-08-27",
	"2018-08-28",
	"2018-08-29",
	"2018-08-30",
	"2018-08-31",
	"2018-09-01",
	"2018-09-04",
	"2018-09-05",
	"2018-09-06",
	"2018-09-07",
	"2018-09-08",
	"2018-09-10",
	"2018-09-11",
	"2018-09-12",
	"2018-09-13",
	"2018-09-14",
	"2018-09-15",
	"2018-09-17",
	"2018-09-18",
	"2018-09-19",
	"2018-09-20",
	"2018-09-21",
	"2018-09-22",
	"2018-09-24",
	"2018-09-25",
	"2018-09-26",
	"2018-09-27",
	"2018-09-28",
	"2018-09-29",
	"2018-10-01",
	"2018-10-02",
	"2018-10-03",
	"2018-10-04",
	"2018-10-05",
	"2018-10-06",
	"2018-10-08",
	"2018-10-09",
	"2018-10-10",
	"2018-10-11",
	"2018-10-12",
	"2018-10-13",
	"2018-10-15",
	"2018-10-16",
	"2018-10-17",
	"2018-10-18",
	"2018-10-19",
	"2018-10-20",
	"2018-10-22",
	"2018-10-23",
	"2018-10-24",
	"2018-10-25",
	"2018-10-26",
	"2018-10-27",
	"2018-10-29",
	"2018-10-30",
	"2018-10-31",
	"2018-11-01",
	"2018-11-02",
	"2018-11-03",
	"2018-11-05",
	"2018-11-06",
	"2018-11-07",
	"2018-11-08",
	"2018-11-09",
	"2018-11-10",
	"2018-11-12",
	"2018-11-13",
	"2018-11-14",
	"2018-11-15",
	"2018-11-16",
	"2018-11-17",
	"2018-11-19",
	"2018-11-20",
	"2018-11-21",
	"2018-11-23",
	"2018-11-24",
	"2018-11-26",
	"2018-11-27",
	"2018-11-28",
	"2018-11-29",
	"2018-11-30",
	"2018-12-01",
	"2018-12-03",
	"2018-12-04",
	"2018-12-05",
	"2018-12-06",
	"2018-12-07",
	"2018-12-08",
	"2018-12-10",
	"2018-12-11",
	"2018-12-12",
	"2018-12-13",
	"2018-12-14",
	"2018-12-15",
	"2018-12-17",
	"2018-12-18",
	"2018-12-19",
	"2018-12-20",
	"2018-12-21",
	"2018-12-22",
	"2018-12-24",
	"2018-12-26",
	"2018-12-27",
	"2018-12-28",
	"2018-12-29",
	"2018-12-31",
	"2019-01-02",
	"2019-01-03",
	"2019-01-04",
	"2019-01-05",
	"2019-01-07",
	"2019-01-08",
	"2019-01-09",
	"2019-01-10",
	"2019-01-11",
	"2019-01-12",
	"2019-01-14",
	"2019-01-15",
	"2019-01-16",
	"2019-01-17",
	"2019-01-18",
	"2019-01-19",
	"2019-01-22",
	"2019-01-23",
	"2019-01-24",
	"2019-01-25",
	"2019-01-26",
	"2019-01-28",
	"2019-01-29",
	"2019-01-30",
	"2019-01-31",
	"2019-02-01",
	"2019-02-02",
	"2019-02-04",
	"2019-02-05",
	"2019-02-06",
	"2019-02-07",
	"2019-02-08",
	"2019-02-09",
	"2019-02-11",
	"2019-02-12",
	"2019-02-13",
	"2019-02-14",
	"2019-02-15",
	"2019-02-16",
	"2019-02-19",
	"2019-02-20",
	"2019-02-21",
	"2019-02-22",
	"2019-02-23",
	"2019-02-25",
	"2019-02-26",
	"2019-02-27",
	"2019-02-28",
	"2019-03-01",
	"2019-03-02",
	"2019-03-04",
	"2019-03-05",
	"2019-03-06",
	"2019-03-07",
	"2019-03-08",
	"2019-03-09",
	"2019-03-11",
	"2019-03-12",
	"2019-03-13",
	"2019-03-14",
	"2019-03-15",
	"2019-03-16",
	"2019-03-18",
	"2019-03-19",
	"2019-03-20",
	"2019-03-21",
	"2019-03-22",
	"2019-03-23",
	"2019-03-25",
	"2019-03-26",
	"2019-03-27",
	"2019-03-28",
	"2019-03-29",
	"2019-03-30",
	"2019-04-01",
	"2019-04-02",
	"2019-04-03",
	"2019-04-04",
	"2019-04-05",
	"2019-04-06",
	"2019-04-08",
	"2019-04-09",
	"2019-04-10",
	"2019-04-11",
	"2019-04-12",
	"2019-04-13",
	"2019-04-15",
	"2019-04-16",
	"2019-04-17",
	"2019-04-18",
	"2019-04-19",
	"2019-04-20",
	"2019-04-22",
	"2019-04-23",
	"2019-04-24",
	"2019-04-25",
	"2019-04-26",
	"2019-04-27",
	"2019-04-29",
	"2019-04-30",
	"2019-05-01",
	"2019-05-02",
	"2019-05-03",
	"2019-05-04",
	"2019-05-06",
	"2019-05-07",
	"2019-05-08",
	"2019-05-09",
	"2019-05-10",
	"2019-05-11",
	"2019-05-13",
	"2019-05-14",
	"2019-05-15",
	"2019-05-16",
	"2019-05-17",
	"2019-05-18",
	"2019-05-20",
	"2019-05-21",
	"2019-05-22",
	"2019-05-23",
	"2019-05-24",
	"2019-05-25",
	"2019-05-28",
	"2019-05-29",
	"2019-05-30",
	"2019-05-31",
	"2019-06-01",
	"2019-06-03",
	"2019-06-04",
	"2019-06-05",
	"2019-06-06",
	"2019-06-07",
	"2019-06-08",
	"2019-06-10",
	"2019-06-11",
	"2019-06-12",
	"2019-06-13",
	"2019-06-14",
	"2019-06-15",
	"2019-06-17",
	"2019-06-18",
	"2019-06-19",
	"2019-06-20",
	"2019-06-21",
	"2019-06-22",
	"2019-06-24",
	"2019-06-25",
	"2019-06-26",
	"2019-06-27",
	"2019-06-28",
	"2019-06-29",
	"2019-07-01",
	"2019-07-02",
	"2019-07-03",
	"2019-07-05",
	"2019-07-06",
	"2019-07-08",
	"2019-07-09",
	"2019-07-10",
	"2019-07-11",
	"2019-07-12",
	"2019-07-13",
	"2019-07-15",
	"2019-07-16",
	"2019-07-17",
	"2019-07-18",
	"2019-07-19",
	"2019-07-20",
	"2019-07-22",
	"2019-07-23",
	"2019-07-24",
	"2019-07-25",
	"2019-07-26",
	"2019-07-27",
	"2019-07-29",
	"2019-07-30",
	"2019-07-31",
	"2019-08-01",
	"2019-08-02",
	"2019-08-03",
	"2019-08-05",
	"2019-08-06",
	"2019-08-07",
	"2019-08-08",
	"2019-08-09",
	"2019-08-10",
	"2019-08-12",
	"2019-08-13",
	"2019-08-14",
	"2019-08-15",
	"2019-08-16",
	"2019-08-17",
	"2019-08-19",
	"2019-08-20",
	"2019-08-21",
	"2019-08-22",
	"2019-08-23",
	"2019-08-24",
	"2019-08-26",
	"2019-08-27",
	"2019-08-28",
	"2019-08-29",
	"2019-08-30",
	"2019-08-31",
	"2019-09-03",
	"2019-09-04",
	"2019-09-05",
	"2019-09-06",
	"2019-09-07",
	"2019-09-09",
	"2019-09-10",
	"2019-09-11",
	"2019-09-12",
	"2019-09-13",
	"2019-09-14",
	"2019-09-16",
	"2019-09-17",
	"2019-09-18",
	"2019-09-19",
	"2019-09-20",
	"2019-09-21",
	"2019-09-23",
	"2019-09-24",
	"2019-09-25",
	"2019-09-26",
	"2019-09-27",
	"2019-09-28",
	"2019-09-30",
	"2019-10-01",
	"2019-10-02",
	"2019-10-03",
	"2019-10-04",
	"2019-10-05",
	"2019-10-07",
	"2019-10-08",
	"2019-10-09",
	"2019-10-10",
	"2019-10-11",
	"2019-10-12",
	"2019-10-14",
	"2019-10-15",
	"2019-10-16",
	"2019-10-17",
	"2019-10-18",
	"2019-10-19",
	"2019-10-21",
	"2019-10-22",
	"2019-10-23",
	"2019-10-24",
	"2019-10-25",
	"2019-10-26",
	"2019-10-28",
	"2019-10-29",
	"2019-10-30",
	"2019-10-31",
	"2019-11-01",
	"2019-11-02",
	"2019-11-04",
	"2019-11-05",
	"2019-11-06",
	"2019-11-07",
	"2019-11-08",
	"2019-11-09",
	"2019-11-11",
	"2019-11-12",
	"2019-11-13",
	"2019-11-14",
	"2019-11-15",
	"2019-11-16",
	"2019-11-18",
	"2019-11-19",
	"2019-11-20",
	"2019-11-21",
	"2019-11-22",
	"2019-11-23",
	"2019-11-25",
	"2019-11-26",
	"2019-11-27",
	"2019-11-29",
	"2019-11-30",
	"2019-12-02",
	"2019-12-03",
	"2019-12-04",
	"2019-12-05",
	"2019-12-06",
	"2019-12-07",
	"2019-12-09",
	"2019-12-10",
	"2019-12-11",
	"2019-12-12",
	"2019-12-13",
	"2019-12-14",
	"2019-12-16",
	"2019-12-17",
	"2019-12-18",
	"2019-12-19",
	"2019-12-20",
	"2019-12-21",
	"2019-12-23",
	"2019-12-24",
	"2019-12-26",
	"2019-12-27",
	"2019-12-28",
	"2019-12-30",
	"2019-12-31",
	"2020-01-02",
	"2020-01-03",
	"2020-01-04",
	"2020-01-06",
	"2020-01-07",
	"2020-01-08",
	"2020-01-09",
	"2020-01-10",
	"2020-01-11",
	"2020-01-13",
	"2020-01-14",
	"2020-01-15",
	"2020-01-16",
	"2020-01-17",
	"2020-01-18",
	"2020-01-21",
	"2020-01-22",
	"2020-01-23",
	"2020-01-24",
	"2020-01-25",
	"2020-01-27",
	"2020-01-28",
	"2020-01-29",
	"2020-01-30",
	"2020-01-31",
	"2020-02-01",
	"2020-02-03",
	"2020-02-04",
	"2020-02-05",
	"2020-02-06",
	"2020-02-07",
	"2020-02-08",
	"2020-02-10",
	"2020-02-11",
	"2020-02-12",
	"2020-02-13",
	"2020-02-14",
	"2020-02-15",
	"2020-02-18",
	"2020-02-19",
	"2020-02-20",
	"2020-02-21",
	"2020-02-22",
	"2020-02-24",
	"2020-02-25",
	"2020-02-26",
	"2020-02-27",
	"2020-02-28",
	"2020-02-29",
	"2020-03-02",
	"2020-03-03",
	"2020-03-04",
	"2020-03-05",
	"2020-03-06",
	"2020-03-07",
	"2020-03-09",
	"2020-03-10",
	"2020-03-11",
	"2020-03-12",
	"2020-03-13",
	"2020-03-14",
	"2020-03-16",
	"2020-03-17",
	"2020-03-18",
	"2020-03-19",
	"2020-03-20",
	"2020-03-21",
	"2020-03-23",
	"2020-03-24",
	"2020-03-25",
	"2020-03-26",
	"2020-03-27",
	"2020-03-28",
	"2020-03-30",
	"2020-03-31",
	"2020-04-01",
	"2020-04-02",
	"2020-04-03",
	"2020-04-04",
	"2020-04-06",
	"2020-04-07",
	"2020-04-08",
	"2020-04-09",
	"2020-04-10",
	"2020-04-11",
	"2020-04-13",
	"2020-04-14",
	"2020-04-15",
	"2020-04-16",
	"2020-04-17",
	"2020-04-18",
	"2020-04-20",
	"2020-04-21",
	"2020-04-22",
	"2020-04-23",
	"2020-04-24",
	"2020-04-25",
	"2020-04-27",
	"2020-04-28",
	"2020-04-29",
	"2020-04-30",
	"2020-05-01",
	"2020-05-02",
	"2020-05-04",
	"2020-05-05",
	"2020-05-06",
	"2020-05-07",
	"2020-05-08",
	"2020-05-09",
	"2020-05-11",
	"2020-05-12",
	"2020-05-13",
	"2020-05-14",
	"2020-05-15",
	"2020-05-16",
	"2020-05-18",
	"2020-05-19",
	"2020-05-20",
	"2020-05-21",
	"2020-05-22",
	"2020-05-23",
	"2020-05-26",
	"2020-05-27",
	"2020-05-28",
	"2020-05-29",
	"2020-05-30",
	"2020-06-01",
	"2020-06-02",
	"2020-06-03",
	"2020-06-04",
	"2020-06-05",
	"2020-06-06",
	"2020-06-08",
	"2020-06-09",
	"2020-06-10",
	"2020-06-11",
	"2020-06-12",
	"2020-06-13",
	"2020-06-15",
	"2020-06-16",
	"2020-06-17",
	"2020-06-18",
	"2020-06-19",
	"2020-06-20",
	"2020-06-22",
	"2020-06-23",
	"2020-06-24",
	"2020-06-25",
	"2020-06-26",
	"2020-06-27",
	"2020-06-29",
	"2020-06-30",
	"2020-07-01",
	"2020-07-02",
	"2020-07-03",
	"2020-07-06",
	"2020-07-07",
	"2020-07-08",
	"2020-07-09",
	"2020-07-10",
	"2020-07-11",
	"2020-07-13",
	"2020-07-14",
	"2020-07-15",
	"2020-07-16",
	"2020-07-17",
	"2020-07-18",
	"2020-07-20",
	"2020-07-21",
	"2020-07-22",
	"2020-07-23",
	"2020-07-24",
	"2020-07-25",
	"2020-07-27",
	"2020-07-28",
	"2020-07-29",
	"2020-07-30",
	"2020-07-31",
	"2020-08-01",
	"2020-08-03",
	"2020-08-04",
	"2020-08-05",
	"2020-08-06",
	"2020-08-07",
	"2020-08-08",
	"2020-08-10",
	"2020-08-11",
	"2020-08-12",
	"2020-08-13",
	"2020-08-14",
	"2020-08-15",
	"2020-08-17",
	"2020-08-18",
	"2020-08-19",
	"2020-08-20",
	"2020-08-21",
	"2020-08-22",
	"2020-08-24",
	"2020-08-25",
	"2020-08-26",
	"2020-08-27",
	"2020-08-28",
	"2020-08-29",
	"2020-08-31",
	"2020-09-01",
	"2020-09-02",
	"2020-09-03",
	"2020-09-04",
	"2020-09-05",
	"2020-09-08",
	"2020-09-09",
	"2020-09-10",
	"2020-09-11",
	"2020-09-12",
	"2020-09-14",
	"2020-09-15",
	"2020-09-16",
	"2020-09-17",
	"2020-09-18",
	"2020-09-19",
	"2020-09-21",
	"2020-09-22",
	"2020-09-23",
	"2020-09-24",
	"2020-09-25",
	"2020-09-26",
	"2020-09-28",
	"2020-09-29",
	"2020-09-30",
	"2020-10-01",
	"2020-10-02",
	"2020-10-03",
	"2020-10-05",
	"2020-10-06",
	"2020-10-07",
	"2020-10-08",
	"2020-10-09",
	"2020-10-10",
	"2020-10-12",
	"2020-10-13",
	"2020-10-14",
	"2020-10-15",
	"2020-10-16",
	"2020-10-17",
	"2020-10-19",
	"2020-10-20",
	"2020-10-21",
	"2020-10-22",
	"2020-10-23",
	"2020-10-24",
	"2020-10-26",
	"2020-10-27",
	"2020-10-28",
	"2020-10-29",
	"2020-10-30",
	"2020-10-31",
	"2020-11-02",
	"2020-11-03",
	"2020-11-04",
	"2020-11-05",
	"2020-11-06",
	"2020-11-07",
	"2020-11-09",
	"2020-11-10",
	"2020-11-11",
	"2020-11-12",
	"2020-11-13",
	"2020-11-14",
	"2020-11-16",
	"2020-11-17",
	"2020-11-18",
	"2020-11-19",
	"2020-11-20",
	"2020-11-21",
	"2020-11-23",
	"2020-11-24",
	"2020-11-25",
	"2020-11-27",
	"2020-11-28",
	"2020-11-30",
	"2020-12-01",
	"2020-12-02",
	"2020-12-03",
	"2020-12-04",
	"2020-12-05",
	"2020-12-07",
	"2020-12-08",
	"2020-12-09",
	"2020-12-10",
	"2020-12-11",
	"2020-12-12",
	"2020-12-14",
	"2020-12-15",
	"2020-12-16",
	"2020-12-17",
	"2020-12-18",
	"2020-12-19",
	"2020-12-21",
	"2020-12-22",
	"2020-12-23",
	"2020-12-24",
	"2020-12-26",
	"2020-12-28",
	"2020-12-29",
	"2020-12-30",
	"2020-12-31",
	"2021-01-02",
	"2021-01-04",
	"2021-01-05",
	"2021-01-06",
	"2021-01-07",
	"2021-01-08",
	"2021-01-09",
	"2021-01-11",
	"2021-01-12",
	"2021-01-13",
	"2021-01-14",
	"2021-01-15",
	"2021-01-16",
	"2021-01-19",
	"2021-01-20",
	"2021-01-21",
	"2021-01-22",
	"2021-01-23",
	"2021-01-25",
	"2021-01-26",
	"2021-01-27",
	"2021-01-28",
	"2021-01-29",
	"2021-01-30",
	"2021-02-01",
	"2021-02-02",
	"2021-02-03",
	"2021-02-04",
	"2021-02-05",
	"2021-02-06",
	"2021-02-08",
	"2021-02-09",
	"2021-02-10",
	"2021-02-11",
	"2021-02-12",
	"2021-02-13",
	"2021-02-16",
	"2021-02-17",
	"2021-02-18",
	"2021-02-19",
	"2021-02-20",
	"2021-02-22",
	"2021-02-23",
	"2021-02-24",
	"2021-02-25",
	"2021-02-26",
	"2021-02-27",
	"2021-03-01",
	"2021-03-02",
	"2021-03-03",
	"2021-03-04",
	"2021-03-05",
	"2021-03-06",
	"2021-03-08",
	"2021-03-09",
	"2021-03-10",
	"2021-03-11",
	"2021-03-12",
	"2021-03-13",
	"2021-03-15",
	"2021-03-16",
	"2021-03-17",
	"2021-03-18",
	"2021-03-19",
	"2021-03-20",
	"2021-03-22",
	"2021-03-23",
	"2021-03-24",
	"2021-03-25",
	"2021-03-26",
	"2021-03-27",
	"2021-03-29",
	"2021-03-30",
	"2021-03-31",
	"2021-04-01",
	"2021-04-02",
	"2021-04-03",
	"2021-04-05",
	"2021-04-06",
	"2021-04-07",
	"2021-04-08",
	"2021-04-09",
	"2021-04-10",
	"2021-04-12",
	"2021-04-13",
	"2021-04-14",
	"2021-04-15",
	"2021-04-16",
	"2021-04-17",
	"2021-04-19",
	"2021-04-20",
	"2021-04-21",
	"2021-04-22",
	"2021-04-23",
	"2021-04-24",
	"2021-04-26",
	"2021-04-27",
	"2021-04-28",
	"2021-04-29",
	"2021-04-30",
	"2021-05-01",
	"2021-05-03",
	"2021-05-04",
	"2021-05-05",
	"2021-05-06",
	"2021-05-07",
	"2021-05-08",
	"2021-05-10",
	"2021-05-11",
	"2021-05-12",
	"2021-05-13",
	"2021-05-14",
	"2021-05-15",
	"2021-05-17",
	"2021-05-18",
	"2021-05-19",
	"2021-05-20",
	"2021-05-21",
	"2021-05-22",
	"2021-05-24",
	"2021-05-25",
	"2021-05-26",
	"2021-05-27",
	"2021-05-28",
	"2021-05-29",
	"2021-06-01",
	"2021-06-02",
	"2021-06-03",
	"2021-06-04",
	"2021-06-05",
	"2021-06-07",
	"2021-06-08",
	"2021-06-09",
	"2021-06-10",
	"2021-06-11",
	"2021-06-12",
	"2021-06-14",
	"2021-06-15",
	"2021-06-16",
	"2021-06-17",
	"2021-06-18",
	"2021-06-19",
	"2021-06-21",
	"2021-06-22",
	"2021-06-23",
	"2021-06-24",
	"2021-06-25",
	"2021-06-26",
	"2021-06-28",
	"2021-06-29",
	"2021-06-30",
	"2021-07-01",
	"2021-07-02",
	"2021-07-03",
	"2021-07-06",
	"2021-07-07",
	"2021-07-08",
	"2021-07-09",
	"2021-07-10",
	"2021-07-12",
	"2021-07-13",
	"2021-07-14",
	"2021-07-15",
	"2021-07-16",
	"2021-07-17",
	"2021-07-19",
	"2021-07-20",
	"2021-07-21",
	"2021-07-22",
	"2021-07-23",
	"2021-07-24",
	"2021-07-26",
	"2021-07-27",
	"2021-07-28",
	"2021-07-29",
	"2021-07-30",
	"2021-07-31",
	"2021-08-02",
	"2021-08-03",
	"2021-08-04",
	"2021-08-05",
	"2021-08-06",
	"2021-08-07",
	"2021-08-09",
	"2021-08-10",
	"2021-08-11",
	"2021-08-12",
	"2021-08-13",
	"2021-08-14",
	"2021-08-16",
	"2021-08-17",
	"2021-08-18",
	"2021-08-19",
	"2021-08-20",
	"2021-08-21",
	"2021-08-23",
	"2021-08-24",
	"2021-08-25",
	"2021-08-26",
	"2021-08-27",
	"2021-08-28",
	"2021-08-30",
	"2021-08-31",
	"2021-09-01",
	"2021-09-02",
	"2021-09-03",
	"2021-09-04",
	"2021-09-07",
	"2021-09-08",
	"2021-09-09",
	"2021-09-10",
	"2021-09-11",
	"2021-09-13",
	"2021-09-14",
	"2021-09-15",
	"2021-09-16",
	"2021-09-17",
	"2021-09-18",
	"2021-09-20",
	"2021-09-21",
	"2021-09-22",
	"2021-09-23",
	"2021-09-24",
	"2021-09-25",
	"2021-09-27",
	"2021-09-28",
	"2021-09-29",
	"2021-09-30",
	"2021-10-01",
	"2021-10-02",
	"2021-10-04",
	"2021-10-05",
	"2021-10-06",
	"2021-10-07",
	"2021-10-08",
	"2021-10-09",
	"2021-10-11",
	"2021-10-12",
	"2021-10-13",
	"2021-10-14",
	"2021-10-15",
	"2021-10-16",
	"2021-10-18",
	"2021-10-19",
	"2021-10-20",
	"2021-10-21",
	"2021-10-22",
	"2021-10-23",
	"2021-10-25",
	"2021-10-26",
	"2021-10-27",
	"2021-10-28",
	"2021-10-29",
	"2021-10-30",
	"2021-11-01",
	"2021-11-02",
	"2021-11-03",
	"2021-11-04",
	"2021-11-05",
	"2021-11-06",
	"2021-11-08",
	"2021-11-09",
	"2021-11-10",
	"2021-11-11",
	"2021-11-12",
	"2021-11-13",
	"2021-11-15",
	"2021-11-16",
	"2021-11-17",
	"2021-11-18",
	"2021-11-19",
	"2021-11-20",
	"2021-11-22",
	"2021-11-23",
	"2021-11-24",
	"2021-11-26",
	"2021-11-27",
	"2021-11-29",
	"2021-11-30",
	"2021-12-01",
	"2021-12-02",
	"2021-12-03",
	"2021-12-04",
	"2021-12-06",
	"2021-12-07",
	"2021-12-08",
	"2021-12-09",
	"2021-12-10",
	"2021-12-11",
	"2021-12-13",
	"2021-12-14",
	"2021-12-15",
	"2021-12-16",
	"2021-12-17",
	"2021-12-18",
	"2021-12-20",
	"2021-12-21",
	"2021-12-22",
	"2021-12-23",
	"2021-12-24",
	"2021-12-27",
	"2021-12-28",
	"2021-12-29",
	"2021-12-30",
	"2021-12-31",
	"2022-01-03",
	"2022-01-04",
	"2022-01-05",
	"2022-01-06",
	"2022-01-07",
	"2022-01-08",
	"2022-01-10",
	"2022-01-11",
	"2022-01-12",
	"2022-01-13",
	"2022-01-14",
	"2022-01-15",
	"2022-01-18",
	"2022-01-19",
	"2022-01-20",
	"2022-01-21",
	"2022-01-22",
	"2022-01-24",
	"2022-01-25",
	"2022-01-26",
	"2022-01-27",
	"2022-01-28",
	"2022-01-29",
	"2022-01-31",
	"2022-02-01",
	"2022-02-02",
	"2022-02-03",
	"2022-02-04",
	"2022-02-05",
	"2022-02-07",
	"2022-02-08",
	"2022-02-09",
	"2022-02-10",
	"2022-02-11",
	"2022-02-12",
	"2022-02-14",
	"2022-02-15",
	"2022-02-16",
	"2022-02-17",
	"2022-02-18",
	"2022-02-19",
	"2022-02-22",
	"2022-02-23",
	"2022-02-24",
	"2022-02-25",
	"2022-02-26",
	"2022-02-28",
	"2022-03-01",
	"2022-03-02",
	"2022-03-03",
	"2022-03-04",
	"2022-03-05",
	"2022-03-07",
	"2022-03-08",
	"2022-03-09",
	"2022-03-10",
	"2022-03-11",
	"2022-03-12",
	"2022-03-14",
	"2022-03-15",
	"2022-03-16",
	"2022-03-17",
	"2022-03-18",
	"2022-03-19",
	"2022-03-21",
	"2022-03-22",
	"2022-03-23",
	"2022-03-24",
	"2022-03-25",
	"2022-03-26",
	"2022-03-28",
	"2022-03-29",
	"2022-03-30",
	"2022-03-31",
	"2022-04-01",
	"2022-04-02",
	"2022-04-04",
	"2022-04-05",
	"2022-04-06",
	"2022-04-07",
	"2022-04-08",
	"2022-04-09",
	"2022-04-11",
	"2022-04-12",
	"2022-04-13",
	"2022-04-14",
	"2022-04-15",
	"2022-04-16",
	"2022-04-18",
	"2022-04-19",
	"2022-04-20",
	"2022-04-21",
	"2022-04-22",
	"2022-04-23",
	"2022-04-25",
	"2022-04-26",
	"2022-04-27",
	"2022-04-28",
	"2022-04-29",
	"2022-04-30",
	"2022-05-02",
	"2022-05-03",
	"2022-05-04",
	"2022-05-05",
	"2022-05-06",
	"2022-05-07",
	"2022-05-09",
	"2022-05-10",
	"2022-05-11",
	"2022-05-12",
	"2022-05-13",
	"2022-05-14",
	"2022-05-16",
	"2022-05-17",
	"2022-05-18",
	"2022-05-19",
	"2022-05-20",
	"2022-05-21",
	"2022-05-23",
	"2022-05-24",
	"2022-05-25",
	"2022-05-26",
	"2022-05-27",
	"2022-05-28",
	"2022-05-31",
	"2022-06-01",
	"2022-06-02",
	"2022-06-03",
	"2022-06-04",
	"2022-06-06",
	"2022-06-07",
	"2022-06-08",
	"2022-06-09",
	"2022-06-10",
	"2022-06-11",
	"2022-06-13",
	"2022-06-14",
	"2022-06-15",
	"2022-06-16",
	"2022-06-17",
	"2022-06-18",
	"2022-06-20",
	"2022-06-21",
	"2022-06-22",
	"2022-06-23",
	"2022-06-24",
	"2022-06-25",
	"2022-06-27",
	"2022-06-28",
	"2022-06-29",
	"2022-06-30",
	"2022-07-01",
	"2022-07-02",
	"2022-07-05",
	"2022-07-06",
	"2022-07-07",
	"2022-07-08",
	"2022-07-09",
	"2022-07-11",
	"2022-07-12",
	"2022-07-13",
	"2022-07-14",
	"2022-07-15",
	"2022-07-16",
	"2022-07-18",
	"2022-07-19",
	"2022-07-20",
	"2022-07-21",
	"2022-07-22",
	"2022-07-23",
	"2022-07-25",
	"2022-07-26",
	"2022-07-27",
	"2022-07-28",
	"2022-07-29",
	"2022-07-30",
	"2022-08-01",
	"2022-08-02",
	"2022-08-03",
	"2022-08-04",
	"2022-08-05",
	"2022-08-06",
	"2022-08-08",
	"2022-08-09",
	"2022-08-10",
	"2022-08-11",
	"2022-08-12",
	"2022-08-13",
	"2022-08-15",
	"2022-08-16",
	"2022-08-17",
	"2022-08-18",
	"2022-08-19",
	"2022-08-20",
	"2022-08-22",
	"2022-08-23",
	"2022-08-24",
	"2022-08-25",
	"2022-08-26",
	"2022-08-27",
	"2022-08-29",
	"2022-08-30",
	"2022-08-31",
	"2022-09-01",
	"2022-09-02",
	"2022-09-03",
	"2022-09-06",
	"2022-09-07",
	"2022-09-08",
	"2022-09-09",
	"2022-09-10",
	"2022-09-12",
	"2022-09-13",
	"2022-09-14",
	"2022-09-15",
	"2022-09-16",
	"2022-09-17",
	"2022-09-19",
	"2022-09-20",
	"2022-09-21",
	"2022-09-22",
	"2022-09-23",
	"2022-09-24",
	"2022-09-26",
	"2022-09-27",
	"2022-09-28",
	"2022-09-29",
	"2022-09-30",
	"2022-10-01",
	"2022-10-03",
	"2022-10-04",
	"2022-10-05",
	"2022-10-06",
	"2022-10-07",
	"2022-10-08",
	"2022-10-10",
	"2022-10-11",
	"2022-10-12",
	"2022-10-13",
	"2022-10-14",
	"2022-10-15",
	"2022-10-17",
	"2022-10-18",
	"2022-10-19",
	"2022-10-20",
	"2022-10-21",
	"2022-10-22",
	"2022-10-24",
	"2022-10-25",
	"2022-10-26",
	"2022-10-27",
	"2022-10-28",
	"2022-10-29",
	"2022-10-31",
	"2022-11-01",
	"2022-11-02",
	"2022-11-03",
	"2022-11-04",
	"2022-11-05",
	"2022-11-07",
	"2022-11-08",
	"2022-11-09",
	"2022-11-10",
	"2022-11-11",
	"2022-11-12",
	"2022-11-14",
	"2022-11-15",
	"2022-11-16",
	"2022-11-17",
	"2022-11-18",
	"2022-11-19",
	"2022-11-21",
	"2022-11-22",
	"2022-11-23",
	"2022-11-25",
	"2022-11-26",
	"2022-11-28",
	"2022-11-29",
	"2022-11-30",
	"2022-12-01",
	"2022-12-02",
	"2022-12-03",
	"2022-12-05",
	"2022-12-06",
	"2022-12-07",
	"2022-12-08",
	"2022-12-09",
	"2022-12-10",
	"2022-12-12",
	"2022-12-13",
	"2022-12-14",
	"2022-12-15",
	"2022-12-16",
	"2022-12-17",
	"2022-12-19",
	"2022-12-20",
	"2022-12-21",
	"2022-12-22",
	"2022-12-23",
	"2022-12-24",
	"2022-12-27",
	"2022-12-28",
	"2022-12-29",
	"2022-12-30",
	"2022-12-31",
}
