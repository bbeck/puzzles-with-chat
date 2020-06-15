import React from "react";
import {Router} from "@reach/router";
import {EventStream} from "common/event-stream";
import {Nav} from "common/nav";
import {AcrosticApp} from "acrostic/app";
import {CrosswordApp} from "crossword/app";
import {SpellingBeeApp} from "spellingbee/app";

export default function App() {
  return (
    <Router>
      <Home path="/"/>
      <ChannelHome path="/:channel/"/>

      <AcrosticApp path="/:channel/acrostic/"/>
      <AcrosticApp path="/:channel/acrostic/:view"/>

      <CrosswordApp path="/:channel/crossword/"/>
      <CrosswordApp path="/:channel/crossword/:view"/>

      <SpellingBeeApp path="/:channel/spellingbee/"/>
      <SpellingBeeApp path="/:channel/spellingbee/:view"/>
    </Router>
  );
}

function Home() {
  const [acrosticChannels, setAcrosticChannels] = React.useState(null);
  const [crosswordChannels, setCrosswordChannels] = React.useState(null);
  const [spellingBeeChannels, setSpellingBeeChannels] = React.useState(null);

  const [channels] = React.useState(new EventStream("/api/channels"));
  React.useEffect(() => {
    channels.setHandler(message => {
      const event = JSON.parse(message.data);
      switch (event.kind) {
        case "channels":
          setAcrosticChannels(event.payload["acrostic"]);
          setCrosswordChannels(event.payload["crossword"]);
          setSpellingBeeChannels(event.payload["spellingbee"]);
          break;

        case "ping":
          break;

        default:
          console.log("unhandled event:", event);
      }
    });
  }, [channels, setCrosswordChannels, setSpellingBeeChannels]);

  return (
    <div>
      <Nav/>
      <div className="jumbotron">
        <h1>Welcome to Puzzles With Chat!</h1>
        <hr className="my-4"/>
        <p>
          Puzzles With Chat is a web application that Twitch streamers can
          use to allow their chat to interactively solve puzzles together.  The
          streamer selects a puzzle to solve and then this application will have
          a chat bot join the streamer's chat. From there participants can input
          answers to the puzzle into the chat and the application will show them
          on screen. Currently both crossword puzzles and spelling bee puzzles
          are supported.
        </p>

        <p>
          If you've ended up on this page you were probably looking to spectate
          a solving session that is already in progress, but didn't get the full
          URL to the Twitch streamer's page. To help you below you'll find a
          list of all of the solving sessions that are in progress.  Please
          click through to the streamer you were looking for.
        </p>

        <h6>Channels with active acrostics:</h6>
        <ActiveChannelList channels={acrosticChannels} puzzle="acrostic"/>

        <h6>Channels with active crosswords:</h6>
        <ActiveChannelList channels={crosswordChannels} puzzle="crossword"/>

        <h6>Channels with active spelling bees:</h6>
        <ActiveChannelList channels={spellingBeeChannels} puzzle="spellingbee"/>

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
    const channel = channels[i];

    let status;
    switch (channel.status) {
      case "created":
      case "selected":
        status = "ABOUT TO START";
        break;
      case "paused":
        status = "PAUSED";
        break;
      case "solving":
        status = "SOLVING";
        break;
      case "complete":
        status = "FINISHED";
        break;
      default:
        status = "UNKNOWN";
        break;
    }

    const description = channel.description
      ? `${channel.description} (${status})`
      : status;

    links.push(
      <a className="list-group-item list-group-item-action" href={`/${channel.name}/${puzzle}`} key={channel.name}>
        {channel.name} - {description}
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
        <hr className="my-4"/>
        <h2>Solve an acrostic</h2>
        <p>
          Cooperatively solve an acrostic puzzle from the New York Times.
          Correctly answer clues to fill in parts of a hidden quote, or if you
          prefer fill in parts of the hidden quote to answer clues!  Create your
          own rules by deciding whether or not to allow incorrect answers in the
          puzzle grid.
        </p>
        <a className="btn btn-primary btn-lg" href={`${channel}/acrostic`} role="button">
          Start Solving
        </a>
      </div>
    </div>
  )
}
