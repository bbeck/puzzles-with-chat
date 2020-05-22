import React from "react";
import {EventStream} from "common/event-stream";
import {Fireworks} from "common/fireworks";
import {Nav, StartPauseButton} from "common/nav";
import {PuzzleDropdown, SettingsDropdown, ViewsDropdown} from "crossword/nav";
import {CrosswordView} from "crossword/view";

export function CrosswordApp(props) {
  // The settings for the crossword app for the current channel.
  const [settings, setSettings] = React.useState({
    clues_to_show: "all",
    clue_font_size: "normal",
    only_allow_correct_answers: false,
    show_notes: false,
  });

  // The current state of the crossword app for the current channel.
  const [state, setState] = React.useState({});

  // Whether or not we're currently showing fireworks.
  const [showFireworks, setShowFireworks] = React.useState(false);

  // Events for the crossword being solved by the channel.
  const [stream] = React.useState(
    new EventStream(`/api/crossword/${props.channel}/events`)
  );

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

          if (event.payload.status === "selected") {
            // We just started a new puzzle -- scroll the clues back to the top
            // of the list.  We'll use the same hack that we use in the
            // show_clue handler below, namely to reach into the DOM and grab
            // the clue element based on what we know the element's ID should
            // be.
            for (let n = 1; n < event.payload.puzzle.cols; n++) {
              const clue = document.getElementById(`${n}a`);
              if (clue !== null) {
                clue.scrollIntoView();
                break;
              }
            }
            for (let n = 1; n < event.payload.puzzle.rows; n++) {
              const clue = document.getElementById(`${n}d`);
              if (clue !== null) {
                clue.scrollIntoView();
                break;
              }
            }
          }
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
  }, [setSettings, stream, setState, setShowFireworks]);

  // The error message we're currently displaying.
  const [error, setError] = React.useState(null);

  // Toggle the status.
  const toggleStatus = () => {
    return fetch(`/api/crossword/${props.channel}/status`, {method: "PUT"});
  }

  return (
    <>
      <Nav puzzle="Crosswords" view={props.view} error={error}>
        <ul className="navbar-nav ml-auto">
          <StartPauseButton puzzle="crossword" status={state.status} onClick={toggleStatus}/>
          <ViewsDropdown channel={props.channel}/>
          <SettingsDropdown channel={props.channel} settings={settings}/>
          <PuzzleDropdown channel={props.channel} setErrorMessage={setError}/>
        </ul>
      </Nav>

      <CrosswordView view={props.view} state={state} settings={settings}/>
      {showFireworks && <Fireworks/>}
    </>
  );
}
