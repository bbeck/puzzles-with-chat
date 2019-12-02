import React from "react";
import {Router} from "@reach/router";
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

        <div className="alert alert-primary">
          We're sorry, there doesn't appear to be any active solves right now.
        </div>

        <p>
          Questions? Comments? Feedback? Feel free to whisper @mistaeksweremade
          on Twitch.
        </p>
      </div>
    </div>
  );
}

function Channel(props) {
  return (
    <React.Fragment>
      <Nav channel={props.channel} view={props.view}/>
      <h1>Channel view</h1>
    </React.Fragment>
  );
}

export default App;
