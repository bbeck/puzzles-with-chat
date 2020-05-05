# Puzzles With Chat

This is an application + twitch.tv bot that allows a group of users in a
channel to collaboratively solve a puzzle together.  Currently, the supported
puzzle types are crosswords (including cryptic crosswords) and spelling bees.

## How does it work? (For streamers)

The application consists of three main components the UI, API and bot.  As a
streamer you navigate a web browser to the website hosting the application and
setup a video capture of the contents of that browser window for your stream
to see.  Using the web site you can select a puzzle for your stream to solve.
One a puzzle has been selected a twitch.tv bot (PuzzlesWithChat) will join your 
stream's chatroom and listen for commands from yourself and viewers.  As the
application is instructed to fill in parts of the puzzle solution they will
automatically appear in the UI of the application for your viewers to see.  They
can then build upon each other's answers in order to fully solve the puzzle.

## Streamers using this application

While originally written for jeffxvx's stream, several other streamers have
successfully used the app in the past.  Below are links to their streams, please
note they may not be regular users of the application.

* [jeffxvx](https://twitch.tv/jeffxvx) - Sunday evenings around 6-7pm Eastern.
* [aidanwould](https://twitch.tv/aidanwould)
* [snapdragon64](https://twitch.tv/snapdragon64)

If you'd like to use the application as part of your stream, please feel free to
do so, it is free for everyone to use.  If you need help or support in figuring 
out how to use it, have discovered a bug, have a feature request, or have
discovered a novel way to use the application that you'd like to share then feel 
free to submit a GitHub issue in this repo or send a whisper to MistaeksWereMade
on Twitch.  Additionally feel free to reach out if you'd like your channel 
listed above.

## Architecture

The architecture of the application is one that is composed of microservices.
There is an API service that acts as the integration point for both the UI
and bot services. There is also a redis data store that the API uses for storing
all data that requires persistence.

The API generally provides RESTful URLs for selecting puzzles, updating settings
or providing answers to parts of the active puzzle.  The API also provides
several event based URLs (powered by server sent events) where events about the
state of the application are sent.  There are SSE endpoints that indicate which
puzzles are being solved as well as endpoints for monitoring the progress of a
specific solve.  These endpoints are used by both the UI and bot services.


## Development environment

The requirements for a working development environment are minimal.  Only
Docker and Docker Compose are expected to be installed on the development 
machine.  There is a `docker-compose.yml` file in the root of the repository as
that defines each of the services and wires them together and runs them in
development mode (a mode where changes to source files transparently recompile
and restart the relevant application).  There is also a `Dockerfile` in each
service directory that describes how to build that particular service's 
container.

A `Makefile` is also provided that provides helpers for invoking Docker Compose.
`make help` will show all supported commands as well as a description of what
each does.

Once running the UI will be visible on [`localhost:3000`](http://localhost:3030).


## Contributions

Contributions are always welcome, however if you plan on making an architectural
or other large or wide sweeping change, please reach out first before doing the 
work -- a GitHub issue is a great way to start that conversation.  There is a 
specific vision for the direction of the project and not all changes align with
that.  Let's have a discussion before you commit your valuable time to make sure
your proposal is something that can ultimately become part of the project.
