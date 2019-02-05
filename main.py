import crosswords
import flask
import flask_socketio

app = flask.Flask(__name__)
socketio = flask_socketio.SocketIO(app)


@app.route("/")
def index():
    r"""Render an index of which crosswords channels are available."""
    return flask.render_template("index.html")


@app.route("/favicon.ico")
def favicon():
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

    The intention is that only the streamer will have access to this particular
    view.
    """
    return flask.render_template("channel.html", owner=channel, streamer=True)


if __name__ == "__main__":
    socketio.on_namespace(crosswords.CrosswordNamespace())
    socketio.run(app, debug=True, host="0.0.0.0")
