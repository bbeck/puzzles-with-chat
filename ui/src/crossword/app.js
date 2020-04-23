import React from "react";
import EventStream from "event-stream";
import Fireworks from "fireworks";
import Nav from "nav";
import {StartPauseButton, ViewsDropdown, SettingsDropdown, PuzzleDropdown} from "crossword/nav";
import CrosswordView from "crossword/view";

export default function CrosswordApp(props) {
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

  return (
    <>
      <Nav view={props.view} error={error}>
        <ul className="navbar-nav ml-auto">
          <StartPauseButton channel={props.channel} status={state.status}/>
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
