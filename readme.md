# Twitch Plays Crosswords

This is an application + twitch.tv bot that allows a group of users in a
channel to collboratively solve a crossword puzzle.

There are several features available to help make things more engaging
depending on the difficulty of the puzzle being solved.

* Correct answers only - In this mode only correct answers from chat will be
entered into the grid.  This is useful when solving a very difficult crossword
or with

* Reduced clues - In this mode only the across or down clues are shown.  This
provides an extra challenge for experienced solvers.


## How it works

The application is comprised of both a web site as well as a twitch.tv bot.
As a streamer you navigate a browser to the web site hosting the application
and capture the contents of that browser for your stream to see.  Once you've
connected to the web portion of the application a twitch.tv bot will join your
stream's chat room and listen for commands.

In your channel commands such as `!12a red velvet cake` can be entered to fill
in the grid cells with answers.  Also chatters can control which clues are
visible through commands like `!show 10d`.


## Streams

While originally written for jeffxvx's stream, several other streamers have
successfully used the app in the past.  I'm not sure if they're regular users
of it, but I've included links to their streams below.

* [jeffxvx](https://twitch.tv/jeffxvx) - Sunday evenings around 6-7pm Eastern.
* [aidanwould](https://twitch.tv/aidanwould)
* [snapdragon64](https://twitch.tv/snapdragon64)

If you'd like to use the application as part of your stream, please feel free to
do so.  If you need help or support in figuring out how to use the application
or have an idea of a feature request then feel free to submit a GitHub issue
or reach out directly to MistaeksWereMade on twitch.tv.  Also feel free to reach
out if you'd like your channel listed above.
