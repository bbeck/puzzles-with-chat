import attr
import flask
import os
import puzzles
import rooms
import settings
from flask_socketio import SocketIO, emit, join_room


class ReverseProxied(object):
    r"""Helper for invoking WSGI requests with proper URL scheme.

    Depending on where the app is run it may be behind a load balander that
    is providing SSL termination transparently to the app.  The Flask `url_for`
    method doeesn't natively know when this is the case so from its perspective
    all requests to the app are coming over HTTP and thus it generates an
    external HTTP url instead of an HTTPS one.  Normally we could just tell
    `url_for` to generate URLs with a scheme of HTTPS, but then in
    development when a load balancer isn't being used it would be doing the
    incorret thing.  So this class wraps the WSGI app within Flask to detect
    if the incoming request has been forwarded to the app via a load balancer,
    and if so uses whatever protocol the request came into the load balancer
    with.
    """

    def __init__(self, app):
        self.app = app

    def __call__(self, environ, start_response):
        scheme = environ.get('HTTP_X_FORWARDED_PROTO')
        if scheme:
            environ['wsgi.url_scheme'] = scheme
        return self.app(environ, start_response)


app = flask.Flask(__name__)
app.wsgi_app = ReverseProxied(app.wsgi_app)
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
    return flask.render_template(
        "channel.html", owner=channel, streamer=False, progress=False)


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
    return flask.render_template(
        "channel.html", owner=channel, streamer=True, progress=False)


@app.route("/<channel>/progress")
def progress(channel):
    r"""Render a progress view of the crossword.

    The progress view is one that shows the progress of the solve without
    revealing any of the answers.  As cells are filled in they become shaded
    in the grid, but the actual letter that they're filled with does not show.

    The intention of this view is that it can be shared with others that are
    solving the puzzle at the same time so that they can see the progress
    that's happening, but not any of the actual answers.
    """
    return flask.render_template(
        "channel.html", owner=channel, streamer=False, progress=True)


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

    room = rooms.get_room(channel)
    if room is None:
        flask.abort(404)  # 404 = Not Found

    if room.play_pause_state != "playing":
        flask.abort(409)  # 409 = Conflict (state conflicted with request)

    answer = flask.request.get_data(as_text=True)
    if answer is None:
        flask.abort(400)  # 400 = Bad Request

    answer = answer.replace(" ", "").strip().upper()
    if len(answer) == 0:
        flask.abort(400)  # 400 = Bad Request

    room_settings = settings.get_settings(channel)
    if room_settings.only_allow_correct_answers:
        correct = rooms.get_correct_answer(channel, clue)

        if len(answer) != len(correct):
            flask.abort(403)  # 403 = Forbidden

        # An answer could have '.' characters in it because only some
        # characters are known.  This is still technically a correct answer
        # so make sure to allow it.
        for i in range(len(correct)):
            if answer[i] != "." and answer[i] != correct[i]:
                flask.abort(403)  # 403 = Forbidden

    allow_clearing = not room_settings.only_allow_correct_answers
    room = rooms.apply_answer(room, channel, clue, answer, allow_clearing)
    if room is None:
        # We should technically distinguish between couldn't find the clue and
        # couldn't fit the answer...
        flask.abort(404)  # 404 = Not Found

    # Now that we've updated the room, send it to everyone, sanitizing the
    # puzzle from having answers first.
    puzzle = attr.evolve(room.puzzle, cells=room.cells)
    room = attr.evolve(room, puzzle=puzzle)
    socketio.emit("state", room.to_json(), room=channel)

    # Check and see if we've solved the puzzle, if so let everyone know.
    if room.play_pause_state == "complete":
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

    room = rooms.get_room(name)
    if room is None:
        return

    # Whenever someone joins the room send the current state of the puzzle
    # to them so that they can render it in their browser.  This will send
    # the message to just the client that joined the room.  If there's not
    # a current puzzle then don't send anything.
    puzzle = attr.evolve(room.puzzle, cells=room.cells)
    room = attr.evolve(room, puzzle=puzzle)
    room_settings = settings.get_settings(name)

    # Let this user know about the puzzle and settings.
    emit("state", room.to_json())
    emit("settings", room_settings.to_json())


@socketio.on("set_puzzle")
def set_crossword(data):
    r"""Handler that's called when the streamer changes the puzzle."""
    room_name = data["room"]
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

    # Setup the set of filled in clues for the new solve.
    across_clues_filled = {num: False for num in puzzle.across_clues}
    down_clues_filled = {num: False for num in puzzle.down_clues}

    # Save the state of the room to the database.  We start all puzzles by
    # default in the paused state.
    room = rooms.Room(
        play_pause_state="created",
        puzzle=puzzle,
        cells=cells,
        across_clues_filled=across_clues_filled,
        down_clues_filled=down_clues_filled,
    )
    rooms.set_room(room_name, room)

    # Update the puzzle to have the empty set of cells before sending to the
    # clients so that we don't send the answers to the browser.
    puzzle = attr.evolve(puzzle, cells=cells)
    room = attr.evolve(room, puzzle=puzzle)

    # Let everyone know about the updated puzzle.
    emit("state", room.to_json(), room=room_name)


@socketio.on("set_settings")
def set_settings(data):
    r"""Handler that's called when the streamer changes the settings."""
    room_name = data["room"]
    changes = data["settings"]  # A key/value dict.

    existing_settings = settings.get_settings(room_name)
    updated_settings = attr.evolve(existing_settings, **changes)
    settings.set_settings(room_name, updated_settings)

    # Let everyone know about the settings update.
    emit("settings", updated_settings.to_json(), room=room_name)

    # If we've enabled only correct answers, we should clear any cells with
    # an incorrect value in them.  We also have to update which puzzle clues
    # have been filled in becuase we might remove some filled in cells.
    if updated_settings.only_allow_correct_answers:
        room = rooms.get_room(room_name)

        room = rooms.clear_incorrect_cells(room, room_name)

        # Update the puzzle to have the empty set of cells before sending to the
        # clients so that we don't send the answers to the browser.
        puzzle = attr.evolve(room.puzzle, cells=room.cells)
        room = attr.evolve(room, puzzle=puzzle)

        # Let everyone know about the updated puzzle.
        emit("state", room.to_json(), room=room_name)


@socketio.on("play_pause")
def play_pause(data):
    r"""Handler that's called when the play/pause button is clicked."""
    room_name = data["room"]

    room = rooms.get_room(room_name)
    if room is None:
        return

    # Update the state of the room.
    state = room.play_pause_state
    if state == "created" or state == "paused":
        state = "playing"
    elif state == "playing":
        state = "paused"

    # Save the updated room.
    room = attr.evolve(room, play_pause_state=state)
    rooms.set_room(room_name, room)

    # Let everyone know about the settings update.
    puzzle = attr.evolve(room.puzzle, cells=room.cells)
    room = attr.evolve(room, puzzle=puzzle)
    emit("state", room.to_json(), room=room_name)


if __name__ == "__main__":
    socketio.run(app, host="0.0.0.0")
