import attr
import db
import flask
import json
import puzzles
import typing


@attr.s(frozen=True)
class Room(object):
    r"""Room represents an active room.

    An active room is one that is currently attempting to solve a crossword and
    should have a chat bot monitoring it.

    Attributes
    ----------
    puzzle : puzzles.Puzzle
        The puzzle that is currently being solved.

    cells : List[List[str]]
        The current cell values.
    """
    puzzle = attr.ib(type=puzzles.Puzzle)
    cells = attr.ib(type=typing.List[typing.List[str]])

    def to_json(self):
        r"""Converts the current room to a JSON string.

        Returns
        -------
        str
            The representation of the room as a JSON string.
        """
        return json.dumps({
            "puzzle": self.puzzle.to_json(),
            "cells": self.cells,
        })

    @classmethod
    def from_json(cls, s):
        d = json.loads(s)
        return cls(
            puzzle=puzzles.Puzzle.from_json(d["puzzle"]),
            cells=d["cells"],
        )


def get_room(name):
    r"""Load a room from the redis database.

    Rooms are stored in the redis database under a key with the hardcoded string
    "room:" concatenated with the the channel's name.  This allows all rooms
    to be easily scanned.  After any read or write operation to a room the
    key's expiration is automatically updated.

    Parameters
    ----------
    name : str
        The name of the room to retrieve from the database.

    Returns
    -------
    Room|None
        The room if it was present in the database or None if it wasn't present.
    """
    redis = db.get_db()
    key = f"room:{name}"

    s = redis.get(key)
    if s is None:
        return None

    redis.expire(key, flask.current_app.config["ROOM_TTL"])

    return Room.from_json(s)


def set_room(name, room):
    r"""Save a room to the redis database.

    See `get_room` for a description of how rooms are stored.  When a room is
    saved to the redis database it's expiration time is also specified.

    Parameters
    ----------
    name : str
        The name of the room to save to the database.

    room : Room
        The room to save to the database.
    """
    key = f"room:{name}"
    s = room.to_json()

    redis = db.get_db()
    redis.set(key, s, ex=flask.current_app.config["ROOM_TTL"])


def apply_answer(name, clue, answer):
    r"""Apply an answer to a puzzle.

    This method will attempt to identify the clue that's been specified and
    apply the provided answer to it.  If the room or the clue cannot be
    identified or the answer doesn't fit properly then None will be returned.

    Parameters
    ----------
    name : str
        The name of the room to apply the answer to.

    clue : str
        The id of the clue to apply the answer to.

    answer : str
        The answer to apply to the clue.
    """
    room = get_room(name)
    if room is None:
        return None

    # Parse the clue into its components
    num, direction = parse_clue(clue)
    if num is None or direction is None:
        return None

    answer = parse_answer(answer)
    coordinates = puzzles.get_answer_cells(room.puzzle, num, direction)
    if len(answer) != len(coordinates):
        return None

    # Otherwise write the answer values into the puzzle cells and save the
    # room's state.
    for (value, (x, y)) in zip(answer, coordinates):
        room.cells[y][x] = value

    set_room(name, room)

    return room


def parse_clue(clue):
    r"""Parse a clue identifier string into the number and direction.

    Parameters
    ----------
    clue : str
        The clue identifier of the form `<number><a|d>`.

    Returns
    -------
    typing.Tuple[int, str]|typing.Tuple[None, None]
        The number and direction of the clue if it can be parsed properly or
        a tuple of None values if it cannot be parsed properly.
    """
    if clue is None:
        return (None, None)

    clue = clue.strip()
    if len(clue) == 0:
        return (None, None)

    try:
        number = int(clue[:-1])
    except:
        return (None, None)

    direction = clue[-1].lower()
    if direction != "a" and direction != "d":
        return (None, None)

    return (number, direction)


def parse_answer(answer):
    r"""Parse an answer string into a list of cell values.

    The answer string is parsed in such a way to look for cell values that
    contain multiple characters (aka a rebus).  It does this by looking for
    parenthesized groups of letters.  For example the string `(red)velvet`
    would be interpreted as ["red", "v", "e", "l", "v", "e", "t"] and fit as
    the answer for a 7 cell clue.

    Additionally if an answer contains a " " character anywhere that particular
    cell will be left empty.  This allows strings like `    s` to be entered
    to indicate that the answer is known to be plural, but the other letters
    aren't known yet.  Within a rebus cell " " characters are kept as-is.

    Parameters
    ----------
    answer : str
        The answer to parse.

    Returns
    -------
    List[str]
        The individual values that should be placed in each cell for this
        answer.  In the case of a rebus the entry may container more than one
        letter.
    """
    cells = []
    inParens = False
    for c in answer:
        # Check if we're closing a pair of parentheses.
        if inParens and c == ")":
            inParens = False
            continue

        # If we're within parentheses then just keep appending to the last
        # cell.
        if inParens:
            # Keep appending to the last cell.
            cells[-1] += c
            continue

        # If we're opening a pair of parentheses then create a new empty cell
        # that we'll accumulate multiple values into.
        if c == "(":
            inParens = True
            cells.append("")
            continue

        # We're not in parentheses, this is just a normal character or an
        # empty cell.
        cells.append(c if c != " " else "")

    # TODO: Should we check for unbalanced parentheses?
    return cells
