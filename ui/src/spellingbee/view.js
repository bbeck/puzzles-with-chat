import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import {Timer} from "common/view";
import "spellingbee/view.css";

export function SpellingBeeView(props) {
  const state = props.state;
  const puzzle = state && state.puzzle;
  if (!state || !puzzle) {
    return (
      <h1>Spelling Bee goes here</h1>
    );
  }

  const status = state.status;
  const settings = props.settings;

  return (
    <div id="spellingbee" className={status === "selected" || status === "paused" ? "blur" : ""}>
      <div className="puzzle">
        <Header
          date={puzzle.published}
          score={state.score}
          last_start_time={state.last_start_time}
          total_solve_duration={state.total_solve_duration}
        />
        <Grid center={puzzle.center} letters={puzzle.letters}/>
        <Footer/>
      </div>
      <WordsList
        font_size={settings.font_size}
        view={props.view}
        words={state.words}
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

  return (
    <div className="header">
      <div className="date">{formatDate(props.date)}</div>
      <div className="score">{props.score} points</div>
      <Timer
        last_start_time={props.last_start_time}
        total_solve_duration={props.total_solve_duration}
      />
    </div>
  );
}

function Grid(props) {
  // These coordinate/dimensions are within the viewbox of the SVG image.  They
  // will be automatically scaled by the size of the SVG container and don't
  // ever need to change.
  const w = 1000;
  const h = 1000;
  const s = 175;
  const x = w / 2;
  const y = h / 2;
  const dx = 3 * s / 2;
  const dy = (Math.sqrt(3) / 2) * s;

  return (
    <div className="grid">
      <svg viewBox={`0 0 ${w} ${h}`}>
        <Cell letter={(props.letters)[0]} x={x} y={y - 2 * dy} s={s}/>
        <Cell letter={(props.letters)[1]} x={x - dx} y={y - dy} s={s}/>
        <Cell letter={(props.letters)[2]} x={x + dx} y={y - dy} s={s}/>
        <Cell letter={props.center} className="center" x={x} y={y} s={s}/>
        <Cell letter={(props.letters)[3]} x={x - dx} y={y + dy} s={s}/>
        <Cell letter={(props.letters)[4]} x={x + dx} y={y + dy} s={s}/>
        <Cell letter={(props.letters)[5]} x={x} y={y + 2 * dy} s={s}/>
      </svg>
    </div>
  );
}

// Draws a hexagonal cell centered at (x, y) with a side length of s.  The
// provided letter will be rendered within the cell.
function Cell(props) {
  const x = props.x;
  const y = props.y;
  const s = props.s;

  const tl = {x: x - s / 2, y: y - Math.sqrt(3) / 2 * s};
  const tr = {x: x + s / 2, y: y - Math.sqrt(3) / 2 * s};
  const ml = {x: x - s, y: y};
  const mr = {x: x + s, y: y};
  const bl = {x: x - s / 2, y: y + Math.sqrt(3) / 2 * s};
  const br = {x: x + s / 2, y: y + Math.sqrt(3) / 2 * s};

  const points = [
    `${tl.x},${tl.y}`,
    `${tr.x},${tr.y}`,
    `${mr.x},${mr.y}`,
    `${br.x},${br.y}`,
    `${bl.x},${bl.y}`,
    `${ml.x},${ml.y}`
  ];

  let className = "cell";
  if (props.className) {
    className += " " + props.className;
  }

  return (
    <>
      <polygon className={className} points={points.join(" ")}/>
      <text className="text" x={x} y={y}>{props.letter}</text>
    </>
  );
}

function WordsList(props) {
  const isProgress = props.view === "progress";
  const className = isProgress ? "word filled" : "word";

  const words = props.words || [];
  if (isProgress) {
    // Obscure the answers when we're only showing progress.
    const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";

    for (let i = 0; i < words.length; i++) {
      const word = words[i];

      let masked = "";
      for (let j = 0; j < word.length; j++) {
        const index = Math.floor(Math.abs(Math.sin(i+j) * 10000));
        masked += alphabet[index % alphabet.length];
      }

      words[i] = masked;
    }
  }

  return (
    <div className="word-list" data-font-size={props.font_size}>
      <div className="header">Found <b>{words.length}</b> words</div>
      <div className="words">
        {
          words.map((word, i) => {
            return <div className={className} key={i}>{word}</div>;
          })
        }
      </div>
    </div>
  );
}

function Footer() {
  return (
    <div className="footer">
      <div>
        <b>RULES:</b>&nbsp;
        Construct words that are at least 4 letters long and use the center
        letter. Letters may be used more than once.
      </div>
      <div>Provide an answer: <code>!answer country</code></div>
      <div>Shuffle the letters: <code>!shuffle</code></div>
    </div>
  );
}
