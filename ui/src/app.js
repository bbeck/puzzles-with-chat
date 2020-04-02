import React from "react";
import {Router} from "@reach/router";
import {Crossword} from "./crossword";
import {EventStream} from "./event-stream";
import {Fireworks} from "./fireworks";
import {Nav} from "./nav";

function App() {
  return (
    <Router>
      <Home path="/"/>
      <Channel path="/:channel/"/>
      <Channel path="/:channel/:view"/>
    </Router>
  );
}

function Home() {
  const [stream] = React.useState(
    new EventStream(`/api/crossword/events`)
  );

  const [channels, setChannels] = React.useState(null);

  React.useEffect(() => {
    stream.setHandler(message => {
      const event = JSON.parse(message.data);
      switch (event.kind) {
        case "channels":
          setChannels(event.payload);
          break;

        case "ping":
          break;

        default:
          console.log("unhandled event:", event);
      }
    });
  }, [stream, setChannels]);

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
        <ActiveChannelList channels={channels}/>
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

  const links = [];
  for (let i = 0; i < channels.length; i++) {
    links.push(
      <a className="list-group-item list-group-item-action" href={`/${channels[i]}`} key={channels[i]}>
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

function Channel(props) {
  const [stream] = React.useState(
    new EventStream(`/api/crossword/${props.channel}/events`)
  );

  const [settings, setSettings] = React.useState({
    clues_to_show: "all",
    clue_font_size: "normal",
    only_allow_correct_answers: false,
    show_notes: false,
  });

  const [state, setState] = React.useState({});

  const [showFireworks, setShowFireworks] = React.useState(false);

  React.useEffect(() => {
    stream.setHandler(message => {
      const event = JSON.parse(message.data);
      switch(event.kind) {
        case "settings":
          setSettings(event.payload);
          break;

        case "state":
          // If we get a state update while watching the fireworks animation,
          // then finish the show and start the new puzzle.
          setShowFireworks(false);
          setState(event.payload);
          break;

        case "show_clue":
          // This is a bit of a hack since we just reach into the DOM to grab
          // the clue element, but this is just presentation logic and not
          // state, so trying to pull a reference to the clue element from deep
          // within the component hierarchy is quite complicated and much uglier
          // than this hack.
          const clue = document.getElementById(event.payload);
          if (clue !== null) {
            clue.scrollIntoView();
            clue.classList.add("shown");
            setTimeout(() => clue.classList.remove("shown"), 2500);
          }
          break;

        case "complete":
          setShowFireworks(true);
          setTimeout(() => setShowFireworks(false), 20000);
          break;

        case "ping":
          break;

        default:
          console.log("unhandled event:", event);
      }
    });
  }, [stream, setSettings, setState, setShowFireworks]);

  return (
    <React.Fragment>
      <Nav channel={props.channel} view={props.view} settings={settings} status={state.status}/>
      <Crossword view={props.view} state={state} settings={settings}/>
      {showFireworks && <Fireworks/>}
    </React.Fragment>
  );
}

export default App;
