import attr
import calendar
import dateparser
import datetime
import html
import json
import requests
import typing


@attr.s(frozen=True)
class Puzzle(object):
    r"""Puzzle represents a crossword puzzle.

    The puzzle is comprised of a grid which has dimensions rows x cols and
    demonstrates which cells of the puzzle are available for placing letters
    into and which are not.  Additionally the puzzle contains a set of clues
    organized by number and whether or not they fill in cells going across or
    down.  Lastly a puzzle has various bits of interesting metdata such as the
    publication that the puzzle is from, the date that it was published, and
    the author(s).

    Attributes
    ----------
    rows : int
        The number of rows in the puzzle.

    cols : int
        The number of colums in the puzzle.

    title : str
        The title of the puzzle.

    publisher : str
        The name of the publisher that published the puzzle.

    published : datetime.date
        The date that the puzzle was published.

    author : str
        The name of the author(s) of the puzzle.

    cells : List[List[str]]
        The cells of the puzzle as a 2D list, entries are the letter (or
        letters in the case of a rebus) that belong in the cell.  If a cell
        cannot be inputted into then it will contain `None`.  The lists are
        first indexed by the `row` coordinate of the cell and then by the `col`
        coordinate of the cell.

    cell_clue_numbers : List[List[int]]
        The clue numbers for all of the cells in the puzzle as a 2D list.
        Cells that cannot be inputted into or that don't have a clue number
        will contain a `0` entry.  Like `cells` the 2D list is first indexed
        by the `row` coordinate of the cell and then by the `col` coordinate.

    cell_circles : List[List[bool]]
        Whether or not a cell contains a circle for all of the cells in the
        puzzle as a 2d list.  Cells that should have a circle rendered in them
        appear as `True` and those that shouldn't have a circle appear as
        `False`.  Like `cells` the 2D list is first indexed by the `row`
        coordinate of the cell and then by the `col` coordinate.

    across_clues : Dict[int, str]
        The clues for the across answers indexed by the clue number.

    down_clues : Dict[int, str]
        The clues for the down answers indexed by the clue number.
    """
    rows = attr.ib(type=int)
    cols = attr.ib(type=int)
    title = attr.ib(type=str)
    publisher = attr.ib(type=str)
    published = attr.ib(type=datetime.date)
    author = attr.ib(type=str)
    cells = attr.ib(type=typing.List[typing.List[str]])
    cell_clue_numbers = attr.ib(type=typing.List[typing.List[int]])
    cell_circles = attr.ib(type=typing.List[typing.List[bool]])
    across_clues = attr.ib(type=typing.Dict[int, str])
    down_clues = attr.ib(type=typing.Dict[int, str])

    def to_json(self):
        r"""Converts a Puzzle instance into a JSON string.

        The majority of the fields of the puzzle type are suitable for direct
        conversion to JSON.  For the ones that are not suitable for direct
        conversion to JSON they are changed to be logical equivalents (for
        example dates are changed to ISO8601 date strings).

        Returns
        -------
        str
            The JSON representation of the current puzzle instance.
        """
        d = attr.asdict(self)

        # JSON doesn't support datetimes, so convert to an ISO8601 date string.
        d["published"] = d["published"].isoformat()

        return json.dumps(d)

    @classmethod
    def from_json(cls, s):
        r"""Converts a JSON string to a new Puzzle instance.

        Similar to how `to_json` will encode any fields that are not suitable
        for direction to conversion to JSON this method will undo any such
        encodings (for example dates are changed back to python datetime.date
        instances).

        Parameters
        ----------
        s : str
            The JSON string representation of a Puzzle.

        Returns
        -------
        Puzzle
            The Puzzle instance corresponding to the inputted JSON string.
        """
        d = json.loads(s)

        # Parse the ISO8601 date string into a datetime.date.
        d["published"] = datetime.date.fromisoformat(d["published"])

        return cls(**d)


def load_puzzle(date, publisher="NYT"):
    r"""Loads a crossword puzzle from a publisher on a specific date.

    By default puzzles from The New York Times are loaded.

    Parameters
    ----------
    date : str|datetime.date
        The publication date of the puzzle to load.

    publisher : str
        The publisher whose puzzle to load.

    Returns
    -------
    Puzzle
        The loaded puzzle object or `None` if a puzzle couldn't be loaded.
    """
    if isinstance(date, str):
        date = dateparser.parse(date).date()

    if publisher == "NYT":
        return load_nyt_puzzle(date)


def load_nyt_puzzle(date):
    r"""Loads a crossword puzzle from the New York Times for a particular date.

    This method uses the xwordinfo.com JSON API to load a New York Times
    crossword puzzle.  While organized slightly differently from the XPF API
    the returned data is mostly the same.  Documentation for the JSON API can
    be found here: https://www.xwordinfo.com/JSON/

    If the puzzle cannot be loaded for some reason, `None` is returned.
    """

    # These headers are required in order to get the server to send a non-empty
    # response back.
    headers = {
        # Referer intentionally misspelled per the HTTP spec.
        "Referer": "https://www.xwordinfo.com/JSON",
    }

    url = f"https://www.xwordinfo.com/JSON/Data.aspx?date={date}"
    response = requests.get(url, headers=headers)

    if 400 <= response.status_code < 600:
        # 4xx and 5xx responses are client and server errors respectively.
        return None

    try:
        data = response.json()
    except:
        return None

    rows = data["size"]["rows"]
    cols = data["size"]["cols"]
    title = data["title"]
    publisher = data["publisher"]
    author = data["author"]

    cells = [[None for col in range(cols)] for row in range(rows)]
    for row in range(rows):
        for col in range(cols):
            c = data["grid"][row * cols + col]
            cells[row][col] = c if c != "." else None

    cell_clue_numbers = [[0 for col in range(cols)] for row in range(rows)]
    for row in range(rows):
        for col in range(cols):
            num = int(data["gridnums"][row * cols + col])
            cell_clue_numbers[row][col] = num

    # Not every puzzle has circles, so make sure we check first if they're
    # present before trying to traverse them.
    cell_circles = [[False for col in range(cols)] for row in range(rows)]
    if data.get("circles"):
        for row in range(rows):
            for col in range(cols):
                cell_circles[row][col] = data["circles"][row * cols + col] == 1

    across_clues = {}
    for clue in data["clues"]["across"]:
        num, clue = parse_nyt_clue(clue)
        across_clues[num] = clue

    down_clues = {}
    for clue in data["clues"]["down"]:
        num, clue = parse_nyt_clue(clue)
        down_clues[num] = clue

    return Puzzle(
        rows=rows,
        cols=cols,
        title=title,
        published=date,
        publisher=publisher,
        author=author,
        cells=cells,
        cell_clue_numbers=cell_clue_numbers,
        cell_circles=cell_circles,
        across_clues=across_clues,
        down_clues=down_clues,
    )


def parse_nyt_clue(s):
    r"Parses a New York Times clue into its number and the clue text."

    # Clues look like the following:
    #   1. 4.0 is a great one
    #   13. &quot;Look out!&quot;
    #   67. ___ raving mad
    #
    # Because of this we need to make sure we split only after the first
    # decimal point as there may be other decimal points in the clue.  Also
    # we should unescape HTML characters that may be present as well.
    n, clue = s.split(". ", 1)
    clue = html.unescape(clue)

    return int(n), clue.strip()


def get_answer_cells(puzzle, num, direction):
    r"""Determines the coordinates of the answer cells for a given clue.

    Parameters
    ----------
    puzzle : Puzzle
        The puzzle instance that the answer cells should be determined from.

    num : int
        The number of the clue.

    direction : str
        The direction of the clue.  Must be either `a` or `d`.

    Returns
    -------
    List[typing.Tuple[int, int]]|None
        The coordinates (x, y) of the cells that the answer go into.  If the
        provided clue is not valid then None is returned.
    """
    # First we'll find out which cell the numbered answer begins in.  This is
    # the same regardless of whether we're looking for across or down answers.
    current = None
    for y in range(puzzle.rows):
        for x in range(puzzle.cols):
            if puzzle.cell_clue_numbers[y][x] == num:
                current = (x, y)
    if current is None:
        return None

    # Now that we know the starting cell, let's traverse in the correct
    # direction until we reach either a cell that can't be inputted into or
    # the edge of the puzzle.
    dx, dy = (1, 0) if direction == "a" else (0, 1)

    coordinates = []
    while True:
        x, y = current

        # We're done if we move outside of the puzzle...
        if x >= puzzle.cols or y >= puzzle.rows:
            break

        # ... or have hit a cell that can't be inputted into.
        if puzzle.cells[y][x] is None:
            break

        coordinates.append(current)
        current = (x + dx, y + dy)

    return coordinates
