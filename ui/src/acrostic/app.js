import React from "react";
import {EventStream} from "common/event-stream";
import {Fireworks} from "common/fireworks";
import {Nav, StartPauseButton} from "common/nav";
import {PuzzleDropdown, SettingsDropdown, ViewsDropdown} from "acrostic/nav";
import {AcrosticView} from "acrostic/view";

export function AcrosticApp(props) {
  // The settings for the acrostic app for the current channel.
  const [settings, setSettings] = React.useState({
    clue_font_size: "normal",
    only_allow_correct_answers: false,
  });

  // The current state of the app for the current channel.
  const [state, setState] = React.useState({});

  // Whether or not we're currently showing fireworks.
  const [showFireworks, setShowFireworks] = React.useState(false);

  // Events for the puzzle being solved by the channel.
  const [stream] = React.useState(
    new EventStream(`/api/acrostic/${props.channel}/events`)
  );

  // The error message we're currently displaying.
  const [error, setError] = React.useState(null);

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

          // If we were displaying an error before due to not being able to
          // select a puzzle then clear that now as well.  This is necessary
          // because another client may have selected a puzzle.
          setError(null);

          // Update the state.
          setState(event.payload);

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

  // Toggle the status.
  const toggleStatus = () => {
    return fetch(`/api/acrostic/${props.channel}/status`, {method: "PUT"});
  }

  return (
    <>
      <Nav puzzle="Acrostics" view={props.view} error={error}>
        <ul className="navbar-nav ml-auto">
          <StartPauseButton puzzle="acrostic" status={state.status} onClick={toggleStatus}/>
          <ViewsDropdown channel={props.channel}/>
          <SettingsDropdown channel={props.channel} settings={settings}/>
          <PuzzleDropdown channel={props.channel} setErrorMessage={setError}/>
        </ul>
      </Nav>

      <AcrosticView view={props.view} state={state} settings={settings}/>
      {showFireworks && <Fireworks/>}
    </>
  );
}
