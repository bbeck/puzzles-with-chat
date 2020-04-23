import React from "react";
import {Router} from "@reach/router";
import EventStream from "event-stream";
import Nav from "nav";
import CrosswordApp from "crossword/app";

export default function App() {
  return (
    <Router>
      <Home path="/"/>
      <ChannelHome path="/:channel/"/>

      <CrosswordApp path="/:channel/crossword/"/>
      <CrosswordApp path="/:channel/crossword/:view"/>

      {/* These routes are temporary redirects from old paths that are no longer valid. */}
      <ChannelRedirect path="/:channel/progress" view="progress"/>
      <ChannelRedirect path="/:channel/streamer" view="streamer"/>
    </Router>
  );
}

function Home() {
  const [crosswords] = React.useState(
    new EventStream(`/api/crossword/channels`)
  );
  const [crosswordChannels, setCrosswordChannels] = React.useState(null);
  React.useEffect(() => {
    crosswords.setHandler(message => {
      const event = JSON.parse(message.data);
      switch (event.kind) {
        case "channels":
          setCrosswordChannels(event.payload);
          break;

        case "ping":
          break;

        default:
          console.log("unhandled event:", event);
      }
    });
  }, [crosswords, setCrosswordChannels]);

  return (
    <div>
      <Nav/>
      <div className="jumbotron">
        <h1>Welcome to Twitch Plays Crosswords!</h1>
        <hr className="my-4"/>
        <p>
          Twitch Plays Crosswords is a web application that Twitch streamers can
          use to allow their chat to interactively solve a crossword puzzle. The
          streamer selects a puzzle to solve and then this application will have
          a chat bot join the streamer's chat. From there participants can input
          answers to the various crossword clues into the chat and the
          application will show them on screen.
        </p>
        <p>
          If you've ended up on this page you were probably looking to spectate
          a crossword solving session that is already in progress, but didn't
          get the full URL to the Twitch streamer's page. To help you below
          you'll find a list of all of the solving sessions that are in
          progress. Please click through to the streamer you were looking for.
        </p>

        <h6>Channels with active crosswords:</h6>
        <ActiveChannelList channels={crosswordChannels} puzzle="crossword"/>

        <hr className="my-4"/>

        <p>
          Questions? Comments? Feedback? Feel free to whisper @mistaeksweremade
          on Twitch.
        </p>
      </div>
    </div>
  );
}

function ActiveChannelList(props) {
  const channels = props.channels;
  if (!channels || channels.length === 0) {
    return (
      <div className="alert alert-primary">
        We're sorry, there doesn't appear to be any active solves right now.
      </div>
    );
  }

  const puzzle = props.puzzle;
  const links = [];
  for (let i = 0; i < channels.length; i++) {
    links.push(
      <a className="list-group-item list-group-item-action" href={`/${channels[i]}/${puzzle}`} key={channels[i]}>
        {channels[i]}
      </a>
    );
  }

  return (
    <div className="list-group mb-3">
      {links}
    </div>
  );
}

function ChannelHome(props) {
  const channel = props.channel;

  return (
    <div>
      <Nav />
      <div className="jumbotron">
        <h2>Solve a crossword</h2>
        <p>
          Cooperatively solve a crossword puzzle from the New York Times, Wall
          Street Journal, or another source using an uploaded .puz file.  Create
          your own rules by deciding whether or not to allow incorrect answers
          in the puzzle grid or to hide some or all of the clues.
        </p>
        <a className="btn btn-primary btn-lg" href={`${channel}/crossword`} role="button">
          Start Solving
        </a>
        <hr className="my-4"/>
        <h2>Solve a spelling bee</h2>
        <p>
          Cooperatively solve a spelling bee puzzle from the New York Times.
          See how many words of length 4 or greater you and your chat can
          discover from a collection of 7 letters.  Remember all words must use
          the center letter.  Create your own rules by deciding whether or not
          to only allow the words from the official New York Times dictionary or
          to use a less restrictive dictionary.
        </p>
        <a className="btn btn-primary btn-lg" href={`${channel}/spellingbee`} role="button">
          Start Solving
        </a>
      </div>
    </div>
  )
}

function ChannelRedirect(props) {
  const channel = props.channel;
  const view = props.view;

  return (
    <div>
      <Nav/>
      <div className="jumbotron">
        <h2>URL paths have changed!</h2>
        <p>
          In order to accommodate multiple types of puzzles being solved the
          URLs of this application have changed slightly. The URL you are
          looking for has moved to:
        </p>
        <div className="alert alert-dark">
          <a href={`/${channel}/crossword/${view}`}>
            {document.location.origin}/{channel}/crossword/{view}
          </a>
        </div>
        <p>
          Please update any bookmarks you may have to reflect the new URL.
        </p>
      </div>
    </div>
  );
}
