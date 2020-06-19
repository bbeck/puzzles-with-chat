import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import parseISO from "date-fns/parseISO";
import "./nav.css";

export function Nav(props) {
  // If we're not on the streamer view then don't show errors or any of the
  // child elements that allow the state of the application to be modified.
  if (!props.view || props.view !== "streamer") {
    return (
      <nav className="navbar navbar-expand navbar-dark text-light bg-dark">
        <BrandDropdown puzzle={props.puzzle}/>
      </nav>
    );
  }

  return (
    <nav className="navbar navbar-expand navbar-dark text-light bg-dark">
      <BrandDropdown puzzle={props.puzzle}/>

      {props.error &&
      <div className="navbar-text text-danger m-auto">
        <strong>{props.error}</strong>
      </div>
      }

      {props.children}
    </nav>
  );
}

function BrandDropdown(props) {
  if (!props.puzzle) {
    return (
      <div className="navbar-brand">Puzzles With Chat</div>
    );
  }

  // template is a URL with a {puzzle} token in it indicating where the puzzle
  // type belongs in the url.
  const template = document.location.href
    .replace("/acrostic/", "/{puzzle}/")
    .replace("/acrostic", "/{puzzle}")
    .replace("/crossword/", "/{puzzle}/")
    .replace("/crossword", "/{puzzle}")
    .replace("/spellingbee/", "/{puzzle}/")
    .replace("/spellingbee", "/{puzzle}");


  return (
    <div className="nav-item dropdown">
      <div className="navbar-brand dropdown-toggle" data-toggle="dropdown">
        {props.puzzle} With Chat
      </div>
      <div className="dropdown-menu">
        <a className="dropdown-item lead" href={template.replace("{puzzle}", "acrostic")}>Acrostics</a>
        <a className="dropdown-item lead" href={template.replace("{puzzle}", "crossword")}>Crosswords</a>
        <a className="dropdown-item lead" href={template.replace("{puzzle}", "spellingbee")}>Spelling Bee</a>
      </div>
    </div>
  );
}

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

  return (
    <form className="form-inline nav-item">
      <button className="btn btn-success" type="button" onClick={props.onClick}>{message}</button>
    </form>
  );
}

export function DateChooser(props) {
  const [selectedDate, setSelectedDate] = React.useState(null);

  return (
    <>
      <DatePicker
        className="form-control"
        placeholderText="mm/dd/yyyy"
        selected={selectedDate}
        onChange={setSelectedDate}
        filterDate={props.filterDate}
        maxDate={new Date()}
        minDate={props.minDate}
        showYearDropdown={true}
        dropdownMode="select"
      />
      <div className="input-group-append">
        <button type="button" className="btn btn-dark" onClick={() => props.onClick(selectedDate)}>Load</button>
      </div>
    </>
  );
}

export function Switch(props) {
  return (
    <label className="switch">
      <input type="checkbox" className="form-check-input success" checked={props.checked} readOnly/>
      <span className="slider round" onClick={props.onClick}/>
    </label>
  );
}

// Fetch the source dates for a given puzzle type and for each source compute
// the minimum and available dates and return them.
export function useSourceDates(puzzle, setErrorMessage) {
  const [minDates, setMinDates] = React.useState({});
  const [dates, setDates] = React.useState({});

  // Fetch the available dates from the API and then index them into the above
  // state variables.
  React.useEffect(() => {
    fetch(`/api/${puzzle}/dates`)
      .then(response => {
        if (!response.ok) {
          throw new Error("Unable to load available puzzle dates.");
        }

        return response.json();
      })
      .then(response => {
        const min = {}
        const dates = {}
        for (const [source, ds] of Object.entries(response)) {
          min[source] = parseISO(ds[0]);
          dates[source] = new Set(ds);
        }

        setMinDates(min);
        setDates(dates);
      })
      .catch(error => setErrorMessage(error.message));
  }, [puzzle, setErrorMessage, setMinDates, setDates]);

  return [minDates, dates];
}