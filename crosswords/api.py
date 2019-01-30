r"""
This module acts as an API for the xwordinfo.com REST service.  Using it
allows the retrieval of a crossword puzzle.
"""
import calendar
import dateparser
import html
import requests
from . import Puzzle


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

    url = f"https://www.xwordinfo.com/JSON/Data.aspx?format=text&date={date}"
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
            cell_clue_numbers[row][col] = data["gridnums"][row * cols + col]

    # Not every puzzle has circles, so make sure we check first if they're
    # present before trying to traverse them.
    cell_circles = [[False for col in range(cols)] for row in range(rows)]
    if data.get("circles"):
        for row in range(rows):
            for col in range(cols):
                cell_circles[row][col] = data["circles"][row * cols + col] == 1

    across_clues = {}
    for clue in data["clues"]["across"]:
        num, clue = _parse_nyt_clue(clue)
        across_clues[num] = clue

    down_clues = {}
    for clue in data["clues"]["down"]:
        num, clue = _parse_nyt_clue(clue)
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


def _parse_nyt_clue(s):
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
