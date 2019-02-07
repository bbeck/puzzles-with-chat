import attr
import crosswords
import flask
import flask_socketio
import os
import rooms

app = flask.Flask(__name__)
app.config.from_mapping(
    REDIS_URL=os.environ.get("REDIS_URL", "redis://localhost:6379"),
    ROOM_TTL=int(os.environ.get("ROOM_TTL", 4 * 60 * 60)),
)

socketio = flask_socketio.SocketIO(app)


@app.route("/")
def index():
    r"""Render an index of which crosswords channels are available."""
    return flask.render_template("index.html")


@app.route("/favicon.ico")
def favicon():
    # We don't currently have a favicon so return a 404.  Without this Flask
    # will call into the channel route and treat favicon.ico as the name of a
    # channel which is not what we want.
    flask.abort(404)


@app.route("/<channel>")
def channel(channel):
    r"""Render a user view of the crossword.

    This view is read-only and intended for distribution to anyone who wants
    to see the crossword as its being solved.
    """
    return flask.render_template("channel.html", owner=channel, streamer=False)


@app.route("/<channel>/streamer")
def streamer(channel):
    r"""Render a streamer friendly view of the crossword.

    A streamer friendly view that is consistent in positioning and sizing, has
    options for controlling which crossword is currently being used as well as
    any configurable display options for streaming, etc.

    In addition when this URL is hit the server will have a chat bot join the
    streamer's chat if the bot is not already there.

    The intention is that only the streamer will have access to this particular
    view.
    """
    # Take the fact that a streamer has loaded their channel page as an
    # indication that they're about to start solving a puzzle with their chat
    # and have the chat bot join their channel.
    # TODO: Code this.

    return flask.render_template("channel.html", owner=channel, streamer=True)


@socketio.on("join")
def join(name):
    r"""Handler that's called when a client has requested to join a room."""
    # Tell the socketio backend that this particular socket is joining the
    # room.  This will allow this socket to receive broadcast events in the
    # future.
    flask_socketio.join_room(name)

    # Whenever someone joins the room send the current state of the puzzle
    # to them so that they can render it in their browser.  This will send
    # the message to just the client that joined the room.  If there's not
    # a current puzzle then don't send anything.
    room = rooms.get_room(name)
    if room is not None:
        puzzle = attr.evolve(room.puzzle, cells=room.cells)

        # Let this user know about the puzzle.
        flask_socketio.emit("crossword", puzzle.to_json())


@socketio.on("set_puzzle")
def set_crossword(data):
    r"""Handler that's called when the streamer changes the puzzle."""
    puzzle = crosswords.load_puzzle(data["date"])
    if puzzle is None:
        # Something went wrong loading the puzzle.  There's nothing more we can
        # do so return.  TODO: Log/emit some type of error here.
        return

    # Setup the cells list for the new solve.
    cells = [[
        "" if puzzle.cells[y][x] is not None else None
        for x in range(puzzle.cols)
    ] for y in range(puzzle.rows)]

    # Save the state of the room to the database.
    rooms.set_room(data["room"], rooms.Room(puzzle=puzzle, cells=cells))

    # Update the puzzle to have the empty set of cells before sending to the
    # clients so that we don't send the answers to the browser.
    puzzle = attr.evolve(puzzle, cells=cells)

    # Let everyone know about the updated puzzle.
    flask_socketio.emit("crossword", puzzle.to_json(), broadcast=True)


if __name__ == "__main__":
    socketio.run(app, host="0.0.0.0")
