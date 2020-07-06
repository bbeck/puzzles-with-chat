package acrostic

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"sort"
	"strings"
	"testing"
	"testing/iotest"
	"time"
)

func TestParseAuthorAndTitle(t *testing.T) {
	tests := []struct {
		name           string
		quote          string
		expectedAuthor string
		expectedTitle  string
	}{
		{
			name:           "basic author",
			quote:          "KEN DRUSE, THE NEW SHADE GARDEN — Plants are moving...",
			expectedAuthor: "KEN DRUSE",
			expectedTitle:  "THE NEW SHADE GARDEN",
		},
		{
			name:           "parenthesized author first name",
			quote:          "(MABEL) WAGNALLS, STARS OF THE OPERA — People seldom appreciate the vast knowledge...",
			expectedAuthor: "MABEL WAGNALLS",
			expectedTitle:  "STARS OF THE OPERA",
		},
		{
			name:           "hyphenated last name",
			quote:          "DORIS NASH-WORTMAN, TITLE — Quote...",
			expectedAuthor: "DORIS NASH-WORTMAN",
			expectedTitle:  "TITLE",
		},
		{
			name:           "parenthesized first name, hyphenated last name",
			quote:          "(DORIS) NASH-WORTMAN, TITLE — Quote...",
			expectedAuthor: "DORIS NASH-WORTMAN",
			expectedTitle:  "TITLE",
		},
		{
			name:           "basic title",
			quote:          "KEN DRUSE, THE NEW SHADE GARDEN — Plants are moving...",
			expectedAuthor: "KEN DRUSE",
			expectedTitle:  "THE NEW SHADE GARDEN",
		},
		{
			name:           "quote in author",
			quote:          "CONAN O'BRIEN, IN THE YEAR 2000 — ...",
			expectedAuthor: "CONAN O'BRIEN",
			expectedTitle:  "IN THE YEAR 2000",
		},
		{
			name:           "html character code in author",
			quote:          "CONAN O&#39;BRIEN, IN THE YEAR 2000 — ...",
			expectedAuthor: "CONAN O'BRIEN",
			expectedTitle:  "IN THE YEAR 2000",
		},
		{
			name:           "quote in title",
			quote:          "FRANS DE WAAL, MAMA'S LAST HUG — An Internet video of a...",
			expectedAuthor: "FRANS DE WAAL",
			expectedTitle:  "MAMA'S LAST HUG",
		},
		{
			name:           "html character code in title",
			quote:          "FRANS DE WAAL, MAMA&#39;S LAST HUG — An Internet video of a...",
			expectedAuthor: "FRANS DE WAAL",
			expectedTitle:  "MAMA'S LAST HUG",
		},
		{
			name:           "2000-01-02",
			quote:          "S(TEPHEN) J(AY) GOULD: DINOSAUR IN A HAYSTACK — A sixth-century monk ...",
			expectedAuthor: "STEPHEN JAY GOULD",
			expectedTitle:  "DINOSAUR IN A HAYSTACK",
		},
		{
			name:           "2000-01-16",
			quote:          "L(ANGSTON) HUGHES: SIMPLE TAKES A WIFE — Bop is ... not to be dug unless you&#39;ve seen dark days, too. Folks who ain&#39;t suffered ... think it&#39;s just crazy crazy. They do not know Bop is also MAD crazy, SAD crazy, FRANTIC WILD CRAZY — beat out of somebody&#39;s head! That&#39;s what Bop is.",
			expectedAuthor: "LANGSTON HUGHES",
			expectedTitle:  "SIMPLE TAKES A WIFE",
		},
		{
			name:           "2000-01-30",
			quote:          "It&#39;s impossible to duplicate... the conditions",
			expectedAuthor: "",
			expectedTitle:  "",
		},
		{
			name:           "2000-02-13",
			quote:          "DAVE BARRY TURNS FORTY — It&#39;s not easy to maintain",
			expectedAuthor: "",
			expectedTitle:  "DAVE BARRY TURNS FORTY",
		},
		{
			name:           "2000-03-26",
			quote:          "(JONATHAN) LETHEM, MOTHERLESS BROOKLYN - I&#39;ve got Tourette&#39;s.",
			expectedAuthor: "JONATHAN LETHEM",
			expectedTitle:  "MOTHERLESS BROOKLYN",
		},
		{
			name:           "2001-01-14",
			quote:          "(BILL) BRYSON: THE MOTHER TONGUE — The average English speaker ...",
			expectedAuthor: "BILL BRYSON",
			expectedTitle:  "THE MOTHER TONGUE",
		},
		{
			name:           "2003-04-20",
			quote:          "DIANE ACKERMAN, MUTE DANCERS (HOW TO WATCH A HUMMINGBIRD) - [Hummingbirds] spell out their ...",
			expectedAuthor: "DIANE ACKERMAN",
			expectedTitle:  "MUTE DANCERS HOW TO WATCH A HUMMINGBIRD",
		},
		{
			name:           "2004-03-21",
			quote:          "THE WPA GUIDE TO FLORIDA — Diddy-Wah-Diddy ...",
			expectedAuthor: "",
			expectedTitle:  "THE WPA GUIDE TO FLORIDA",
		},
		{
			name:           "2008-01-06",
			quote:          "[HARPER] LEE, TO KILL A MOCKINGBIRD — Never...",
			expectedAuthor: "HARPER LEE",
			expectedTitle:  "TO KILL A MOCKINGBIRD",
		},
		{
			name:           "2010-03-14",
			quote:          "(SYLVIA TOWNSEND) WARNER, MR. FORTUNE&#39;S MAGGOT - Most Englishmen ...",
			expectedAuthor: "SYLVIA TOWNSEND WARNER",
			expectedTitle:  "MR. FORTUNE'S MAGGOT",
		},
		{
			name:           "2016-01-24",
			quote:          "(V.S.) RAMACHANDRAN, (THE) TELL-TALE BRAIN — How can a three-pound ...",
			expectedAuthor: "V.S. RAMACHANDRAN",
			expectedTitle:  "THE TELL-TALE BRAIN",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			author, title := ParseAuthorAndTitle(test.quote)
			assert.Equal(t, test.expectedAuthor, author)
			assert.Equal(t, test.expectedTitle, title)
		})
	}
}

func TestParseQuote(t *testing.T) {
	tests := []struct {
		name     string
		quote    string
		expected string
	}{
		{
			name:     "basic quote",
			quote:    "KEN DRUSE, THE NEW SHADE GARDEN — Plants are moving...",
			expected: "Plants are moving...",
		},
		{
			name:     "trims leading whitespace",
			quote:    "KEN DRUSE, THE NEW SHADE GARDEN —  Plants are moving...",
			expected: "Plants are moving...",
		},
		{
			name:     "trims trailing whitespace",
			quote:    "KEN DRUSE, THE NEW SHADE GARDEN — Plants are moving... ",
			expected: "Plants are moving...",
		},
		{
			name:     "html character code",
			quote:    "A LITTLE BOOK OF YANKEE HUMOR — &quot;I don&#39;t know about your...",
			expected: `"I don't know about your...`,
		},
		{
			name:     "quote characters",
			quote:    `A LITTLE BOOK OF YANKEE HUMOR — "I don't know about your...`,
			expected: `"I don't know about your...`,
		},
		{
			name:     "no author or title",
			quote:    "No author or title...",
			expected: "No author or title...",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quote := ParseQuote(test.quote)
			assert.Equal(t, test.expected, quote)
		})
	}
}

func TestGetClueLetter(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		expected string
	}{
		{name: "A", index: 0, expected: "A"},
		{name: "B", index: 1, expected: "B"},
		{name: "C", index: 2, expected: "C"},
		{name: "D", index: 3, expected: "D"},
		{name: "E", index: 4, expected: "E"},
		{name: "F", index: 5, expected: "F"},
		{name: "G", index: 6, expected: "G"},
		{name: "H", index: 7, expected: "H"},
		{name: "I", index: 8, expected: "I"},
		{name: "J", index: 9, expected: "J"},
		{name: "K", index: 10, expected: "K"},
		{name: "L", index: 11, expected: "L"},
		{name: "M", index: 12, expected: "M"},
		{name: "N", index: 13, expected: "N"},
		{name: "O", index: 14, expected: "O"},
		{name: "P", index: 15, expected: "P"},
		{name: "Q", index: 16, expected: "Q"},
		{name: "R", index: 17, expected: "R"},
		{name: "S", index: 18, expected: "S"},
		{name: "T", index: 19, expected: "T"},
		{name: "U", index: 20, expected: "U"},
		{name: "V", index: 21, expected: "V"},
		{name: "W", index: 22, expected: "W"},
		{name: "X", index: 23, expected: "X"},
		{name: "Y", index: 24, expected: "Y"},
		{name: "Z", index: 25, expected: "Z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := GetClueLetter(test.index)
			require.NoError(t, err)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetClueLetter_Error(t *testing.T) {
	tests := []struct {
		name  string
		index int
	}{
		{name: "-1", index: -1},
		{name: "26", index: 26},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := GetClueLetter(test.index)
			require.Error(t, err)
		})
	}
}

func TestParseInts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
	}{
		{
			name:     "one int",
			input:    "1",
			expected: []int{1},
		},
		{
			name:     "multiple ints",
			input:    "1,2,3,12",
			expected: []int{1, 2, 3, 12},
		},
		{
			name:     "non-sorted ints",
			input:    "12,3,1,2",
			expected: []int{12, 3, 1, 2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ParseInts(test.input)
			require.NoError(t, err)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestParseInts_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "non integer",
			input: "a",
		},
		{
			name:  "non integer embedded within integers",
			input: "1,2,3,a,5",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseInts(test.input)
			require.Error(t, err)
		})
	}
}

func TestParseXWordInfoPuzzleResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  io.ReadCloser
		verify func(t *testing.T, puzzle *Puzzle)
	}{
		{
			name:  "description",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := "New York Times puzzle from 2020-05-24"
				assert.Equal(t, expected, puzzle.Description)
			},
		},
		{
			name:  "size",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 27, puzzle.Cols)
				assert.Equal(t, 8, puzzle.Rows)
			},
		},
		{
			name:  "publisher",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "The New York Times", puzzle.Publisher)
			},
		},
		{
			name:  "published date",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, 2020, puzzle.PublishedDate.Year())
				assert.Equal(t, time.May, puzzle.PublishedDate.Month())
				assert.Equal(t, 24, puzzle.PublishedDate.Day())
			},
		},
		{
			name:  "author",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "MABEL WAGNALLS", puzzle.Author)
			},
		},
		{
			name:  "title",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "STARS OF THE OPERA", puzzle.Title)
			},
		},
		{
			name:  "title with HTML character entity",
			input: load(t, "xwordinfo-nyt-20200510-html-character-in-title.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				assert.Equal(t, "MAMA'S LAST HUG", puzzle.Title)
			},
		},
		{
			name:  "quote",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := "<p>People seldom appreciate the vast knowledge of music " +
					"and the remarkable ability in sight-reading which these orchestra " +
					"players possess. Not one of them but has worked at his art from " +
					"childhood; most of them play several different instruments; and " +
					"they all hold as a creed that a false note is a sin, and a " +
					"variation in rhythm is a fall from grace.</p>"
				assert.Equal(t, expected, puzzle.Quote)
			},
		},
		{
			name:  "no full quote field",
			input: load(t, "xwordinfo-nyt-20020811-no-full-quote.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := `"I don't know about your farm in Maine, mister, but I ` +
					`have a ranch in Texas and it takes me five days to drive around ` +
					`my entire spread," says the Texan. The Maine farmer replies, "Oh ` +
					`yes, I have a car just like that myself."`
				assert.Equal(t, expected, puzzle.Quote)
			},
		},
		{
			name:  "cells",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"P", "E", "O", "P", "L", "E", "", "S", "E", "L", "D", "O", "M", "", "A", "P", "P", "R", "E", "C", "I", "A", "T", "E", "", "T", "H"},
					{"E", "", "V", "A", "S", "T", "", "K", "N", "O", "W", "L", "E", "D", "G", "E", "", "W", "H", "I", "C", "H", "", "O", "R", "C", "H"},
					{"E", "S", "T", "R", "A", "", "P", "L", "A", "Y", "E", "R", "S", "", "P", "O", "S", "S", "E", "S", "S", "", "M", "O", "S", "T", ""},
					{"O", "F", "", "T", "H", "E", "M", "", "P", "L", "A", "Y", "", "S", "E", "V", "E", "R", "A", "L", "", "I", "N", "S", "T", "R", "U"},
					{"M", "E", "N", "T", "S", "", "A", "N", "D", "", "T", "H", "E", "Y", "", "A", "L", "L", "", "H", "O", "L", "D", "", "A", "S", ""},
					{"A", "", "C", "R", "E", "E", "D", "", "T", "H", "A", "T", "", "A", "", "F", "A", "L", "S", "E", "", "N", "O", "T", "E", "", "I"},
					{"S", "", "A", "", "S", "I", "N", "", "A", "N", "D", "", "A", "", "V", "A", "R", "I", "A", "T", "I", "O", "N", "", "I", "N", ""},
					{"R", "H", "Y", "T", "H", "M", "", "I", "S", "", "A", "", "F", "A", "L", "L", "", "F", "R", "O", "M", "", "G", "R", "A", "C", "E"},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:  "blocks",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, false, true, false, false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, true, false, false},
					{false, true, false, false, false, false, true, false, false, false, false, false, false, false, false, false, true, false, false, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, true},
					{false, false, true, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, true, false, false, false, false, true, false, false, false, true, false, false, false, false, true, false, false, true},
					{false, true, false, false, false, false, false, true, false, false, false, false, true, false, true, false, false, false, false, false, true, false, false, false, false, true, false},
					{false, true, false, true, false, false, false, true, false, false, false, true, false, true, false, false, false, false, false, false, false, false, false, true, false, false, true},
					{false, false, false, false, false, false, true, false, false, true, false, true, false, false, false, false, true, false, false, false, false, true, false, false, false, false, false},
				}
				assert.Equal(t, expected, puzzle.CellBlocks)
			},
		},
		{
			name:  "cell numbers",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]int{
					{1, 2, 3, 4, 5, 6, 0, 7, 8, 9, 10, 11, 12, 0, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 0, 23, 24},
					{25, 0, 26, 27, 28, 29, 0, 30, 31, 32, 33, 34, 35, 36, 37, 38, 0, 39, 40, 41, 42, 43, 0, 44, 45, 46, 47},
					{48, 49, 50, 51, 52, 0, 53, 54, 55, 56, 57, 58, 59, 0, 60, 61, 62, 63, 64, 65, 66, 0, 67, 68, 69, 70, 0},
					{71, 72, 0, 73, 74, 75, 76, 0, 77, 78, 79, 80, 0, 81, 82, 83, 84, 85, 86, 87, 0, 88, 89, 90, 91, 92, 93},
					{94, 95, 96, 97, 98, 0, 99, 100, 101, 0, 102, 103, 104, 105, 0, 106, 107, 108, 0, 109, 110, 111, 112, 0, 113, 114, 0},
					{115, 0, 116, 117, 118, 119, 120, 0, 121, 122, 123, 124, 0, 125, 0, 126, 127, 128, 129, 130, 0, 131, 132, 133, 134, 0, 135},
					{136, 0, 137, 0, 138, 139, 140, 0, 141, 142, 143, 0, 144, 0, 145, 146, 147, 148, 149, 150, 151, 152, 153, 0, 154, 155, 0},
					{156, 157, 158, 159, 160, 161, 0, 162, 163, 0, 164, 0, 165, 166, 167, 168, 0, 169, 170, 171, 172, 0, 173, 174, 175, 176, 177},
				}
				assert.Equal(t, expected, puzzle.CellNumbers)
			},
		},
		{
			name:  "cell clue letters",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"T", "I", "K", "V", "E", "R", "", "D", "O", "F", "U", "G", "B", "", "N", "Q", "M", "K", "A", "V", "P", "E", "R", "L", "", "S", "H"},
					{"F", "", "D", "U", "W", "Q", "", "T", "M", "B", "A", "I", "J", "P", "C", "L", "", "O", "N", "G", "S", "V", "", "T", "R", "K", "B"},
					{"I", "W", "P", "U", "A", "", "L", "D", "E", "M", "F", "J", "O", "", "H", "K", "V", "Q", "B", "W", "S", "", "E", "N", "A", "I", ""},
					{"F", "O", "", "R", "J", "D", "G", "", "C", "E", "V", "M", "", "W", "S", "T", "Q", "F", "O", "P", "", "K", "D", "H", "L", "V", "U"},
					{"M", "G", "R", "J", "I", "", "N", "W", "E", "", "F", "P", "T", "C", "", "L", "O", "G", "", "M", "S", "Q", "V", "", "D", "B", ""},
					{"J", "", "R", "L", "H", "U", "K", "", "G", "A", "W", "I", "", "P", "", "T", "S", "D", "O", "R", "", "F", "M", "B", "E", "", "G"},
					{"W", "", "P", "", "C", "T", "K", "", "Q", "J", "H", "", "R", "", "S", "B", "U", "D", "I", "G", "V", "T", "P", "", "W", "E", ""},
					{"H", "Q", "C", "F", "D", "N", "", "B", "M", "", "K", "", "G", "W", "A", "O", "", "Q", "T", "V", "U", "", "F", "B", "H", "K", "E"},
				}
				assert.Equal(t, expected, puzzle.CellClueLetters)
			},
		},
		{
			name:  "clues",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[string]string{
					"A": "Sources of nonhuman &quot;songs&quot; on a 1970s album with more than 10 million copies",
					"B": "Rock band called the &quot;Bad Boys from Boston&quot;",
					"C": "Hit Broadway musical with the songs &quot;Let Me Entertain You&quot; and &quot;You Gotta Get a Gimmick&quot;",
					"D": "Country Music Hall of Fame site",
					"E": "Turn in square dancing",
					"F": "Prop whose name comes from a French word meaning &quot;squinting&quot;",
					"G": "Recurrent theme",
					"H": "Taken from a B to a C, say",
					"I": "Fertile area for 1990s grunge music",
					"J": "Capital hosting the Fajr Music Festival ",
					"K": "Equipment for a busker, maybe",
					"L": "What &quot;da capo&quot; tells you to do",
					"M": "Work for a pit crew?",
					"N": "City that hosted jazz&#39;s storied Dreamland Ballroom",
					"O": "Brought off without a single fluff",
					"P": "Nation whose country music is &quot;luk thung&quot;",
					"Q": "Interval from ti to do (2 wds.)",
					"R": "Pause for a change of scenery, perhaps",
					"S": "Jumps a prima donna may make",
					"T": "&quot;Peter and the Wolf&quot; composer",
					"U": "Catcher of some waves",
					"V": "Affected by &quot;Ode to Joy,&quot; perhaps",
					"W": "Killer musical by Stephen Sondheim",
				}
				assert.Equal(t, expected, puzzle.Clues)
			},
		},
		{
			name:  "clue numbers",
			input: load(t, "xwordinfo-nyt-20200524.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := map[string][]int{
					"A": {33, 122, 52, 167, 17, 69},
					"B": {146, 64, 174, 32, 114, 12, 162, 133, 47},
					"C": {37, 105, 77, 138, 158},
					"D": {89, 113, 7, 160, 26, 148, 54, 128, 75},
					"E": {55, 5, 78, 134, 67, 20, 155, 101, 177},
					"F": {9, 71, 85, 173, 131, 25, 102, 159, 57},
					"G": {108, 95, 135, 150, 76, 11, 121, 41, 165},
					"H": {90, 24, 175, 156, 60, 118, 143},
					"I": {98, 48, 149, 70, 124, 34, 2},
					"J": {97, 35, 74, 58, 115, 142},
					"K": {164, 176, 46, 61, 16, 120, 88, 3, 140},
					"L": {117, 22, 53, 38, 106, 91},
					"M": {163, 80, 94, 15, 109, 132, 31, 56},
					"N": {68, 161, 13, 40, 99},
					"O": {72, 107, 86, 39, 168, 8, 129, 59},
					"P": {50, 103, 125, 19, 87, 137, 153, 36},
					"Q": {157, 141, 111, 169, 63, 29, 84, 14},
					"R": {130, 96, 21, 45, 144, 116, 73, 6},
					"S": {110, 42, 23, 127, 145, 82, 66},
					"T": {1, 170, 152, 30, 44, 126, 139, 104, 83},
					"U": {119, 27, 51, 10, 147, 93, 172},
					"V": {92, 43, 79, 4, 62, 171, 112, 151, 18},
					"W": {123, 136, 49, 166, 81, 65, 154, 100, 28},
				}
				assert.Equal(t, expected, puzzle.ClueNumbers)
			},
		},
		{
			name:  "cells (partial last row)",
			input: load(t, "xwordinfo-nyt-20200607-partial-last-row.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"T", "H", "E", "R", "E", "", "I", "S", "", "A", "", "L", "A", "W", "", "W", "H", "I", "C", "H", "", "D", "E", "C", "R", "E", "E"},
					{"S", "", "T", "H", "A", "T", "", "T", "W", "O", "", "O", "B", "J", "E", "C", "T", "S", "", "M", "A", "Y", "", "N", "O", "T", ""},
					{"O", "C", "C", "U", "P", "Y", "", "T", "H", "E", "", "S", "A", "M", "E", "", "P", "L", "A", "C", "E", "", "A", "T", "", "T", "H"},
					{"E", "", "S", "A", "M", "E", "", "T", "I", "M", "E", "", "T", "W", "O", "", "P", "E", "O", "P", "L", "E", "", "C", "A", "N", "N"},
					{"O", "T", "", "S", "E", "E", "", "T", "H", "I", "N", "G", "S", "", "F", "R", "O", "M", "", "T", "H", "E", "", "S", "A", "M", "E"},
					{"", "P", "O", "I", "N", "T", "", "O", "F", "", "V", "I", "E", "W", "", "A", "N", "D", "", "T", "H", "E", "", "S", "L", "I", "G"},
					{"H", "T", "E", "S", "T", "", "D", "I", "F", "F", "E", "R", "E", "N", "C", "E", "", "I", "N", "", "A", "N", "G", "L", "E", "", "C"},
					{"H", "A", "N", "G", "E", "S", "", "T", "H", "E", "", "T", "H", "I", "N", "G", "", "S", "E", "E", "N", "", "", "", "", "", ""},
				}
				assert.Equal(t, expected, puzzle.Cells)
			},
		},
		{
			name:  "blocks (partial last row)",
			input: load(t, "xwordinfo-nyt-20200607-partial-last-row.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]bool{
					{false, false, false, false, false, true, false, false, true, false, true, false, false, false, true, false, false, false, false, false, true, false, false, false, false, false, false},
					{false, true, false, false, false, false, true, false, false, false, true, false, false, false, false, false, false, false, true, false, false, false, true, false, false, false, true},
					{false, false, false, false, false, false, true, false, false, false, true, false, false, false, false, true, false, false, false, false, false, true, false, false, true, false, false},
					{false, true, false, false, false, false, true, false, false, false, false, true, false, false, false, true, false, false, false, false, false, false, true, false, false, false, false},
					{false, false, true, false, false, false, true, false, false, false, false, false, false, true, false, false, false, false, true, false, false, false, true, false, false, false, false},
					{true, false, false, false, false, false, true, false, false, true, false, false, false, false, true, false, false, false, true, false, false, false, true, false, false, false, false},
					{false, false, false, false, false, true, false, false, false, false, false, false, false, false, false, false, true, false, false, true, false, false, false, false, false, true, false},
					{false, false, false, false, false, false, true, false, false, false, true, false, false, false, false, false, true, false, false, false, false, true, true, true, true, true, true},
				}
				assert.Equal(t, expected, puzzle.CellBlocks)
			},
		},
		{
			name:  "cell numbers (partial last row)",
			input: load(t, "xwordinfo-nyt-20200607-partial-last-row.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]int{
					{1, 2, 3, 4, 5, 0, 6, 7, 0, 8, 0, 9, 10, 11, 0, 12, 13, 14, 15, 16, 0, 17, 18, 19, 20, 21, 22},
					{23, 0, 24, 25, 26, 27, 0, 28, 29, 30, 0, 31, 32, 33, 34, 35, 36, 37, 0, 38, 39, 40, 0, 41, 42, 43, 0},
					{44, 45, 46, 47, 48, 49, 0, 50, 51, 52, 0, 53, 54, 55, 56, 0, 57, 58, 59, 60, 61, 0, 62, 63, 0, 64, 65},
					{66, 0, 67, 68, 69, 70, 0, 71, 72, 73, 74, 0, 75, 76, 77, 0, 78, 79, 80, 81, 82, 83, 0, 84, 85, 86, 87},
					{88, 89, 0, 90, 91, 92, 0, 93, 94, 95, 96, 97, 98, 0, 99, 100, 101, 102, 0, 103, 104, 105, 0, 106, 107, 108, 109},
					{0, 110, 111, 112, 113, 114, 0, 115, 116, 0, 117, 118, 119, 120, 0, 121, 122, 123, 0, 124, 125, 126, 0, 127, 128, 129, 130},
					{131, 132, 133, 134, 135, 0, 136, 137, 138, 139, 140, 141, 142, 143, 144, 145, 0, 146, 147, 0, 148, 149, 150, 151, 152, 0, 153},
					{154, 155, 156, 157, 158, 159, 0, 160, 161, 162, 0, 163, 164, 165, 166, 167, 0, 168, 169, 170, 171, 0, 0, 0, 0, 0, 0},
				}
				assert.Equal(t, expected, puzzle.CellNumbers)
			},
		},
		{
			name:  "cell clue letters (partial last row)",
			input: load(t, "xwordinfo-nyt-20200607-partial-last-row.json"),
			verify: func(t *testing.T, puzzle *Puzzle) {
				expected := [][]string{
					{"K", "S", "V", "D", "F", "", "B", "N", "", "T", "", "G", "H", "Q", "", "M", "E", "W", "A", "O", "", "C", "D", "Y", "F", "I", "B"},
					{"J", "", "V", "H", "U", "P", "", "E", "T", "R", "", "C", "K", "L", "Q", "I", "J", "A", "", "U", "F", "V", "", "X", "N", "T", ""},
					{"H", "G", "R", "C", "O", "E", "", "I", "J", "Y", "", "U", "P", "K", "D", "", "A", "B", "Q", "F", "X", "", "C", "E", "", "G", "I"},
					{"T", "", "Y", "W", "H", "R", "", "U", "J", "V", "D", "", "M", "Q", "G", "", "X", "E", "L", "I", "K", "S", "", "U", "A", "W", "C"},
					{"O", "B", "", "F", "M", "Y", "", "H", "G", "E", "Q", "T", "I", "", "N", "D", "P", "L", "", "A", "F", "S", "", "B", "V", "U", "Y"},
					{"", "H", "O", "G", "J", "D", "", "I", "N", "", "Q", "T", "E", "X", "", "L", "P", "W", "", "M", "R", "K", "", "Y", "F", "U", "J"},
					{"C", "N", "H", "V", "O", "", "S", "G", "E", "B", "I", "W", "A", "T", "P", "X", "", "D", "R", "", "K", "Y", "H", "L", "Q", "", "E"},
					{"I", "W", "V", "S", "M", "T", "", "P", "X", "C", "", "R", "U", "J", "O", "L", "", "G", "B", "N", "K", "", "", "", "", "", ""},
				}
				assert.Equal(t, expected, puzzle.CellClueLetters)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()

			puzzle, err := ParseXWordInfoPuzzleResponse(test.input)
			require.NoError(t, err)
			test.verify(t, puzzle)
		})
	}
}

func TestParseXWordInfoPuzzleResponse_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "malformed response",
			input: `{true}`,
		},
		{
			name:  "empty response",
			input: ``,
		},
		{
			name:  "empty puzzle",
			input: `{"date":"1/1/2001"}`,
		},
		{
			name:  "malformed published date",
			input: `{"date":"hello world"}`,
		},
		{
			name: "too many clues",
			input: `{
                "date": "1/1/2001",
                "answerKey": "answers go here",
								"clues": [
                  "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", 
                  "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", 
                  "Y", "Z", "extra"
                ],
                "clueData": [
									"1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", 
                  "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1", 
                  "1", "1", "1"
                ]
							}`,
		},
		{
			name: "bad clue data (numbers)",
			input: `{
                "date": "1/1/2001",
                "answerKey": "answers go here",
								"clues": ["A"],
                "clueData": ["1,2,a,3"]
							}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseXWordInfoPuzzleResponse(strings.NewReader(test.input))
			require.Error(t, err)
		})
	}
}

func TestParseXWordInfoAvailableDatesResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  io.ReadCloser
		verify func(t *testing.T, dates []time.Time)
	}{
		{
			name:  "dates",
			input: load(t, "xwordinfo-select-acrostic-20200610.html"),
			verify: func(t *testing.T, dates []time.Time) {
				expected := []time.Time{
					time.Date(1999, time.September, 12, 0, 0, 0, 0, time.UTC),
					time.Date(1999, time.September, 26, 0, 0, 0, 0, time.UTC),
					time.Date(1999, time.October, 10, 0, 0, 0, 0, time.UTC),
					time.Date(1999, time.October, 24, 0, 0, 0, 0, time.UTC),
					time.Date(1999, time.November, 7, 0, 0, 0, 0, time.UTC),
					time.Date(1999, time.November, 21, 0, 0, 0, 0, time.UTC),
					time.Date(1999, time.December, 5, 0, 0, 0, 0, time.UTC),
					time.Date(1999, time.December, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.January, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.February, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.February, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.March, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.March, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.April, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.April, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.May, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.May, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.June, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.June, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.July, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.July, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.July, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.August, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.August, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.September, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.September, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.October, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.October, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.November, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.November, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.December, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.December, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2000, time.December, 31, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.January, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.January, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.February, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.February, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.March, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.March, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.April, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.April, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.May, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.May, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.June, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.June, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.July, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.July, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.July, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.August, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.August, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.September, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.September, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.October, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.October, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.November, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.November, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.November, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.December, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.December, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2001, time.December, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.January, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.January, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.February, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.February, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.March, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.March, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.April, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.April, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.May, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.May, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.June, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.June, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.June, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.July, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.July, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.August, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.August, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.September, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.September, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.October, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.October, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.November, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.November, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.December, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.December, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2002, time.December, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.January, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.January, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.February, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.February, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.March, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.March, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.April, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.April, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.May, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.May, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.June, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.June, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.June, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.July, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.July, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.August, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.August, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.September, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.September, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.October, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.October, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.November, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.November, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.November, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.December, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2003, time.December, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.January, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.January, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.February, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.February, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.March, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.March, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.April, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.April, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.May, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.May, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.May, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.June, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.June, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.July, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.July, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.August, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.August, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.September, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.September, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.October, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.October, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.October, 31, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.November, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.November, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.December, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2004, time.December, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.January, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.January, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.February, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.February, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.March, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.March, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.April, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.April, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.May, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.May, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.May, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.June, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.June, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.July, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.July, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.August, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.August, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.September, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.September, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.October, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.October, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.October, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.November, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.November, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.December, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2005, time.December, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.January, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.January, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.February, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.February, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.March, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.March, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.April, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.April, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.April, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.May, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.May, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.June, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.June, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.July, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.July, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.August, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.August, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.September, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.September, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.October, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.October, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.October, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.November, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.November, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.December, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2006, time.December, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.January, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.January, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.February, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.February, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.March, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.March, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.April, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.April, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.April, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.May, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.May, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.June, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.June, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.July, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.July, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.August, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.August, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.September, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.September, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.September, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.October, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.October, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.November, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.November, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.December, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2007, time.December, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.January, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.January, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.February, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.February, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.March, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.March, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.March, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.April, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.April, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.May, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.May, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.June, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.June, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.July, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.July, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.August, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.August, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.August, 31, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.September, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.September, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.October, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.October, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.November, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.November, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.December, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2008, time.December, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.January, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.January, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.January, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.February, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.March, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.March, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.March, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.April, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.April, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.May, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.May, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.June, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.June, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.July, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.July, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.August, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.August, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.August, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.September, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.September, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.October, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.October, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.November, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.November, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.December, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2009, time.December, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.January, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.January, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.January, 31, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.February, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.February, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.March, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.March, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.April, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.April, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.May, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.May, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.June, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.June, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.July, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.July, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.August, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.August, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.August, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.September, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.September, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.October, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.October, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.November, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.November, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.December, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2010, time.December, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.January, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.January, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.January, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.February, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.March, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.March, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.March, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.April, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.April, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.May, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.May, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.June, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.June, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.July, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.July, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.July, 31, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.August, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.August, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.September, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.September, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.October, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.October, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.November, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.November, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.December, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2011, time.December, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.January, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.January, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.January, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.February, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.February, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.March, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.March, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.April, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.April, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.May, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.May, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.June, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.June, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.July, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.July, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.July, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.August, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.August, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.September, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.September, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.October, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.October, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.November, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.November, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.December, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.December, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2012, time.December, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.January, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.January, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.February, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.February, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.March, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.March, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.April, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.April, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.May, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.May, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.June, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.June, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.June, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.July, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.July, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.August, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.August, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.September, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.September, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.October, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.October, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.November, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.November, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.December, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.December, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2013, time.December, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.January, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.January, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.February, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.February, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.March, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.March, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.April, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.April, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.May, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.May, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.June, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.June, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.June, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.July, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.July, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.August, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.August, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.September, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.September, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.October, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.October, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.November, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.November, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.November, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.December, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2014, time.December, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.January, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.January, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.February, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.February, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.March, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.March, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.April, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.April, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.May, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.May, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.May, 31, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.June, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.June, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.July, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.July, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.August, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.August, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.September, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.September, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.October, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.October, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.November, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.November, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.November, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.December, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2015, time.December, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.January, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.January, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.February, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.February, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.March, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.March, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.April, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.April, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.May, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.May, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.May, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.June, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.June, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.July, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.July, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.August, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.August, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.September, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.September, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.October, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.October, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.October, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.November, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.November, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.December, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2016, time.December, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.January, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.January, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.February, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.February, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.March, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.March, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.April, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.April, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.April, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.May, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.May, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.June, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.June, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.July, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.July, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.August, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.August, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.September, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.September, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.October, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.October, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.October, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.November, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.November, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.December, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2017, time.December, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.January, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.January, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.February, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.February, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.March, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.March, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.April, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.April, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.April, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.May, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.May, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.June, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.June, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.July, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.August, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.August, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.September, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.September, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.September, 30, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.October, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.October, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.November, 11, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.November, 25, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.December, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2018, time.December, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.January, 6, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.January, 20, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.February, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.February, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.March, 3, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.March, 17, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.March, 31, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.April, 14, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.April, 28, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.May, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.May, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.June, 9, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.June, 23, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.July, 7, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.July, 21, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.August, 4, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.August, 18, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.September, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.September, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.September, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.October, 13, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.October, 27, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.November, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.November, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.December, 8, 0, 0, 0, 0, time.UTC),
					time.Date(2019, time.December, 22, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.January, 5, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.January, 19, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.February, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.February, 16, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.March, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.March, 15, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.March, 29, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.April, 12, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.April, 26, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.May, 10, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.May, 24, 0, 0, 0, 0, time.UTC),
					time.Date(2020, time.June, 7, 0, 0, 0, 0, time.UTC),
				}
				assert.ElementsMatch(t, expected, dates)
			},
		},
		{
			name:  "dates are sorted",
			input: load(t, "xwordinfo-select-acrostic-20200610.html"),
			verify: func(t *testing.T, dates []time.Time) {
				assert.True(t, sort.SliceIsSorted(dates, func(i, j int) bool {
					return dates[i].Before(dates[j])
				}))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer test.input.Close()
			dates, err := ParseXWordInfoAvailableDatesResponse(test.input)
			require.NoError(t, err)
			test.verify(t, dates)
		})
	}
}

func TestParseXWordInfoAvailableDatesResponse_Error(t *testing.T) {
	tests := []struct {
		name  string
		input io.Reader
	}{
		{
			name:  "reader returning error",
			input: iotest.TimeoutReader(strings.NewReader("random input")),
		},
		{
			name:  "anchor with no href",
			input: strings.NewReader(`<a class="dtlink"/>`),
		},
		{
			name:  "href missing prefix",
			input: strings.NewReader(`<a class="dtlink" href="/MissingPrefix?date=1/1/2020"/>`),
		},
		{
			name:  "error takes precedence",
			input: strings.NewReader(`<a class="dtlink"/><a class="dtlink" href="/Acrostic?date=1/1/2020"/>`),
		},
		{
			name:  "no links found",
			input: strings.NewReader(``),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseXWordInfoAvailableDatesResponse(test.input)
			require.Error(t, err)
		})
	}
}
