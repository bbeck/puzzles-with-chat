import os
import requests
import schedule
import sys
import threading
import time
import twitch

API_ENDPOINT = os.getenv("API_ENDPOINT")
BOT_USERNAME = os.getenv("BOT_USERNAME", "twitch_plays_crosswords")
BOT_OAUTH_TOKEN = os.getenv("BOT_OAUTH_TOKEN")

# The channels that the bot has currently joined.  This is kept up to date as
# the bot joins and parts from channels.
channels = []


def check_channels(join_func, part_func):
    r"""Check whether or not the bot should join or part any channels.

    This method will query the web app to determine which rooms have been
    setup for a solve, and join the necessary chat channels for the room if
    the bot is not already in the channel.

    A memory of which channels the bot is in is kept in the `channels` variable
    and consulted each time the room list is obtained from the server.

    If there is an error communicating with the web app then this function does
    nothing and returns -- using the current state of the world for which
    channels the bot should be in.

    Parameters
    ----------
    join_func : typing.Callable[[str], None]
        A function that can be used to indicate that the bot should join a
        channel.  The argument to the function is the name of the channel to
        join.

    part_func : typing.Callable[[str], None]
        A function that can be used to indicate that the bot should part a
        channel.  The argument to the function is the name of the channel to
        part.
    """
    global channels

    response = requests.get(f"{API_ENDPOINT}/bot/channels")
    if 400 <= response.status_code < 600:
        # 4xx and 5xx responses are client and server errors respectively.
        return

    try:
        data = response.json()
    except:
        return

    new_channels = data.get("channels")
    if new_channels is None:
        return

    # Join newly added channels
    for channel in set(new_channels) - set(channels):
        join_func("#" + channel)

    # Part removed channels
    for channel in set(channels) - set(new_channels):
        part_func("#" + channel)

    channels = new_channels


def handle_message(channel, message):
    words = message.split(" ", 2)

    if words[0] == "!answer" and len(words) == 3:
        clue = words[1].lower()
        answer = words[2]
        requests.put(f"{API_ENDPOINT}/{channel}/answer/{clue}", data=answer)

    if words[0] == "!show" and len(words) == 2:
        clue = words[1].lower()
        requests.get(f"{API_ENDPOINT}/{channel}/show/{clue}")


def main():
    if API_ENDPOINT is None:
        print("API_ENDPOINT environment variable must be present.")
        sys.exit(1)

    if BOT_OAUTH_TOKEN is None:
        print("BOT_AUTH_TOKEN environment variable must be present.")
        sys.exit(1)

    bot = twitch.Bot(BOT_USERNAME, BOT_OAUTH_TOKEN, handle_message)

    # Start the bot in a background thread.
    threading.Thread(target=bot.start).start()

    # Start polling in the background for the channels to join.
    schedule.every(10).to(20).seconds.do(
        check_channels, join_func=bot.join, part_func=bot.part)

    # Block forever while running the scheduler.  The bot is running in a
    # background thread.
    while True:
        schedule.run_pending()
        time.sleep(1)


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        # The user asked the program to exit
        sys.exit(1)
