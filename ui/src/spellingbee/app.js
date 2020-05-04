import React from "react";
import {EventStream} from "common/event-stream";
import {Fireworks} from "common/fireworks";
import {Nav, StartPauseButton} from "common/nav";
import {PuzzleDropdown, SettingsDropdown, ViewsDropdown} from "spellingbee/nav";
import {SpellingBeeView} from "spellingbee/view";

export function SpellingBeeApp(props) {
  // The settings for the spelling bee app for the current channel.
  const [settings, setSettings] = React.useState({
    allow_unofficial_answers: false,
    font_size: "normal"
  });

  // The current state of the spelling bee app for the current channel.
  const [state, setState] = React.useState({});

  // Whether or not we're currently showing fireworks.
  const [showFireworks, setShowFireworks] = React.useState(false);

  // Events for the crossword being solved by the channel.
  const [stream] = React.useState(
    new EventStream(`/api/spellingbee/${props.channel}/events`)
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
    return fetch(`/api/spellingbee/${props.channel}/status`, {method: "PUT"});
  }

  return (
    <>
      <Nav puzzle="Spelling Bee" view={props.view} error={error}>
        <ul className="navbar-nav ml-auto">
          <StartPauseButton puzzle="spellingbee" status={state.status} onClick={toggleStatus}/>
          <ViewsDropdown channel={props.channel}/>
          <SettingsDropdown channel={props.channel} settings={settings}/>
          <PuzzleDropdown channel={props.channel} setErrorMessage={setError}/>
        </ul>
      </Nav>

      <SpellingBeeView view={props.view} state={state} settings={settings}/>
      {showFireworks && <Fireworks/>}
    </>
  );
}
