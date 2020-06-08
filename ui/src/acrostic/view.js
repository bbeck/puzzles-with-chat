import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import "acrostic/view.css";

export function AcrosticView(props) {
  const state = props.state;
  const puzzle = state && state.puzzle;
  if (!state || !puzzle) {
    return (
      <div className="jumbotron">
        <h1>Welcome to Acrostics!</h1>
        <hr className="my-4"/>
        <p>
          Please select an acrostic puzzle to solve using the&nbsp;
          <b>Puzzle</b> menu on the top right of the screen.
        </p>
        <p>
          Once a puzzle is started Puzzles With Chat will allow a Twitch
          streamer's viewers to cooperatively solve an acrostic puzzle by
          inputting answers into the chat.  As correct answers are inputted
          they will appear on screen within the grid.  Solve the full puzzle
          correctly for a fun congratulatory demonstration.
        </p>
        <hr className="my-4"/>
        <p>
          Questions? Comments? Feedback? Feel free to whisper @mistaeksweremade
          on Twitch.
        </p>
      </div>
    );
  }

  const status = state.status;

  return (
    <div id="acrostic" className={status === "selected" || status === "paused" ? "blur" : ""}>
      <div className="puzzle">
        <Header />
        <Grid/>
        <Footer/>
      </div>
      <Clues/>
    </div>
  );
}

function Header() {
  return (<div>Header goes here</div>);
}

function Footer() {
  return (<div>Footer goes here</div>);
}

function Grid() {
  return (<div>Grid goes here</div>);
}

function Clues() {
  return (<div>Clues go here</div>);
}