import flask
import puzzle

app = flask.Flask(__name__)


@app.route("/puz", methods=["POST"])
def puz():
    if flask.request.content_length > 10 * 1024 * 1024:
        flask.abort(413)  # 413 = Payload Too Large

    p = puzzle.load_puz_puzzle_from_bytes(flask.request.get_data())
    if p is None:
        flask.abort(400)  # 400 = Bad Request

    return p.to_json()


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5001)
