import attr
import flask
import os
import puzzles
import rooms
from flask_socketio import SocketIO, emit, join_room

app = flask.Flask(__name__)
app.config.from_mapping(
    REDIS_URL=os.environ.get("REDIS_URL", "redis://localhost:6379"),
    ROOM_TTL=int(os.environ.get("ROOM_TTL", 4 * 60 * 60)),
)

socketio = SocketIO(app)


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
    return flask.render_template("channel.html", owner=channel, streamer=True)


@app.route("/<channel>/show/<clue>")
def show_clue(channel, clue):
    r"""Update the UI to make a clue visibe.

    This route is a REST handler that triggers a request to make a clue
    visible on the screen.  No response body is returned and this method always
    returns a HTTP 204 regardless of its success or not.  The reason a HTTP 204
    is always returned is because making a clue visible is a client side
    operation, and the server currently doesn't have a way to know if it has
    succeeded or not, it just knows that the request has been made.
    """
    socketio.emit("show_clue", clue, room=channel)
    return ("", 204, {})


@app.route("/<channel>/answer/<clue>", methods=["PUT"])
def answer(channel, clue):
    r"""Apply an answer to the specified crossword clue.

    This route is a REST handler that expects a string body whose text
    contains the answer to the clue specified in the URL.  No response body
    is returned, just an HTTP 200 upon successful specifying of the answer and
    a HTTP 4xx when applying the answer fails.
    """
    if flask.request.content_length > 1024:
        flask.abort(413)  # 413 = Payload Too Large

    answer = flask.request.get_data(as_text=True)
    if answer is None or len(answer.strip()) == 0:
        flask.abort(400)  # 400 = Bad Request
    answer = answer.upper()

    room = rooms.apply_answer(channel, clue, answer)
    if room is None:
        # We should technically distinguish between couldn't find the clue and
        # couldn't fit the answer...
        flask.abort(404)  # 404 = Not Found

    # Now that we've updated the room, send the puzzle to everyone.
    puzzle = attr.evolve(room.puzzle, cells=room.cells)
    socketio.emit("crossword", puzzle.to_json(), room=channel)

    # Check and see if we've solved the puzzle, if so let everyone know.
    if room.puzzle.cells == room.cells:
        socketio.emit("solved", room=channel)

    # ...and return a HTTP 204 = No Content (server processed the request but
    # hasn't generated any content to return).
    return ("", 204, {})


@app.route("/bot/channels")
def channels():
    r"""Return the channels that the bot should be present in.

    This route is a REST endpoint that tells the bot which channels have an
    active puzzle and thus which channels the bot should join.  The response is
    a JSON object with a "channels" property that is the list of channels to
    join.
    """
    channels = rooms.get_all_room_names()
    return flask.jsonify(channels=channels)


@socketio.on("join")
def join(name):
    r"""Handler that's called when a client has requested to join a room."""
    # Tell the socketio backend that this particular socket is joining the
    # room.  This will allow this socket to receive room events in the future.
    join_room(name)

    # Whenever someone joins the room send the current state of the puzzle
    # to them so that they can render it in their browser.  This will send
    # the message to just the client that joined the room.  If there's not
    # a current puzzle then don't send anything.
    room = rooms.get_room(name)
    if room is not None:
        puzzle = attr.evolve(room.puzzle, cells=room.cells)

        # Let this user know about the puzzle.
        emit("crossword", puzzle.to_json())


@socketio.on("set_puzzle")
def set_crossword(data):
    r"""Handler that's called when the streamer changes the puzzle."""
    room = data["room"]
    date = data["date"]

    puzzle = puzzles.load_puzzle(date)
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
    rooms.set_room(room, rooms.Room(puzzle=puzzle, cells=cells))

    # Update the puzzle to have the empty set of cells before sending to the
    # clients so that we don't send the answers to the browser.
    puzzle = attr.evolve(puzzle, cells=cells)

    # Let everyone know about the updated puzzle.
    emit("crossword", puzzle.to_json(), room=room)


if __name__ == "__main__":
    socketio.run(app, host="0.0.0.0")
