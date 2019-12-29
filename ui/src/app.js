import React from "react";
import {Router} from "@reach/router";
import {Crossword} from "./crossword";
import {EventStream} from "./event-stream";
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

        {/*{{if .rooms}}*/}
        {/*<p>*/}
        {/*  <div className="list-group">*/}
        {/*    {{range .rooms}}*/}
        {/*    <a className="list-group-item list-group-item-action"*/}
        {/*       href="/{{.}}">{{.}}</a>*/}
        {/*    {{end}}*/}
        {/*  </div>*/}
        {/*</p>*/}
        {/*{{else}}*/}
        <div className="alert alert-primary">
          We're sorry, there doesn't appear to be any active solves right now.
        </div>
        {/*{{end}}*/}

        <p>
          Questions? Comments? Feedback? Feel free to whisper @mistaeksweremade
          on Twitch.
        </p>
      </div>
    </div>
  );
}

function Channel(props) {
  const [events, setEventStream] = React.useState(
    new EventStream(`/api/channel/${props.channel}/events`)
  );

  const [settings, setSettings] = React.useState({
    clues_to_show: "all",
    clue_font_size: "normal",
    only_allow_correct_answers: false,
  });

  const [state, setState] = React.useState({});

  React.useEffect(() => {
    events.setHandler(message => {
      const event = JSON.parse(message.data);

      if (event.kind === "settings") {
        setSettings(event.payload);
      }

      if (event.kind === "state") {
        setState(event.payload);
      }
    });
  }, [events, setEventStream, setSettings, setState]);


  return (
    <React.Fragment>
      <Nav channel={props.channel} view={props.view} settings={settings}/>
      <Crossword view={props.view} state={state}/>
    </React.Fragment>
  );
}

export default App;
