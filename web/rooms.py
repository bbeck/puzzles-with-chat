import attr
import crosswords
import db
import flask
import json
import typing


@attr.s(frozen=True)
class Room(object):
    r"""Room represents an active room.

    An active room is one that is currently attempting to solve a crossword and
    should have a chat bot monitoring it.

    Attributes
    ----------
    puzzle : crosswords.Puzzle
        The puzzle that is currently being solved.

    cells : List[List[str]]
        The current cell values.
    """
    puzzle = attr.ib(type=crosswords.Puzzle)
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
            puzzle=crosswords.Puzzle.from_json(d["puzzle"]),
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
