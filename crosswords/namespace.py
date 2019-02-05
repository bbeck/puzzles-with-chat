import attr
import flask_socketio as socketio
import json
import typing
from .api import load_puzzle
from .puzzle import Puzzle


class CrosswordNamespace(socketio.Namespace):
    r"""SocketIO namespace for crosswords.

    This namespace keeps track of which rooms exist as well as the state
    of each room's current puzzle.  This namespace is where all of the
    intelligence goes about what happens when various events are received.

    Attributes
    ----------
    rooms : Set[str]
        The name of the rooms that are currently active.  These correspond to
        the SocketIO rooms that can have events emitted to them as well as the
        channel owners that are actively solving a crossword puzzle.

    states : Dict[str, State]
        The current state indexed by the name of the active room.  The state
        contains all of the relevant information about the particular solution
        for the puzzle.
    """

    def __init__(self):
        super().__init__("/crosswords")
        self.rooms = set()
        self.states = {}

    def on_set_crossword(self, data):
        r"""Handler for the streamer setting which puzzle they're working on."""
        room = data["room"]
        date = data["date"]

        puzzle = load_puzzle(date)
        if puzzle is None:
            # Something went wrong loading the puzzle.  There's nothing more we
            # can do so return.
            # TODO: Log/emit some type of error here.
            return

        # Create the current (empty) solution to the puzzle making sure to
        # preserve any cells which are None.
        cells = []
        for y in range(puzzle.rows):
            row = []
            for x in range(puzzle.cols):
                # TODO: Change this to append an empty string instead of the
                # solution from the puzzle.
                row.append(puzzle.cells[y][x]
                           if puzzle.cells[y][x] is not None else None)
            cells.append(row)

        # Setup the new state
        state = State(puzzle=puzzle, cells=cells)
        self.states[room] = state

        # Let everyone know about the updated puzzle.
        socketio.emit("crossword", state.to_json(), broadcast=True)

    def on_join(self, data):
        r"""Handler for a client requesting to join a particular room."""
        room = data["room"]

        # Tell the socketio backend that this particular socket is joining the
        # room.  This will allow this socket to receive broadcast events in the
        # future.
        socketio.join_room(room)
        self.rooms.add(room)

        # Whenever someone joins the room send the current state of the puzzle
        # to them so that they can render it in their browser.  This will send
        # the message to just the client that joined the room.  If there's not
        # a current puzzle then don't send anything.
        state = self.states.get(room)
        if state is not None:
            socketio.emit("crossword", state.to_json())


@attr.s
class State(object):
    r"""State represents the current state of a room's crossword puzzle solve.

    The state is comprised of the puzzle that is being solved as well as the
    current cell values.

    Attributes
    ----------
    puzzle : crosswords.Puzzle
        The puzzle (including solution) that is being solved.

    cells : List[List[str]]
        The current cell values.
    """
    puzzle = attr.ib(type=Puzzle)
    cells = attr.ib(type=typing.List[typing.List[str]])

    def to_json(self):
        r"""Converts the current state to a JSON string.

        The JSON representation of the state is a mashup of the puzzle and the
        current cells.  This is done so as to not send the solution to the
        puzzle to the browser.  The structure of the state as JSON looks as if
        it were a complete Puzzle object converted to JSON.

        The majority of the fields of the puzzle type are suitable for direct
        conversion to JSON.  For the ones that are not suitable for direct
        conversion to JSON they are changed to be logical equivalents (for
        example dates are changed to ISO8601 date strings).

        Returns
        -------
        str
            The representation of the state as a JSON string.
        """
        d = attr.asdict(self.puzzle)
        d["cells"] = self.cells

        # JSON doesn't support datetimes, so convert to an ISO8601 date string.
        d["published"] = d["published"].isoformat()

        return json.dumps(d)
