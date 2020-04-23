import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import "nav.css";

export default function Nav(props) {
  // If we're not on the streamer view then don't show erorrs or any of the
  // child elements that allow the state of the application to be modified.
  if (!props.view || props.view !== "streamer") {
    return (
      <nav className="navbar navbar-expand navbar-dark text-light bg-dark">
        <div className="navbar-brand">Twitch Plays Crosswords</div>
      </nav>
    );
  }

  return (
    <nav className="navbar navbar-expand navbar-dark text-light bg-dark">
      <div className="navbar-brand">Twitch Plays Crosswords</div>

      {props.error &&
      <div className="navbar-text text-danger m-auto">
        <strong>{props.error}</strong>
      </div>
      }

      {props.children}
    </nav>
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
