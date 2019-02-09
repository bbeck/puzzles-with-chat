import irc.bot
import irc.client


class Bot(irc.bot.SingleServerIRCBot):
    def __init__(self, username, token, handle_message):
        self.token = token
        self.rooms = []
        self.handler = handle_message

        # Create IRC bot connection
        server = "irc.chat.twitch.tv"
        port = 6667
        super().__init__([(server, port, token)], username, username)

    def join(self, channel, key=""):
        # Remember this channel for later.
        if channel not in self.rooms:
            self.rooms.append(channel)

        self.connection.join(channel, key)

    def part(self, channels, message=""):
        # Forget the channels that we're parting.
        for channel in channels:
            try:
                self.rooms.remove(channel)
            except ValueError:
                # The channel wasn't in the list, ignore it.
                pass

        self.connection.part(channels, message)

    def on_welcome(self, connection, event):
        # You must request specific capabilities before joining channels
        connection.cap("REQ", ":twitch.tv/membership")
        connection.cap("REQ", ":twitch.tv/tags")
        connection.cap("REQ", ":twitch.tv/commands")

        # Join all of the rooms that we are configured to be in.
        for channel in self.rooms:
            connection.join("#" + channel)

    def on_pubmsg(self, connection, event):
        channel = event.target[1:] if event.target[0] == "#" else event.target
        message = event.arguments[0]

        self.handler(channel, message)
