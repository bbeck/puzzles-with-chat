import React from "react";
import {DateChooser, Switch, useSourceDates} from "common/nav";
import formatISO from "date-fns/formatISO";

export function ViewsDropdown(props) {
  const base = `${document.location.origin}/${props.channel}/spellingbee`;
  const copyToClipboard = (id) => {
    const input = document.getElementById(id);
    if (input !== null) {
      input.select();
      document.execCommand("copy");
      document.getSelection().empty();
    }
  };

  return (
    <li className="nav-item dropdown">
      <button type="button" className="btn btn-dark dropdown-toggle" data-toggle="dropdown">
        Views
      </button>
      <div id="views-dropdown-menu" className="dropdown-menu dropdown-menu-right" aria-labelledby="views-dropdown">
        <form>
          <div className="dropdown-item">
            <div className="lead">Streamer View</div>
            <div>
              <small className="text-muted">
                This link allows changing of settings as well as the active
                puzzle. Only share it with those you fully trust.
              </small>
            </div>
            <div className="input-group">
              <input id="streamer-url" type="text" className="form-control" value={`${base}/streamer`} readOnly/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark" onClick={() => copyToClipboard("streamer-url")}>Copy</button>
              </div>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Progress View</div>
            <div>
              <small className="text-muted">
                This link will only show the progress of the puzzle solve. As
                cells are populated with values they will change color, but the
                value put in the cells will not be visible. This link is
                intended to be shared with others who are solving the puzzle at
                the same time, but shouldn't see any answers.
              </small>
            </div>
            <div className="input-group">
              <input id="progress-url" type="text" className="form-control" value={`${base}/progress`} readOnly/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark" onClick={() => copyToClipboard("progress-url")}>Copy</button>
              </div>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Participant View</div>
            <div>
              <small className="text-muted">
                This link will show the complete progress of the solve as it
                happens in real-time. This link is safe to share with anyone.
              </small>
            </div>
            <div className="input-group">
              <input id="participant-url" type="text" className="form-control" value={`${base}`} readOnly/>
              <div className="input-group-append">
                <button type="button" className="btn btn-dark" onClick={() => copyToClipboard("participant-url")}>Copy</button>
              </div>
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}

export function SettingsDropdown(props) {
  const settings = props.settings;

  // Update a setting with its new value.
  function update(name, value) {
    return function() {
      if (settings[name] === value) {
        return;
      }

      fetch(`/api/spellingbee/${props.channel}/setting/${name}`,
        {
          method: "PUT",
          body: JSON.stringify(value),
        });
    };
  }

  return (
    <li className="nav-item dropdown">
      <button type="button" className="btn btn-dark dropdown-toggle" data-toggle="dropdown">
        Settings
      </button>
      <div id="settings-dropdown-menu" className="dropdown-menu dropdown-menu-right" aria-labelledby="settings-dropdown">
        <form>
          <div className="dropdown-item">
            <div className="lead">Allow unofficial answers</div>
            <div>
              <small className="text-muted">
                This setting enables answers to be accepted that are valid words
                but not found in the dictionary the New York Times used for this
                puzzle.

                NOTE: Some words have been excluded from the New York Times
                dictionary because they're considered offensive.
              </small>
            </div>
            <Switch checked={settings.allow_unofficial_answers} onClick={update("allow_unofficial_answers", !settings.allow_unofficial_answers)}/>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Show answer placeholders</div>
            <div>
              <small className="text-muted">
                This setting enables showing a slot for every possible answer.
                This is useful to help narrow down where undiscovered answers
                fit in the alphabetical list of answers.
              </small>
            </div>
            <Switch checked={settings.show_answer_placeholders} onClick={update("show_answer_placeholders", !settings.show_answer_placeholders)}/>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Font size</div>
            <div>
              <small className="text-muted">
                This setting allows the size of the font used to render the
                found words to be adjusted.
              </small>
            </div>
            <div className="btn-group" role="group">
              <button type="button" className={settings.font_size === "normal" ? "btn btn-success" : "btn btn-dark"} onClick={update("font_size", "normal")}>Normal</button>
              <button type="button" className={settings.font_size === "large" ? "btn btn-success" : "btn btn-dark"} onClick={update("font_size", "large")}>Large</button>
              <button type="button" className={settings.font_size === "xlarge" ? "btn btn-success" : "btn btn-dark"} onClick={update("font_size", "xlarge")}>Extra Large</button>
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}

export function PuzzleDropdown({channel, setErrorMessage}) {
  const [minDates, dates] = useSourceDates("spellingbee", setErrorMessage);

  // Select a puzzle for the channel.  If the puzzle fails to load properly
  // then a simple error message will be displayed until a page reload or a
  // successful puzzle load.
  const setPuzzle = (source, date) => {
    if (!date) {
      return;
    }
    date = formatISO(date, {representation: "date"});

    return fetch(`/api/spellingbee/${channel}`,
      {
        method: "PUT",
        body: JSON.stringify({[source + "_date"]: date}),
      })
      .then(response => {
        if (!response.ok) {
          throw new Error("Unable to load puzzle.")
        }

        setErrorMessage(null);
      })
      .then(() => {
        // Hide the dropdown menu after a selection is successfully made.
        const menu = document.getElementById("puzzle-dropdown-menu");
        if (menu) {
          menu.classList.remove("show");
        }
      })
      .catch(error => setErrorMessage(error.message));
  };

  // Determine if a puzzle is available for a particular date.
  const isPuzzleAvailableForDate = (source, date) => {
    date = formatISO(date, {representation: "date"});
    return dates[source] && dates[source].has(date);
  };

  return (
    <li className="nav-item dropdown">
      <button type="button" className="btn btn-dark dropdown-toggle" data-toggle="dropdown">
        Puzzle
      </button>
      <div id="puzzle-dropdown-menu" className="dropdown-menu dropdown-menu-right" aria-labelledby="puzzle-dropdown">
        <form>
          <div className="dropdown-item">
            <div className="lead">New York Times (via NYTBee)</div>
            <div>
              <small className="text-muted">
                Select a date to solve that day's puzzle from the archives of
                the New York Times.
              </small>
            </div>
            <div className="input-group">
              <DateChooser
                onClick={(date) => setPuzzle("new_york_times", date)}
                filterDate={(date) => isPuzzleAvailableForDate("new_york_times", date)}
                minDate={minDates["new_york_times"]}
              />
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}