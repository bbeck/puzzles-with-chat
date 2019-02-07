import flask
import redis

connection = None


def get_db():
    r"""Obtain a connection to the database.

    This method will attempt to reuse a global connection to the database if
    one is available.  Otherwise it will create the global connection and
    return it.
    """
    global connection
    if connection is None:
        url = flask.current_app.config["REDIS_URL"]
        connection = redis.Redis.from_url(url)

    return connection
