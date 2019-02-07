import attr
import datetime
import json
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
