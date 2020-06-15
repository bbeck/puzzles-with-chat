import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import "acrostic/view.css";
import {Timer} from "../common/view";

export function AcrosticView({state, settings, view}) {
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

  return (
    <div id="acrostic" className={status === "selected" || status === "paused" ? "blur" : ""}>
      <div className="puzzle">
        <Header
          description={puzzle.description}
          last_start_time={state.last_start_time}
          total_solve_duration={state.total_solve_duration}
        />
        <Grid puzzle={puzzle} cells={state.cells} view={view}/>
        <Clues
          puzzle={puzzle}
          cells={state.cells}
          filled={state.clues_filled}
          font_size={settings.clue_font_size}
          view={view}
        />
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
  return (
    <div className="footer">
      <div>Answer a clue: <code>!Q half step</code></div>
      <div>Partially answer a clue: <code>!M sym..ony</code></div>
      <div>Fill grid cells: <code>!26 vast knowledge</code></div>
      <div>Make a clue visible: <code>!show H</code></div>
    </div>
  );
}

function Grid({puzzle, cells, view}) {
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

  const boxes = [];
  for (let cy = 0; cy < puzzle.rows; cy++) {
    for (let cx = 0; cx < puzzle.cols; cx++) {
      const number = puzzle.cell_clue_numbers[cy][cx] || "";
      const letter = puzzle.cell_clue_letters[cy][cx] || "";
      const content = cells[cy][cx] || "";
      const isBlock = puzzle.cell_blocks[cy][cx];
      const isFilled = view === "progress" && content !== "";
      const className = isBlock ? "cell block" : isFilled ? "cell filled" : "cell";
      const x = cx * s;
      const y = cy * s;

      boxes.push(
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
        {boxes}
        {border}  {/* Add this after the cells so it's drawn on top. */}
      </svg>
    </div>
  );
}

function Clues({puzzle, cells, filled, font_size, view}) {
  // First, produce a mapping of cell number to the contents of that cell.  To
  // do this we have to go through the cells 2d array and lookup the cell number
  // for each position.
  const contents = {};
  for (let cy = 0; cy < puzzle.rows; cy++) {
    for (let cx = 0; cx < puzzle.cols; cx++) {
      const number = puzzle.cell_clue_numbers[cy][cx];
      if (number === 0) {
        continue;
      }

      contents[number] = cells[cy][cx] || "";
    }
  }

  // Next, generate each clue row.
  const rows = [];
  for (const clue of Object.keys(puzzle.clues).sort()) {
    const className = filled[clue] ? "clue-row filled" : "clue-row";

    rows.push(
      <div id={clue} className={className} key={clue}>
        <div className="clue-letter">{clue}.</div>
        <div className="clue-body">
          <div className="clue" dangerouslySetInnerHTML={{__html: puzzle.clues[clue]}}/>
          <ClueBoxes
            letters={puzzle.clue_numbers[clue].map(n => contents[n])}
            numbers={puzzle.clue_numbers[clue]}
            view={view}
          />
        </div>
      </div>
    );
  }

  // Finally we're going to render things as three separate columns so that we
  // can take advantage of extra whitespace on the right side of the screen.
  // To do this we'll just split the rows up into multiple sets.
  const split1 = Math.ceil(rows.length/3);
  const split2 = Math.ceil(split1 + (rows.length - split1)/2);

  return (
    <div className="clues" data-font-size={font_size}>
      <div className="clue-column">{rows.slice(0, split1)}</div>
      <div className="clue-column">{rows.slice(split1, split2)}</div>
      <div className="clue-column">{rows.slice(split2)}</div>
    </div>
  );
}

function ClueBoxes({letters, numbers, view}) {
  // The width of each cell in pixels -- needed for calculations.  This value is
  // arbitrary and shouldn't ever need to change.
  const w = 100;

  // How far to offset the x-coordinate in pixels for the horizontal line or
  // progress rectangle.
  const dx = w * 0.2;

  // How far from the top of the box to offset the y-coordinate in pixels for
  // the horizontal line.
  const dy = w * 0.6;

  // The amount of padding between cells in pixels -- needed for calculations.
  // This value is arbitrary and shouldn't ever need to change.  We use a
  // negative padding to get the boxes closer to each other since most of the
  // horizontal portion of the boxes is whitespace.
  const p = -18;

  const boxes = [];
  for (let i = 0; i < letters.length; i++) {
    const x = i*(w+p) + p;

    let content;
    if (view !== "progress") {
      content = (<text className="letter" x={x} y="0">{letters[i]}</text>);
    } else {
      const className = letters[i] !== "" ? "filled" : "";
      content = (<rect className={className} x={x + dx} y="0" width={w - 2*dx} height={dy}/>);
    }

    boxes.push(
      <g key={i}>
        {content}
        <text className="number" x={x} y="0">{numbers[i]}</text>
        <line x1={x + dx} y1="0" x2={x + w - dx} y2="0"/>
      </g>
    );
  }

  const width = (w + p) * letters.length;
  const height = w;

  return (
    <svg className="clue-boxes" viewBox={`0 0 ${width} ${height}`}>
      {boxes}
    </svg>
  );
}