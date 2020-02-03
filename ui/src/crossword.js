import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import "./crossword.css";

export function Crossword(props) {
  const state = props.state;
  const puzzle = state && state.puzzle;
  const status = state && state.status;
  const settings = props.settings;
  const view = props.view;

  if (!state || !puzzle) {
    return (
      <h1>Crossword goes here</h1>
    );
  }

  return (
    <div id="crossword" className={status === "created" || status === "paused" ? "blur" : ""} data-size={Math.max(puzzle.cols, puzzle.rows)}>
      <div className="puzzle">
        <Header title={puzzle.title} author={puzzle.author} date={puzzle.published}/>
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
      />
    </div>
  );
}

function Header(props) {
  const formatDate = (s) => {
    const date = s.split("T")[0];
    const [year, month, day] = date.split("-");
    return year + "-" + month + "-" + day;
  };

  // TODO: Enable the timer below
  return (
    <div className="header">
      <div className="title" dangerouslySetInnerHTML={{__html: props.title}}/>
      <div className="author">by {props.author}</div>
      <div className="date">{formatDate(props.date)}</div>
      <div className="timer">0h 00m 00s</div>
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
    notes = <div id="clue-notes" className="notes" dangerouslySetInnerHTML={{__html: clue_notes}}/>;
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
        <span className="clue">{clues[number]}</span>
      </li>
    );
  }

  return (
    <ul>
      {items}
    </ul>
  );
}