import attr
import datetime
import html
import json
import puz
import typing


@attr.s(frozen=True)
class Puzzle(object):
    r"""Puzzle represents a crossword puzzle in the GoLang app format."""
    rows = attr.ib(type=int)
    cols = attr.ib(type=int)
    title = attr.ib(type=str)
    publisher = attr.ib(type=typing.Optional[str])
    published = attr.ib(type=typing.Optional[datetime.date])
    author = attr.ib(type=str)
    cells = attr.ib(type=typing.List[typing.List[str]])
    cell_blocks = attr.ib(type=typing.List[typing.List[bool]])
    cell_clue_numbers = attr.ib(type=typing.List[typing.List[int]])
    cell_circles = attr.ib(type=typing.List[typing.List[bool]])
    clues_across = attr.ib(type=typing.Dict[int, str])
    clues_down = attr.ib(type=typing.Dict[int, str])
    notes = attr.ib(type=typing.Optional[str], default="")

    def to_dict(self):
        r"""Converts a Puzzle instance into a python dictionary.

        This method converts the current instance into a dictionary suitable for
        transforming to JSON.  This means that any field that's not suitable
        for a direct conversion to JSON will be changed to their logical
        equivalents (for example dates are change to ISO8601 date strings).

        The benefit of this method is that it can be called instead of `to_json`
        when a Puzzle object is stored as just one part of a larger tree of
        objects.
        """
        d = attr.asdict(self)

        # JSON doesn't support datetime types, so convert to an ISO8601 date
        # string.
        if d["published"] is not None:
            d["published"] = d["published"].isoformat()

        return d

    def to_json(self):
        r"""Converts a Puzzle instance into a JSON string."""
        return json.dumps(self.to_dict())


def load_puz_puzzle_from_bytes(bs):
    r"""Loads a crossword puzzle from the bytes of a .puz file.

    This method uses the puzpy library to load a crossword puzzle in .puz
    format.  Documentation for the puzpy library can be found here:
    https://github.com/alexdej/puzpy

    If the puzzle cannot be loaded for some reason, `None` is returned.
    """
    data = puz.load(bs)

    # Check if the puzzle is locked, if it is we might not be able to load the
    # correct solution from it.  Try to brute force what the 4 digit key is.
    if data.is_solution_locked():
        for key in range(1000, 10000):
            if data.unlock_solution(key):
                break

    if data.is_solution_locked():
        return None

    numbering = data.clue_numbering()
    rebus = data.rebus()

    rows = data.height
    cols = data.width
    title = html.unescape(data.title)
    published = None  # not exposed in .puz format
    publisher = None  # not exposed in .puz format
    author = html.unescape(data.author).strip()
    if author.startswith("by ") or author.startswith("By "):
        author = author[3:]

    notes = data.notes or ""

    cells = [["" for _ in range(cols)] for _ in range(rows)]
    cell_blocks = [[True for _ in range(cols)] for _ in range(rows)]
    for row in range(rows):
        for col in range(cols):
            offset = row * cols + col

            # Make sure to handle rebus cells.
            if data.has_rebus() and rebus.table[offset] != 0:
                c = rebus.solutions[rebus.table[offset] - 1]
            else:
                c = data.solution[offset]

            cells[row][col] = c if c != "." else ""
            cell_blocks[row][col] = c == "."

    cell_clue_numbers = [[0 for _ in range(cols)] for _ in range(rows)]
    for clue in numbering.across + numbering.down:
        num = clue["num"]
        row = clue["cell"] // cols
        col = clue["cell"] % cols
        cell_clue_numbers[row][col] = num

    cell_circles = [[False for _ in range(cols)] for _ in range(rows)]
    if data.extensions.get(puz.Extensions.Markup):
        for row in range(rows):
            for col in range(cols):
                mu = data.extensions[puz.Extensions.Markup][row * cols + col]
                cell_circles[row][col] = (mu & puz.GridMarkup.Circled) > 0

    clues_across = {}
    for clue in numbering.across:
        clues_across[clue["num"]] = clue["clue"]

    clues_down = {}
    for clue in numbering.down:
        clues_down[clue["num"]] = clue["clue"]

    return Puzzle(
        rows=rows,
        cols=cols,
        title=title,
        published=published,
        publisher=publisher,
        author=author,
        cells=cells,
        cell_blocks=cell_blocks,
        cell_clue_numbers=cell_clue_numbers,
        cell_circles=cell_circles,
        clues_across=clues_across,
        clues_down=clues_down,
        notes=notes,
    )
