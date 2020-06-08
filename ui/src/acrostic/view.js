import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import "acrostic/view.css";
import {Timer} from "../common/view";

export function AcrosticView(props) {
  const state = props.state;
  const puzzle = state && state.puzzle;
  if (!state || !puzzle) {
    return (
      <div className="jumbotron">
        <h1>Welcome to Acrostics!</h1>
        <hr className="my-4"/>
        <p>
          Please select an acrostic puzzle to solve using the&nbsp;
          <b>Puzzle</b> menu on the top right of the screen.
        </p>
        <p>
          Once a puzzle is started Puzzles With Chat will allow a Twitch
          streamer's viewers to cooperatively solve an acrostic puzzle by
          inputting answers into the chat.  As correct answers are inputted
          they will appear on screen within the grid.  Solve the full puzzle
          correctly for a fun congratulatory demonstration.
        </p>
        <hr className="my-4"/>
        <p>
          Questions? Comments? Feedback? Feel free to whisper @mistaeksweremade
          on Twitch.
        </p>
      </div>
    );
  }

  const status = state.status;
  const view = props.view;

  return (
    <div id="acrostic" className={status === "selected" || status === "paused" ? "blur" : ""}>
      <div className="puzzle">
        <Header
          description={puzzle.description}
          last_start_time={state.last_start_time}
          total_solve_duration={state.total_solve_duration}
        />
        <Grid puzzle={puzzle} cells={state.cells} view={view}/>
        <Clues/>
        <Footer/>
      </div>
    </div>
  );
}

function Header({description, last_start_time, total_solve_duration}) {
  return (
    <div className="header">
      <div className="description">{description}</div>
      <Timer
        last_start_time={last_start_time}
        total_solve_duration={total_solve_duration}
      />
    </div>
  );
}

function Footer() {
  return (<div className="footer">Footer goes here</div>);
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
      const letter = puzzle.cell_clue_letters[cy][cx] || "";
      const content = contents[cy][cx] || "";
      const isBlock = puzzle.cell_blocks[cy][cx];
      const isFilled = view === "progress" && content !== "";
      const className = isBlock ? "cell block" : isFilled ? "cell filled" : "cell";
      const x = cx * s;
      const y = cy * s;

      cells.push(
        <g key={cy * puzzle.cols + cx}>
          <rect x={x} y={y} width={s} height={s} className={className}/>
          <text x={x} y={y} className="number">{number}</text>
          <text x={x+s} y={y} className="letter">{letter}</text>
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

  return (
    <div className="grid">
      <svg className="grid" viewBox={`${minX} ${minY} ${width} ${height}`}>
        {cells}
        {border}  {/* Add this after the cells so it's drawn on top. */}
      </svg>
    </div>
  );
}

function Clues() {
  return (<div>Clues go here</div>);
}