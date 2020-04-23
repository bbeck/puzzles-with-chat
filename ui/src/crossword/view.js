import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import "crossword/view.css";

export default function CrosswordView(props) {
  const state = props.state;
  const puzzle = state && state.puzzle;
  if (!state || !puzzle) {
    return (
      <h1>Crossword goes here</h1>
    );
  }

  const status = state.status;
  const last_start_time = state.last_start_time;
  const total_solve_duration = state.total_solve_duration;
  const settings = props.settings;
  const view = props.view;

  return (
    <div id="crossword" className={status === "selected" || status === "paused" ? "blur" : ""} data-size={Math.max(puzzle.cols, puzzle.rows)}>
      <div className="puzzle">
        <Header
          title={puzzle.title}
          author={puzzle.author}
          date={puzzle.published}
          last_start_time={last_start_time}
          total_solve_duration={total_solve_duration}
        />
        <Grid puzzle={puzzle} cells={state.cells} view={view}/>
        <Footer/>
      </div>
      <Clues
        across_clues={puzzle.clues_across}
        across_clues_filled={state.across_clues_filled}
        down_clues={puzzle.clues_down}
        down_clues_filled={state.down_clues_filled}
        notes={puzzle.notes}
        clue_font_size={settings.clue_font_size}
        clues_to_show={settings.clues_to_show}
        show_notes={settings.show_notes}
      />
    </div>
  );
}

function Header(props) {
  // Format an ISO-8601 datetime string as an ISO-8601 date string.
  const formatDate = (s) => {
    const date = s.split("T")[0];
    const [year, month, day] = date.split("-");
    return year + "-" + month + "-" + day;
  };

  // The .puz file format doesn't contain a field for the puzzle's publish date,
  // so when we load a .puz file the date comes across as 0001-01-01.  Detect
  // when this happens and when it does don't include the date in the header.
  let date;
  if (props.date !== "0001-01-01T00:00:00Z") {
    date = <div className="date">{formatDate(props.date)}</div>;
  }

  return (
    <div className="header">
      <div className="title" title={props.title} dangerouslySetInnerHTML={{__html: props.title}}/>
      <div className="author" title={props.author}>by {props.author}</div>
      {date}
      <Timer last_start_time={props.last_start_time} total_solve_duration={props.total_solve_duration}/>
    </div>
  );
}

function Timer(props) {
  // Parse the duration string (1h10m3s) into the total number of seconds.
  const total_solve_duration = ((duration) => {
    let re = /(?:(?<h>[0-9]+)h)?(?:(?<m>[0-9]+)m)?(?:(?<s>[0-9.]+)s)?/;
    let match = re.exec(duration);

    return (parseInt(match.groups.h || 0, 10)) * 3600 +
      (parseInt(match.groups.m || 0, 10)) * 60 +
      Math.round(parseFloat(match.groups.s || 0));
  })(props.total_solve_duration);

  // Convert the last start time into a timestamp as a number of seconds since
  // the epoch.  If there isn't a last start time, then this will return NaN.
  const last_start_time = Date.parse(props.last_start_time) / 1000;

  // Given an amount of time that the solve has gone for in the past as well as
  // time time that the current segment was started at, compute the total
  // duration in seconds that the solve has been going for.
  const compute = (total, start) => {
    if (!isNaN(start)) {
      total += new Date().getTime() / 1000 - start;
    }
    return Math.round(total);
  };

  const [total, setTotal] = React.useState(
    compute(total_solve_duration, last_start_time)
  );

  React.useEffect(() => {
    const interval = setInterval(() => {
      const total = compute(total_solve_duration, last_start_time);
      setTotal(total);
    }, 500);
    return () => clearInterval(interval)
  }, [total_solve_duration, last_start_time, setTotal]);

  return (
    <div className="timer">
      <Duration total={total}/>
    </div>
  );
}

function Duration(props) {
  const pad = (n) => {
    return (n < 10) ? "0" + n : n;
  };

  const total = props.total;
  const hours = Math.floor(total / 3600);
  const minutes = pad(Math.floor(total % 3600 / 60));
  const seconds = pad(Math.floor(total % 60));

  return (
    <React.Fragment>
      {`${hours}h ${minutes}m ${seconds}s`}
    </React.Fragment>
  );
}

function Footer() {
  return (
    <div className="footer">
      <div>Answer a clue: <code>!12a red velvet cake</code></div>
      <div>Partially answer a clue: <code>!12a gr.y goose</code></div>
      <div>Answer with a rebus: <code>!12a (gray)goose</code></div>
      <div>Make a clue visible: <code>!show 10d</code></div>
    </div>
  );
}

function Grid(props) {
  const puzzle = props.puzzle;
  const cells = props.cells;
  const view = props.view;
  if (!puzzle) {
    return (
      <div>Puzzle grid goes here.</div>
    );
  }

  const rows = [];
  for (let y = 0; y < puzzle.rows; y++) {
    const cols = [];
    for (let x = 0; x < puzzle.cols; x++) {
      const number = puzzle.cell_clue_numbers[y][x] || "";
      const content = cells[y][x] || "";
      const isBlock = puzzle.cell_blocks[y][x];
      const isCircle = puzzle.cell_circles[y][x];
      const isFilled = view === "progress" && content !== "";

      cols.push(
        <td key={y*puzzle.cols + x}>
          <div className={isBlock ? "cell block" : isFilled ? "cell filled" : isCircle ? "cell shaded" : "cell"}>
            <div className="number">{number}</div>
            <div className="content" data-length={content.length || null}>{view !== "progress" ? content : ""}</div>
          </div>
        </td>
      );
    }

    rows.push(
      <tr key={y}>
        {cols}
      </tr>
    );
  }
  return (
    <div className="grid">
      <table>
        <tbody>
          {rows}
        </tbody>
      </table>
    </div>
  );
}

function Clues(props) {
  const across_clues = props.across_clues;
  const across_clues_filled = props.across_clues_filled;
  const down_clues = props.down_clues;
  const down_clues_filled = props.down_clues_filled;
  const clue_notes = props.notes || "";
  const clues_to_show = props.clues_to_show;
  const clue_font_size = props.clue_font_size;
  const show_notes = props.show_notes;

  let across;
  if (clues_to_show === "all" || clues_to_show === "across") {
    across = <div className="across">
      <div className="clue-title">Across</div>
      <div id="across-clues" className="clue-list">
        <ClueList clues={across_clues} filled={across_clues_filled} side="a"/>
      </div>
    </div>;
  }

  let down;
  if (clues_to_show === "all" || clues_to_show === "down") {
    down = <div className="down">
      <div className="clue-title">Down</div>
      <div id="down-clues" className="clue-list">
        <ClueList clues={down_clues} filled={down_clues_filled} side="d"/>
      </div>
    </div>;
  }

  let notes;
  if (clues_to_show !== "none") {
    // If there are notes and we want to display them render them into the notes
    // element.
    if (clue_notes !== "" && show_notes) {
      notes = (
        <div
          id="clue-notes"
          className="notes"
          dangerouslySetInnerHTML={{__html: clue_notes}}
        />
      );
    }

    // If there are notes, but we don't want to display them, then show an
    // informational message instead.
    if (clue_notes !== "" && !show_notes) {
      notes = (
        <div id="clue-notes" className="notes">
          <p>
            This puzzle contains notes, please consider enabling the 'Show
            notes' setting.  <em>WARNING: Notes may contain spoilers.</em>
          </p>
        </div>
      );
    }
  }

  return (
    <div className="clues" data-font-size={clue_font_size}>
      <div className="content">
        {across}
        {down}
      </div>
      {notes}
    </div>
  );
}

function ClueList(props) {
  const clues = props.clues;
  if (!clues) {
    return (
      <div>Clues go here</div>
    );
  }

  const side = props.side;
  const filled = props.filled || {};

  // Make sure to always list the clues in sorted order.
  const numbers = Object.keys(clues);
  numbers.sort(function (a, b) {
    const ia = parseInt(a);
    const ib = parseInt(b);
    return (ia < ib) ? -1 : (ia === ib) ? 0 : 1;
  });

  const items = [];
  for (const number of numbers) {
    items.push(
      <li id={number + side} className={filled[number] ? "filled" : ""} key={number}>
        <span className="number">{number}</span>
        <span className="clue" dangerouslySetInnerHTML={{__html: clues[number]}}/>
      </li>
    );
  }

  return (
    <ul>
      {items}
    </ul>
  );
}
