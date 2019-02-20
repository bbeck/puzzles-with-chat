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
    play_pause_state : str
        The state of the solve.  Valid values are: "created", "paused",
        "playing", or "complete".

    puzzle : puzzles.Puzzle
        The puzzle that is currently being solved.

    cells : List[List[str]]
        The current cell values.

    across_clues_filled : Dict[int, bool]
        Whether or not an across clue has had an answer filled in, indexed by
        clue number.

    down_clues_filled : Dict[int, bool]
        Whether or not a down clue has had an answer filled in, indexed by clue
        number.
    """
    play_pause_state = attr.ib(type=str)
    puzzle = attr.ib(type=puzzles.Puzzle)
    cells = attr.ib(type=typing.List[typing.List[str]])
    across_clues_filled = attr.ib(type=typing.Dict[int, bool])
    down_clues_filled = attr.ib(type=typing.Dict[int, bool])

    def to_json(self):
        r"""Converts the current room to a JSON string.

        Returns
        -------
        str
            The representation of the room as a JSON string.
        """
        return json.dumps({
            "play_pause_state": self.play_pause_state,
            "puzzle": self.puzzle.to_dict(),
            "cells": self.cells,
            "across_clues_filled": self.across_clues_filled,
            "down_clues_filled": self.down_clues_filled,
        })

    @classmethod
    def from_json(cls, s):
        d = json.loads(s)

        # Convert the keys for across_clues_filled and down_clues_filled to
        # ints.  JSON doesn't support non-string keys for objects.
        d["across_clues_filled"] = {
            int(n): v
            for (n, v) in d["across_clues_filled"].items()
        }
        d["down_clues_filled"] = {
            int(n): v
            for (n, v) in d["down_clues_filled"].items()
        }

        return cls(
            play_pause_state=d["play_pause_state"],
            puzzle=puzzles.Puzzle.from_dict(d["puzzle"]),
            cells=d["cells"],
            across_clues_filled=d["across_clues_filled"],
            down_clues_filled=d["down_clues_filled"],
        )


def get_all_room_names():
    r"""Return all active room names from the redis database.

    Returns
    -------
    List[str]
        The active room names.
    """
    redis = db.get_db()
    return [parse_room_name(name) for name in redis.scan_iter(match="room:*")]


def get_room(name):
    r"""Load a room from the redis database.

    Rooms are stored in the redis database under a key with the hardcoded string
    "room:" concatenated with the channel's name.  This allows all rooms to be
    easily scanned.  After any read or write operation to a room the key's
    expiration is automatically updated.

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


def get_correct_answer(name, clue):
    r"""Get the correct answer for a clue from the current puzzle.

    This method will attempt to identify the clue that's been specified and
    provide the answer for it.  If the room or the clue cannot be identified
    then None will be returned.

    Parameters
    ----------
    name : str
        The name of the room to retrieve the answer of.

    clue : str
        The id of the clue to retrieve the answer of.
    """
    room = get_room(name)
    if room is None:
        return None

    # Parse the clue into its components
    num, direction = parse_clue(clue)
    if num is None or direction is None:
        return None

    # Build the answer taking into account rebus cells.
    answer = ""
    for x, y in puzzles.get_answer_cells(room.puzzle, num, direction):
        value = room.puzzle.cells[y][x]
        if len(value) == 1:
            answer += value
        else:
            answer += f"({value})"

    return answer


def apply_answer(room, name, clue, answer):
    r"""Apply an answer to a puzzle.

    This method will attempt to identify the clue that's been specified and
    apply the provided answer to it.  If the room or the clue cannot be
    identified or the answer doesn't fit properly then None will be returned.

    Parameters
    ----------
    room : Room
        The room to apply the answer to.

    name : str
        The name of the room to apply the answer to.

    clue : str
        The id of the clue to apply the answer to.

    answer : str
        The answer to apply to the clue.
    """
    # Parse the clue into its components
    num, direction = parse_clue(clue)
    if num is None or direction is None:
        return None

    answer = parse_answer(answer)
    coordinates = puzzles.get_answer_cells(room.puzzle, num, direction)
    if len(answer) != len(coordinates):
        return None

    # Otherwise write the answer values into the puzzle cells.
    for (value, (x, y)) in zip(answer, coordinates):
        room.cells[y][x] = value

    # A single answer can actually answer multiple clues, so instead of
    # checking if we've just filled this one clue we'll do a quick check of
    # the entire puzzle.
    for num in room.puzzle.across_clues:
        room.across_clues_filled[num] = all(
            room.cells[y][x] != ""
            for (x, y) in puzzles.get_answer_cells(room.puzzle, num, "a"))
    for num in room.puzzle.down_clues:
        room.down_clues_filled[num] = all(
            room.cells[y][x] != ""
            for (x, y) in puzzles.get_answer_cells(room.puzzle, num, "d"))

    # Check and see if we've finished the puzzle, if so update the state to
    # complete.
    if room.cells == room.puzzle.cells:
        room = attr.evolve(room, play_pause_state="complete")

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

    Additionally if an answer contains a "." character anywhere that particular
    cell will be left empty.  This allows strings like `....s` to be entered
    to indicate that the answer is known to be plural, but the other letters
    aren't known yet.  Within a rebus cell "." characters are kept as-is.

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
        cells.append(c if c != "." else "")

    # TODO: Should we check for unbalanced parentheses?
    return cells


def parse_room_name(b):
    r"""Parse a room name from the redis database.

    The room name is returned from redis as a set of bytes with a prefix of
    "room:".  This method converts the bytes to a string and removes the
    prefix.

    Parameters
    ----------
    b : List[byte]
        The bytes returned from redis as the name of the room.

    Returns
    -------
    str
        The name of the room.
    """
    s = b.decode("utf-8")
    return s[5:]
