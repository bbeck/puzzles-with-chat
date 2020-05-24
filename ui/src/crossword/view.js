import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import {Timer} from "common/view";
import "crossword/view.css";

export function CrosswordView(props) {
  const state = props.state;
  const puzzle = state && state.puzzle;
  if (!state || !puzzle) {
    return (
      <div className="jumbotron">
        <h1>Welcome to Crosswords!</h1>
        <hr className="my-4"/>
        <p>
          Please select a crossword puzzle to solve using the&nbsp;
          <b>Puzzle</b> menu on the top right of the screen.  You can select
          a crossword from the archives of the New York Times or Wall Street
          Journal, or upload your own URL or .puz file to solve one from another
          source.
        </p>
        <p>
          Once a puzzle is started Puzzles With Chat will allow a Twitch
          streamer's viewers to cooperatively solve a crossword puzzle by
          inputting answers into the chat.  As correct answers are inputted
          they will appear on screen within the grid.  Solve the full puzzle
          correctly for a fun congratulatory demonstration.
        </p>
        <div>
          Potential sources of crossword .puz files:
          <ul>
            <li><a href="https://www.fleetingimage.com/wij/xyzzy/nyt-links.html">Puzzle Pointers</a></li>
            <li><a href="https://crosswordfiend.com/download/">Crossword Fiend</a></li>
          </ul>

          Potential sources of cryptic crossword .puz files (keep in mind that
          cryptic crosswords use a different set of&nbsp;
          <a href="https://en.wikipedia.org/wiki/Cryptic_crossword">rules</a> for
          clues than regular crosswords):
          <ul>
            <li><a href="https://www.xwordinfo.com/SelectVariety">XWord Info</a> (other types of .puz files available as well)</li>
            <li><a href="https://www.fleetingimage.com/wij/xyzzy/cryptic-links.html">Cryptic Pointers</a></li>
            <li><a href="http://world.std.com/~wij/puzzles/cru/">Cru Cryptics</a></li>
          </ul>
        </div>
        <hr className="my-4"/>
        <p>
          Questions? Comments? Feedback? Feel free to whisper @mistaeksweremade
          on Twitch.
        </p>
      </div>
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
      <Timer
        last_start_time={props.last_start_time}
        total_solve_duration={props.total_solve_duration}
      />
    </div>
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
  const contents = props.cells;
  const view = props.view;

  // Because we're rendering as a SVG we'll make the size of each cell fixed
  // regardless of the width or height of the puzzle.  We'll then change the
  // view box of the SVG to contain the complete puzzle adding padding where
  // necessary and render it accordingly.

  // The side length of each cell in pixels -- needed for calculations.  This
  // value is arbitrary and shouldn't ever need to change.
  const s = 100;

  // Border thickness of each cell in pixels -- needed for calculations.  This
  // value should not change without also changing the CSS rule that defines the
  // stroke width on each cell.
  const b = 2;

  // Thickness of the border surrounding the puzzle in pixels.  Half of this
  // value encroaches into the interior of the cells so it shouldn't be made
  // bigger than the cell's border or else cell area is lost.
  const B = 2 * b;

  // Define the border rectangle.
  const border = (
    <rect x={-b / 2} y={-b / 2}
          width={s * puzzle.cols + b} height={s * puzzle.rows + b}
          className="border"
    />
  );

  const cells = [];
  for (let cy = 0; cy < puzzle.rows; cy++) {
    for (let cx = 0; cx < puzzle.cols; cx++) {
      const number = puzzle.cell_clue_numbers[cy][cx] || "";
      const content = contents[cy][cx] || "";
      const isBlock = puzzle.cell_blocks[cy][cx];
      const isCircle = puzzle.cell_circles[cy][cx];
      const isFilled = view === "progress" && content !== "";
      const className = isBlock ? "cell block" : isFilled ? "cell filled" : isCircle ? "cell shaded" : "cell";

      const x = cx * s;
      const y = cy * s;
      cells.push(
        <g key={cy * puzzle.cols + cx}>
          <rect x={x} y={y} width={s} height={s} className={className}/>
          <text x={x} y={y} className="number">{number}</text>
          <text x={x} y={y} className="content" data-length={content.length}>
            {view !== "progress" ? content : ""}
          </text>
        </g>
      );
    }
  }

  let minX = -b / 2 - B/2;
  let minY = -b / 2 - B/2;
  let width = s * puzzle.cols + b + B;
  let height = s * puzzle.rows + b + B;

  if (puzzle.rows < puzzle.cols) {
    // There are extra cells horizontally, to compensate we need padding on the
    // top and bottom sides.
    const dy = width - height;
    minY -= dy / 2;
    height += dy;
  }

  if (puzzle.rows > puzzle.cols) {
    // There are extra cells vertically, to compensate we need padding on the
    // left and right sides.
    const dx = height - width;
    minX -= dx / 2;
    width += dx;
  }

  return (
    <div className="grid">
      <svg viewBox={`${minX} ${minY} ${width} ${height}`}>
        {cells}
        {border}  {/* Add this after the cells so it's drawn on top. */}
      </svg>
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
