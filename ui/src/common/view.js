import React from "react";

export function TwitchChat({channel}) {
  const parent = document.location.hostname;

  return (
    <iframe
      className="chat"
      title="chat"
      frameBorder="0"
      src={`https://www.twitch.tv/embed/${channel}/chat?parent=${parent}`}
      scrolling="yes"
    />
  );
}

export function Timer({total_solve_duration, last_start_time}) {
  // Parse the duration into the total number of seconds that the solve has
  // accumulated prior to this most recent start.
  const prior = parseDuration(total_solve_duration);

  // Parse the last start time into the number of seconds after the epoch that
  // we started the solving segment.  If there isn't a last start time then this
  // will be NaN.
  const started = Date.parse(last_start_time) / 1000;

  // Render a duration that will self-update showing the number of seconds that
  // the solve has been progressing for.  We do this separately from this
  // component to ensure that we don't re-parse each time the duration updates.
  return (<Duration prior={prior} started={started}/>);
}

function Duration({prior, started}) {
  const [duration, setDuration] = React.useState("");

  // Repeatedly call a callback to recompute the number of seconds that the
  // solve has been running for.
  useInterval(() => {
    const delta = !isNaN(started)
      ? new Date().getTime() / 1000 - started
      : 0;

    const total = Math.max(Math.round(prior + delta), 0);
    const hours = Math.floor(total / 3600);
    const minutes = Math.floor(total % 3600 / 60);
    const seconds = Math.floor(total % 60);

    setDuration(`${hours}h ${pad(minutes)}m ${pad(seconds)}s`);
  }, 500);

  return <div className="timer">{duration}</div>;
}

// Parse the provided duration string (e.g. 1h10m3s) into the total number of
// seconds that the duration contains.
function parseDuration(duration) {
  const re = /(?:(?<h>[0-9]+)h)?(?:(?<m>[0-9]+)m)?(?:(?<s>[0-9.]+)s)?/;
  const match = re.exec(duration);

  return (parseInt(match.groups.h || 0, 10)) * 3600 +
    (parseInt(match.groups.m || 0, 10)) * 60 +
    Math.round(parseFloat(match.groups.s || 0));
}

// Pad a number to 2 digits.
function pad(n) {
  return (n < 10) ? "0" + n : n;
}

function useInterval(callback, delay) {
  const savedCallback = React.useRef(callback);

  React.useEffect(
    () => {
      savedCallback.current = callback;
    },
    [callback]
  );

  React.useEffect(
    () => {
      const handler = (...args) => savedCallback.current(...args);

      if (delay !== null) {
        const id = setInterval(handler, delay);
        return () => clearInterval(id);
      }
    },
    [delay]
  );
}