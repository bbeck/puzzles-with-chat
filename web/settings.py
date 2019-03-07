import attr
import db
import json
import typing


@attr.s(frozen=True)
class Settings(object):
    r"""Settings represents the settings for a channel.

    Settings are the optional behaviors that can be enabled or disabled by
    the streamer for their channel.

    Attributes
    ----------
    only_allow_correct_answers : bool
        When enabled only correct answers will be filled into the puzzle grid.

    hide_clues : str
        Which clues to hide from the user.
    """
    # It's important to provide sane default values so that settings can be
    # easily added over time.
    only_allow_correct_answers = attr.ib(type=bool, default=False)
    hide_clues = attr.ib(type=str, default="none")

    def to_json(self):
        r"""Converts a Settings instance into a JSON string.

        Returns
        -------
        str
            The JSON representation of the current settings instance.
        """
        d = attr.asdict(self)
        return json.dumps(d)

    @classmethod
    def from_json(cls, s):
        r"""Converts a JSON string to a new Settings instance.

        Parameters
        ----------
        s : str
            The JSON string representation of a Settings object.

        Returns
        -------
        Settings
            The Settings instance corresponding to the inputted JSON string.
        """
        d = json.loads(s)
        return cls(**d)


def get_settings(name):
    r"""Load settings for a channel from the redis database.

    Settings are stored in the redis database under a key with the hardcoded
    string "settings:" concatenated with the channel's name.  Settings do not
    expire.

    Parameters
    ----------
    name : str
        The name of the channel to retrieve from the database.

    Returns
    -------
    Settings
        The settings for the channel from the database.  If there was no entry
        in the database then the default settings object is returned.
    """
    redis = db.get_db()
    key = f"settings:{name}"

    s = redis.get(key)
    if s is None:
        return Settings()

    return Settings.from_json(s)


def set_settings(name, settings):
    r"""Save settings for a channel to the redis database.

    See `get_settings` for a description of how settings are stored.

    Parameters
    ----------
    name : str
        The name of the settings to save to the database.

    settings : Settings
        The settings to save to the database.
    """
    key = f"settings:{name}"
    s = settings.to_json()

    redis = db.get_db()
    redis.set(key, s)
