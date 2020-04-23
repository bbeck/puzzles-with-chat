import React from "react";
import formatISO from "date-fns/formatISO";
import {DateChooser, Switch} from "nav";
import {
  isNewYorkTimesDateAllowed,
  isWallStreetJournalDateAllowed,
  nytFirstPuzzleDate,
  wsjFirstPuzzleDate
} from "crossword/allowed-dates";
import "crossword/nav.css";

export function StartPauseButton(props) {
  const status = props.status;

  let message;
  if (status === "selected") {
    message = "Start";
  } else if (status === "paused") {
    message = "Unpause";
  } else if (status === "solving") {
    message = "Pause";
  } else if (status === undefined || status === "created" || status === "complete") {
    return null;
  }

  // Toggle the status.
  function toggle() {
    fetch(`/api/crossword/${props.channel}/status`,
      {
        method: "PUT",
      });
  }

  return (
    <form className="form-inline nav-item">
      <button className="btn btn-success" type="button" onClick={toggle}>{message}</button>
    </form>
  );
}

export function ViewsDropdown(props) {
  const base = `${document.location.origin}/${props.channel}/crossword`;
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

      fetch(`/api/crossword/${props.channel}/setting/${name}`,
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
            <div className="lead">Only correct answers</div>
            <div>
              <small className="text-muted">
                This setting enables the behavior that only correct answers or
                correct partial answers will be accepted. With this enabled a
                cell in the puzzle can never contain an incorrect value.
              </small>
            </div>
            <Switch checked={settings.only_allow_correct_answers} onClick={update("only_allow_correct_answers", !settings.only_allow_correct_answers)}/>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Reduced set of clues</div>
            <div>
              <small className="text-muted">
                This setting enables some of the clues to be hidden thus making
                the puzzle significantly more difficult to solve.
              </small>
            </div>
            <div className="btn-group" role="group">
              <button type="button" className={settings.clues_to_show === "all" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "all")}>Show All</button>
              <button type="button" className={settings.clues_to_show === "down" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "down")}>Hide Across</button>
              <button type="button" className={settings.clues_to_show === "across" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "across")}>Hide Down</button>
              <button type="button" className={settings.clues_to_show === "none" ? "btn btn-success" : "btn btn-dark"} onClick={update("clues_to_show", "none")}>Hide All</button>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Clue font size</div>
            <div>
              <small className="text-muted">
                This setting allows the size of the font used to render clues to
                be adjusted.
              </small>
            </div>
            <div className="btn-group" role="group">
              <button type="button" className={settings.clue_font_size === "normal" ? "btn btn-success" : "btn btn-dark"} onClick={update("clue_font_size", "normal")}>Normal</button>
              <button type="button" className={settings.clue_font_size === "large" ? "btn btn-success" : "btn btn-dark"} onClick={update("clue_font_size", "large")}>Large</button>
              <button type="button" className={settings.clue_font_size === "xlarge" ? "btn btn-success" : "btn btn-dark"} onClick={update("clue_font_size", "xlarge")}>Extra Large</button>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Show notes</div>
            <div>
              <small className="text-muted">
                This setting enables showing the notes about the puzzle when
                they are present.  The notes are usually notes from the puzzle's
                constructor, but there are also sometimes notes from a solver
                that may contain spoilers.  Be careful when enabling this
                setting.
              </small>
            </div>
            <Switch checked={settings.show_notes} onClick={update("show_notes", !settings.show_notes)}/>
          </div>
        </form>
      </div>
    </li>
  );
}

export function PuzzleDropdown(props) {
  // Select a puzzle for the channel.  If the puzzle fails to load properly
  // then a simple error message will be displayed until a page reload or a
  // successful puzzle load.
  const setPuzzle = (payload) => {
    return fetch(`/api/crossword/${props.channel}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      })
      .then(response => {
        if (!response.ok) {
          throw new Error("Unable to load puzzle.")
        }

        props.setErrorMessage(null);
      })
      .then(() => {
        // Hide the dropdown menu after a selection is successfully made.
        const menu = document.getElementById("puzzle-dropdown-menu");
        if (menu) {
          menu.classList.remove("show");
        }
      })
      .catch(error => props.setErrorMessage(error.message));
  };

  // Helper to read a file or blob to its bytes.
  const read = (f) => new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = reject;
    reader.readAsBinaryString(f);
  });

  // Select a New York Times puzzle for a specific date.
  const onNewYorkTimesDateSelected = (date) => {
    if (!date) {
      return;
    }

    date = formatISO(date, {representation: "date"});
    return setPuzzle({"new_york_times_date": date});
  };

  // Select a Wall Street Journal puzzle for a specific date.
  const onWallStreetJournalDateSelected = (date) => {
    if (!date) {
      return;
    }

    date = formatISO(date, {representation: "date"});
    return setPuzzle({"wall_street_journal_date": date});
  };

  // Select a .puz puzzle based on a URL to the .puz file.
  const onPuzUrlSelected = (url) => {
    if (!url) {
      return;
    }

    return setPuzzle({"puz_file_url": url});
  };

  // Select a .puz puzzle based on the uploaded .puz file.
  const onPuzFileSelected = (file) => {
    if (!file) {
      return;
    }

    return read(file)
      .then(btoa)
      .then(bs => setPuzzle({"puz_file_bytes": bs}));
  };

  return (
    <li className="nav-item dropdown">
      <button type="button" className="btn btn-dark dropdown-toggle" data-toggle="dropdown">
        Puzzle
      </button>
      <div id="puzzle-dropdown-menu" className="dropdown-menu dropdown-menu-right" aria-labelledby="puzzle-dropdown">
        <form>
          <div className="dropdown-item">
            <div className="lead">New York Times</div>
            <div>
              <small className="text-muted">
                Select a date to solve that day's puzzle from the archives of
                the New York Times.
              </small>
            </div>
            <div className="input-group">
              <DateChooser
                onClick={onNewYorkTimesDateSelected}
                filterDate={isNewYorkTimesDateAllowed}
                minDate={nytFirstPuzzleDate}
              />
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Wall Street Journal</div>
            <div>
              <small className="text-muted">
                Select a date to solve that day's puzzle from the archives of
                the Wall Street Journal.
              </small>
            </div>
            <div className="input-group">
              <DateChooser
                onClick={onWallStreetJournalDateSelected}
                filterDate={isWallStreetJournalDateAllowed}
                minDate={wsjFirstPuzzleDate}
              />
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Download a .puz file</div>
            <div>
              <small className="text-muted">
                Input a URL to a .puz file to solve an existing puzzle hosted
                somewhere online.
              </small>
            </div>
            <div className="input-group">
              <input id="puz-url-input" type="url" className="form-control" />
              <div className="input-group-append">
                <label htmlFor="puz-url-input" className="btn btn-dark" onClick={e => onPuzUrlSelected(e.target.control.value)}>Load</label>
              </div>
            </div>
          </div>
          <div className="dropdown-divider"/>
          <div className="dropdown-item">
            <div className="lead">Upload a .puz file</div>
            <div>
              <small className="text-muted">
                Upload your own puzzle file in .puz format. More info about
                the .puz file format can be found&nbsp;
                <a href="http://fileformats.archiveteam.org/wiki/PUZ_(crossword_puzzles)">here</a>.
              </small>
            </div>
            <div className="input-group">
              <input id="puz-file-input" type="file" accept=".puz" className="d-none" onChange={e => onPuzFileSelected(e.target.files[0])}/>
              <label htmlFor="puz-file-input" className="btn btn-dark" onClick={e => {e.target.control.value = null}}>Choose file</label>
            </div>
          </div>
        </form>
      </div>
    </li>
  );
}