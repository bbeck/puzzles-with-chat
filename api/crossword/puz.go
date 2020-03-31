package crossword

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"io"
	"strings"
)

//
// PuzFile represents the binary file format of a .puz file that contains a
// crossword puzzle.
//
// In short it contains a fixed size header that outlines the various checksums
// and dimensions of the puzzle followed by the variable length portions of the
// puzzle such as the solution and clues.  Lastly the file can contain a set of
// extension sections to the file format to represent things like rebus squares,
// circled squares, etc.
//
// Of note is that the file format itself doesn't actually contain clue numbers,
// they must be derived from the structure of the black cells within the
// puzzle's grid.
//
// Details on the file format can be found at:
//  	https://code.google.com/archive/p/puz/wikis/FileFormat.wiki
//
// Implementations of parsers can be found at:
//    https://github.com/alexdej/puzpy
//    https://github.com/kobolabs/puz
//
type PuzFile struct {
	Header struct {
		GlobalChecksum      uint16
		MagicNumber         [12]byte // ACROSS&DOWN\000 (null terminated)
		HeaderChecksum      uint16
		MaskedChecksum      [8]byte
		Version             [4]byte // e.g. 1.2\0 (null terminated)
		_                   uint16
		UnscrambledChecksum uint16
		_                   [12]byte
		Width               uint8
		Height              uint8
		NumClues            uint16
		UnknownBitmask      uint16
		ScrambledTag        uint16
	}

	Solution   []byte
	Cells      []byte
	Title      []byte
	Author     []byte
	Copyright  []byte
	Clues      [][]byte
	Notes      []byte
	Extensions map[string]*PuzFileExtension
}

var MagicNumber = []byte("ACROSS&DOWN\000")

// LoadFromEncodedPuzFile will base64 decode the input and then attempt to load
// the resulting binary as a .puz file into a Puzzle object.
func LoadFromEncodedPuzFile(encoded string) (*Puzzle, error) {
	if testCachedPuzzle != nil {
		return testCachedPuzzle, nil
	}

	if testCachedError != nil {
		return nil, testCachedError
	}

	bs, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		err = fmt.Errorf("unable to base64 decode .puz bytes: %+v", err)
		return nil, err
	}

	return LoadPuzFile(bytes.NewReader(bs))
}

// LoadPuzFile parses a binary .puz file into a Puzzle object.
func LoadPuzFile(in io.Reader) (*Puzzle, error) {
	var err error

	// Some files have extra text before the header that isn't represented
	// anywhere in the .puz file format.  Seek past that information.
	in, err = seekToHeader(in)
	if err != nil {
		return nil, err
	}

	var f PuzFile
	if err := binary.Read(in, binary.LittleEndian, &f.Header); err != nil {
		return nil, err
	}

	f.Solution = make([]byte, int(f.Header.Width)*int(f.Header.Height))
	if err := binary.Read(in, binary.LittleEndian, &f.Solution); err != nil {
		return nil, err
	}

	f.Cells = make([]byte, int(f.Header.Width)*int(f.Header.Height))
	if err := binary.Read(in, binary.LittleEndian, &f.Cells); err != nil {
		return nil, err
	}

	if f.Title, err = ReadUntil(in, 0); err != nil {
		return nil, err
	}

	if f.Author, err = ReadUntil(in, 0); err != nil {
		return nil, err
	}

	if f.Copyright, err = ReadUntil(in, 0); err != nil {
		return nil, err
	}

	f.Clues = make([][]byte, f.Header.NumClues)
	for i := uint16(0); i < f.Header.NumClues; i++ {
		f.Clues[i], err = ReadUntil(in, 0)
		if err != nil {
			return nil, err
		}
	}

	if f.Notes, err = ReadUntil(in, 0); err != nil {
		return nil, err
	}

	f.Extensions = make(map[string]*PuzFileExtension)
	for {
		var ext PuzFileExtension
		if err = binary.Read(in, binary.LittleEndian, &ext.Header); err != nil {
			// Since extensions are the last thing in the file at some point in this
			// loop we'll end up trying to load another extension when we're at the
			// end of the file.  This is normal and we can ignore the EOF error that
			// is returned.
			if err == io.EOF {
				break
			}

			// Any other error while reading is a problem.
			return nil, err
		}

		if ext.Data, err = ReadLength(in, ext.Header.Length); err != nil {
			return nil, err
		}

		// Extensions have a trailing null byte.
		if _, err = ReadLength(in, 1); err != nil {
			return nil, err
		}

		f.Extensions[string(ext.Header.Code[:])] = &ext
	}

	// At this point the entire .puz file has been read.  Now verify that
	// everything has been interpreted correctly by making sure all of the
	// checksums are correct.
	if f.Header.HeaderChecksum != f.HeaderChecksum() {
		return nil, errors.New("header checksums do not match")
	}

	if f.Header.GlobalChecksum != f.GlobalChecksum() {
		return nil, errors.New("global checksums do not match")
	}

	if f.Header.MaskedChecksum != f.MaskedChecksum() {
		return nil, errors.New("masked checksums do not match")
	}

	for _, e := range f.Extensions {
		if e.Header.Checksum != e.Checksum() {
			err = fmt.Errorf("extension %s checksums do not match", e.Header.Code)
			return nil, err
		}
	}

	// Diagram-less puzzles often use : characters to indicate that a square is a
	// block that shouldn't be rendered to the user.  We currently don't support
	// rendering these diagram-less puzzles properly so for the time being just
	// convert it into a normal puzzle.  We wait until after verifying checksums
	// to do this because some of the checksum calculations require the :
	// characters in their computations.
	f.Solution = bytes.ReplaceAll(f.Solution, []byte(":"), []byte("."))

	// If the puzzle is scrambled then unscramble it now.  We do this after
	// checksum verification since some of the checksums use the solution field
	// in its scrambled form in their computations and unscrambling will change
	// the value of that field.
	if f.Header.ScrambledTag == 0x0004 {
		var success bool
		for key := 1000; key <= 9999; key++ {
			if f.Unscramble(key) {
				success = true
				break
			}
		}

		if !success {
			return nil, errors.New("unable to unscramble solution")
		}
	}

	// Now that the PuzFile has been verified and unscrambled we can convert it
	// to a Puzzle object.
	puzzle, err := f.Convert()
	if err != nil {
		return nil, err
	}

	return puzzle, nil
}

// SeekToHeader returns an io.Reader that is pointing at the beginning of the
// header within the passed in reader.  The reader returned is not the same as
// the inputted reader.
func seekToHeader(in io.Reader) (io.Reader, error) {
	// Start by finding the location of the ACROSS&DOWN\0 magic number in the
	// header.  This allows us to read files where there to be some data before
	// the contents of the .puz file header.  While data before the header doesn't
	// happen frequently, it does show up periodically.
	//
	// To facilitate this searching we'll use a buffered reader so that we can
	// look ahead at a chunk of bytes and then reset back to the correct point
	// where the header actually begins.
	reader := bufio.NewReader(in)

	// Extract all of the data in the reader's buffer so that we can more easily
	// search it for the magic number.  An EOF error can happen if the entire
	// contents of the reader fits in the buffer, which is fine.
	buffer, err := reader.Peek(reader.Size())
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Search for the magic number in the buffer.
	index := bytes.Index(buffer, MagicNumber)
	if index == -1 {
		return nil, errors.New("unable to find magic number in reader prefix")
	}

	// We found the magic number, advance the reader to 2 bytes before it's index.
	// We move an extra 2 bytes before because the global check is a 2 byte field
	// that appears before the magic number in the header.
	if _, err := reader.Discard(index - 2); err != nil {
		return nil, err
	}

	// At this point our buffered reader is pointing right at the beginning of
	// the header.
	return reader, nil
}

func (f *PuzFile) HeaderChecksum() uint16 {
	var crc CRC
	crc = crc.Write8(f.Header.Width)
	crc = crc.Write8(f.Header.Height)
	crc = crc.Write16(f.Header.NumClues)
	crc = crc.Write16(f.Header.UnknownBitmask)
	crc = crc.Write16(f.Header.ScrambledTag)
	return uint16(crc)
}

func (f *PuzFile) GlobalChecksum() uint16 {
	var crc = CRC(f.Header.HeaderChecksum)
	crc = crc.Write(f.Solution)
	crc = crc.Write(f.Cells)
	crc = f.StringsChecksum(crc)
	return uint16(crc)
}

func (f *PuzFile) MaskedChecksum() [8]byte {
	header := f.Header.HeaderChecksum
	solution := CRC(0).Write(f.Solution)
	grid := CRC(0).Write(f.Cells)
	part := f.StringsChecksum(0)

	return [8]byte{
		'I' ^ byte(header&0x00FF),
		'C' ^ byte(solution&0x00FF),
		'H' ^ byte(grid&0x00FF),
		'E' ^ byte(part&0x00FF),
		'A' ^ byte(header>>8),
		'T' ^ byte(solution>>8),
		'E' ^ byte(grid>>8),
		'D' ^ byte(part>>8),
	}
}

func (f *PuzFile) StringsChecksum(crc CRC) CRC {
	if len(f.Title) > 0 {
		crc = crc.Write(f.Title)
		crc = crc.Write8(0)
	}

	if len(f.Author) > 0 {
		crc = crc.Write(f.Author)
		crc = crc.Write8(0)
	}

	if len(f.Copyright) > 0 {
		crc = crc.Write(f.Copyright)
		crc = crc.Write8(0)
	}

	for _, clue := range f.Clues {
		if len(clue) > 0 {
			crc = crc.Write(clue)
		}
	}

	// Notes are only in the strings checksum starting in version 1.3 of the
	// file format.  We'll only include it if we know for sure that the version
	// is 1.3 or later.
	major, minor := f.Version()
	if len(f.Notes) > 0 && (major > 1 || (major == 1 && minor >= 3)) {
		crc = crc.Write(f.Notes)
		crc = crc.Write8(0)
	}

	return crc
}

// Version returns the version number of the puzzle file.
func (f *PuzFile) Version() (int, int) {
	var major, minor int
	if _, err := fmt.Sscanf(string(f.Header.Version[:]), "%d.%d", &major, &minor); err != nil {
		return 0, 0
	}

	return major, minor
}

// Unscramble attempts to undo a scramble operation with the provided key.  If
// the operation succeeds then the solution will be updated and true will be
// returned.
func (f *PuzFile) Unscramble(key int) bool {
	// Perform a matrix transpose operation on the bytes of the input.
	transpose := func(in []byte, width, height int) []byte {
		N := len(in)

		out := make([]byte, N)
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				out[x*height+y] = in[y*width+x]
			}
		}

		return out
	}

	// Undo an interleaving of the first half of the input with the second.
	unshuffle := func(in []byte) []byte {
		N := len(in)

		out := make([]byte, N)
		for i := 0; i < N/2; i++ {
			out[i] = in[2*i+1]
		}
		for i := N / 2; i < N; i++ {
			out[i] = in[2*(i-N/2)]
		}

		return out
	}

	// Shift each character in the input left by the number of digits in the key.
	// Each character uses a different key digit reusing key digits when they are
	// exhausted.
	unshift := func(in []byte, key [4]byte) []byte {
		K := len(key)

		out := make([]byte, len(in))
		for i, c := range in {
			out[i] = UnshiftTables[key[i%K]][c]
		}

		return out
	}

	// Rotate the slice to the right by a number of digits.  Things that rotate
	// off the end of the array will come back in at the beginning of it.
	rotate := func(in []byte, k int) []byte {
		N := len(in)
		return append(in[N-k:], in[:N-k]...)
	}

	unscramble := func(in []byte, key [4]byte) []byte {
		out := in
		for i := len(key) - 1; i >= 0; i-- {
			k := int(key[i])
			out = unshift(rotate(unshuffle(out), k), key)
		}

		return out
	}

	// Return a copy of the input without any '.' bytes.
	filter := func(in []byte) []byte {
		out := make([]byte, 0, len(in))
		for _, b := range in {
			if b != '.' {
				out = append(out, b)
			}
		}

		return out
	}

	// Take two board representations, one with '.' bytes and one without.
	// Insert the '.' bytes into the correct positions in the board without them.
	restore := func(with, without []byte) []byte {
		var out = make([]byte, len(with))
		for i, j := 0, 0; i < len(with); i++ {
			if with[i] == '.' {
				out[i] = '.'
				continue
			}

			out[i] = without[j]
			j++
		}

		return out
	}

	// Compute a CRC of the inputted board representation.
	checksum := func(in []byte, width, height int) uint16 {
		var crc CRC
		for _, b := range transpose(in, width, height) {
			if b != '.' {
				crc = crc.Write8(b)
			}
		}
		return uint16(crc)
	}

	// Determine the 4 digits of the key.
	digits := [4]byte{
		byte((key / 1000) % 10),
		byte((key / 100) % 10),
		byte((key / 10) % 10),
		byte((key / 1) % 10),
	}

	// Run the unscramble algorithm.
	width := int(f.Header.Width)
	height := int(f.Header.Height)

	transposed := transpose(f.Solution, width, height)
	unscrambled := unscramble(filter(transposed), digits)
	restored := restore(transposed, unscrambled)
	solution := transpose(restored, height, width)

	// Check to see if this was the right key by verifying the checksum against
	// what the header says it should be when unscrambled.
	if f.Header.UnscrambledChecksum == checksum(solution, width, height) {
		f.Solution = solution
		f.Header.UnscrambledChecksum = 0
		f.Header.ScrambledTag = 0x0000
		return true
	}

	return false
}

// These tables help to perform unshift operations quickly by precomputing the
// 10 possible unshift operations we might have to do.
var UnshiftTables = [10]map[byte]byte{
	// 0
	{
		'A': 'A', 'B': 'B', 'C': 'C', 'D': 'D', 'E': 'E', 'F': 'F', 'G': 'G',
		'H': 'H', 'I': 'I', 'J': 'J', 'K': 'K', 'L': 'L', 'M': 'M', 'N': 'N',
		'O': 'O', 'P': 'P', 'Q': 'Q', 'R': 'R', 'S': 'S', 'T': 'T', 'U': 'U',
		'V': 'V', 'W': 'W', 'X': 'X', 'Y': 'Y', 'Z': 'Z',
	},
	// 1
	{
		'A': 'Z', 'B': 'A', 'C': 'B', 'D': 'C', 'E': 'D', 'F': 'E', 'G': 'F',
		'H': 'G', 'I': 'H', 'J': 'I', 'K': 'J', 'L': 'K', 'M': 'L', 'N': 'M',
		'O': 'N', 'P': 'O', 'Q': 'P', 'R': 'Q', 'S': 'R', 'T': 'S', 'U': 'T',
		'V': 'U', 'W': 'V', 'X': 'W', 'Y': 'X', 'Z': 'Y',
	},
	// 2
	{
		'A': 'Y', 'B': 'Z', 'C': 'A', 'D': 'B', 'E': 'C', 'F': 'D', 'G': 'E',
		'H': 'F', 'I': 'G', 'J': 'H', 'K': 'I', 'L': 'J', 'M': 'K', 'N': 'L',
		'O': 'M', 'P': 'N', 'Q': 'O', 'R': 'P', 'S': 'Q', 'T': 'R', 'U': 'S',
		'V': 'T', 'W': 'U', 'X': 'V', 'Y': 'W', 'Z': 'X',
	},
	// 3
	{
		'A': 'X', 'B': 'Y', 'C': 'Z', 'D': 'A', 'E': 'B', 'F': 'C', 'G': 'D',
		'H': 'E', 'I': 'F', 'J': 'G', 'K': 'H', 'L': 'I', 'M': 'J', 'N': 'K',
		'O': 'L', 'P': 'M', 'Q': 'N', 'R': 'O', 'S': 'P', 'T': 'Q', 'U': 'R',
		'V': 'S', 'W': 'T', 'X': 'U', 'Y': 'V', 'Z': 'W',
	},
	// 4
	{
		'A': 'W', 'B': 'X', 'C': 'Y', 'D': 'Z', 'E': 'A', 'F': 'B', 'G': 'C',
		'H': 'D', 'I': 'E', 'J': 'F', 'K': 'G', 'L': 'H', 'M': 'I', 'N': 'J',
		'O': 'K', 'P': 'L', 'Q': 'M', 'R': 'N', 'S': 'O', 'T': 'P', 'U': 'Q',
		'V': 'R', 'W': 'S', 'X': 'T', 'Y': 'U', 'Z': 'V',
	},
	// 5
	{
		'A': 'V', 'B': 'W', 'C': 'X', 'D': 'Y', 'E': 'Z', 'F': 'A', 'G': 'B',
		'H': 'C', 'I': 'D', 'J': 'E', 'K': 'F', 'L': 'G', 'M': 'H', 'N': 'I',
		'O': 'J', 'P': 'K', 'Q': 'L', 'R': 'M', 'S': 'N', 'T': 'O', 'U': 'P',
		'V': 'Q', 'W': 'R', 'X': 'S', 'Y': 'T', 'Z': 'U',
	},
	// 6
	{
		'A': 'U', 'B': 'V', 'C': 'W', 'D': 'X', 'E': 'Y', 'F': 'Z', 'G': 'A',
		'H': 'B', 'I': 'C', 'J': 'D', 'K': 'E', 'L': 'F', 'M': 'G', 'N': 'H',
		'O': 'I', 'P': 'J', 'Q': 'K', 'R': 'L', 'S': 'M', 'T': 'N', 'U': 'O',
		'V': 'P', 'W': 'Q', 'X': 'R', 'Y': 'S', 'Z': 'T',
	},
	// 7
	{
		'A': 'T', 'B': 'U', 'C': 'V', 'D': 'W', 'E': 'X', 'F': 'Y', 'G': 'Z',
		'H': 'A', 'I': 'B', 'J': 'C', 'K': 'D', 'L': 'E', 'M': 'F', 'N': 'G',
		'O': 'H', 'P': 'I', 'Q': 'J', 'R': 'K', 'S': 'L', 'T': 'M', 'U': 'N',
		'V': 'O', 'W': 'P', 'X': 'Q', 'Y': 'R', 'Z': 'S',
	},
	// 8
	{
		'A': 'S', 'B': 'T', 'C': 'U', 'D': 'V', 'E': 'W', 'F': 'X', 'G': 'Y',
		'H': 'Z', 'I': 'A', 'J': 'B', 'K': 'C', 'L': 'D', 'M': 'E', 'N': 'F',
		'O': 'G', 'P': 'H', 'Q': 'I', 'R': 'J', 'S': 'K', 'T': 'L', 'U': 'M',
		'V': 'N', 'W': 'O', 'X': 'P', 'Y': 'Q', 'Z': 'R',
	},
	// 9
	{
		'A': 'R', 'B': 'S', 'C': 'T', 'D': 'U', 'E': 'V', 'F': 'W', 'G': 'X',
		'H': 'Y', 'I': 'Z', 'J': 'A', 'K': 'B', 'L': 'C', 'M': 'D', 'N': 'E',
		'O': 'F', 'P': 'G', 'Q': 'H', 'R': 'I', 'S': 'J', 'T': 'K', 'U': 'L',
		'V': 'M', 'W': 'N', 'X': 'O', 'Y': 'P', 'Z': 'Q',
	},
}

// Convert takes the in-memory representation of a .puz file and converts it
// into a Puzzle object.
func (f *PuzFile) Convert() (*Puzzle, error) {
	var errs []error

	// The spec claims that all strings are ISO-8859-1, but this seems to not be
	// entirely true.  Some .puz files in the wild
	// (puzpy-nyt-20080224-diagramless.puz) include characters like a right quote
	// which is defined in Windows-1252 which is a superset of ISO-8859-1.  If an
	// error in decoding happens then the error will be added to the errs slice.
	decode := func(bs []byte) string {
		decoded, err := charmap.Windows1252.NewDecoder().Bytes(bs)
		if err != nil {
			errs = append(errs, err)
		}

		return string(decoded)
	}

	var puzzle Puzzle
	puzzle.Rows = int(f.Header.Height)
	puzzle.Cols = int(f.Header.Width)
	puzzle.Title = decode(f.Title)

	puzzle.Author = strings.TrimSpace(decode(f.Author))
	if strings.HasPrefix(puzzle.Author, "by ") || strings.HasPrefix(puzzle.Author, "By ") {
		puzzle.Author = puzzle.Author[3:]
	}

	puzzle.Notes = strings.TrimSpace(decode(f.Notes))

	// Parse the entries of the rebus table if one exists.
	var rebusTable = make(map[int]string)
	if extension := f.Extensions["RTBL"]; extension != nil {
		for _, entry := range bytes.Split(extension.Data, []byte{';'}) {
			// Since each entry ends with a ; we'll have an empty entry at the end
			if len(entry) == 0 {
				break
			}

			var key int
			var value string
			if _, err := fmt.Sscanf(string(entry), "%2d:%s", &key, &value); err != nil {
				err = fmt.Errorf("unable to parse RTBL entry (%s): %+v\n", entry, err)
				return nil, err
			}

			rebusTable[key] = value
		}
	}

	// Determine if there are rebus cells.
	var rebusCells [][]int
	if extension := f.Extensions["GRBS"]; extension != nil {
		for y := 0; y < puzzle.Rows; y++ {
			rebusCells = append(rebusCells, make([]int, puzzle.Cols))
			for x := 0; x < puzzle.Cols; x++ {
				rebusCells[y][x] = int(extension.Data[y*puzzle.Cols+x])
			}
		}
	}

	// Determine the value for each cell and whether or not it is a block.
	for y := 0; y < puzzle.Rows; y++ {
		puzzle.Cells = append(puzzle.Cells, make([]string, puzzle.Cols))
		puzzle.CellBlocks = append(puzzle.CellBlocks, make([]bool, puzzle.Cols))
		for x := 0; x < puzzle.Cols; x++ {
			var cell string
			if rebusCells != nil && rebusCells[y][x] != 0 {
				cell = rebusTable[rebusCells[y][x]-1]
			} else {
				cell = string(f.Solution[y*puzzle.Cols+x])
			}

			if cell != "." {
				puzzle.Cells[y][x] = cell
			} else {
				puzzle.CellBlocks[y][x] = true
			}
		}
	}

	// Assign the clue numbers.
	puzzle.CluesAcross = make(map[int]string)
	puzzle.CluesDown = make(map[int]string)

	var nextClueNumber = 1 // The next clue number we'll assign
	var nextClueIndex = 0  // The index of the next clue we'll consume
	for y := 0; y < puzzle.Rows; y++ {
		puzzle.CellClueNumbers = append(puzzle.CellClueNumbers, make([]int, puzzle.Cols))

		for x := 0; x < puzzle.Cols; x++ {
			// If this cell is a block there can't be a number.
			if puzzle.CellBlocks[y][x] {
				continue
			}

			// We need an across number if left of us is a block and right isn't
			isLeftABlock := x == 0 || puzzle.CellBlocks[y][x-1]
			isRightABlock := x >= puzzle.Cols-1 || puzzle.CellBlocks[y][x+1]
			if isLeftABlock && !isRightABlock {
				if puzzle.CellClueNumbers[y][x] == 0 {
					puzzle.CellClueNumbers[y][x] = nextClueNumber
					nextClueNumber++
				}

				puzzle.CluesAcross[puzzle.CellClueNumbers[y][x]] = decode(f.Clues[nextClueIndex])
				nextClueIndex++
			}

			// We need a down number if above us is a block and below us isn't.
			isUpABlock := y == 0 || puzzle.CellBlocks[y-1][x]
			isDownABlock := y >= puzzle.Rows-1 || puzzle.CellBlocks[y+1][x]
			if isUpABlock && !isDownABlock {
				if puzzle.CellClueNumbers[y][x] == 0 {
					puzzle.CellClueNumbers[y][x] = nextClueNumber
					nextClueNumber++
				}

				puzzle.CluesDown[puzzle.CellClueNumbers[y][x]] = decode(f.Clues[nextClueIndex])
				nextClueIndex++
			}
		}
	}

	// Determine if there are circles in any cells.
	for y := 0; y < puzzle.Rows; y++ {
		puzzle.CellCircles = append(puzzle.CellCircles, make([]bool, puzzle.Cols))

		for x := 0; x < puzzle.Cols; x++ {
			if extension := f.Extensions["GEXT"]; extension != nil {
				puzzle.CellCircles[y][x] = extension.Data[y*puzzle.Cols+x]&0x80 != 0
			}
		}
	}

	// Check if an error occurred anywhere.
	if errs != nil {
		var err = errors.New("an error occurred while converting")
		for i := 0; i < len(errs); i++ {
			err = fmt.Errorf("%v: %w", err, errs[i])
		}

		return nil, err
	}

	return &puzzle, nil
}

// PuzFileExtension represents the contents of an extension within a .puz file.
// An extension contains a fixed size header followed by variable length data
// that is interpreted according to which type of extension it is.
type PuzFileExtension struct {
	Header struct {
		Code     [4]byte
		Length   uint16
		Checksum uint16
	}

	Data []byte
}

func (e *PuzFileExtension) Checksum() uint16 {
	return uint16(CRC(0).Write(e.Data))
}

func ReadUntil(in io.Reader, delimiter byte) ([]byte, error) {
	var bs []byte

	c := make([]byte, 1)
	for {
		if _, err := in.Read(c); err != nil {
			return nil, err
		}

		if c[0] == delimiter {
			break
		}

		bs = append(bs, c[0])
	}

	return bs, nil
}

func ReadLength(in io.Reader, length uint16) ([]byte, error) {
	var bs []byte

	c := make([]byte, 1)
	for i := uint16(0); i < length; i++ {
		if _, err := in.Read(c); err != nil {
			return nil, err
		}

		bs = append(bs, c[0])
	}

	return bs, nil
}

type CRC uint16

func (crc CRC) Write8(value uint8) CRC {
	return crc.Write([]byte{value})
}

func (crc CRC) Write16(value uint16) CRC {
	return crc.Write([]byte{byte(value & 0x00FF), byte(value >> 8)})
}

func (crc CRC) Write(value []byte) CRC {
	for _, b := range value {
		if crc&0x0001 == 0x0001 {
			crc = (crc >> 1) | 0x8000
		} else {
			crc = crc >> 1
		}

		crc += CRC(b)
	}

	return crc
}
