import React from "react";
import "bootstrap/dist/js/bootstrap.bundle.min";
import "bootstrap/dist/css/bootstrap.min.css";
import {Timer, TwitchChat} from "common/view";
import "spellingbee/view.css";

export function SpellingBeeView({channel, view, state, settings}) {
  if (!state.puzzle) {
    return (
      <div className="jumbotron">
        <h1>Welcome to Spelling Bee!</h1>
        <hr className="my-4"/>
        <p>
          Please select a spelling bee puzzle to solve using the&nbsp;
          <b>Puzzle</b> menu on the top right of the screen.
        </p>
        <p>
          Once a puzzle is started Puzzles With Chat will allow a Twitch
          streamer's viewers to cooperatively solve a spelling bee puzzle by
          inputting answers into the chat.  As correct answers are inputted
          they will appear on the screen and the score will increase.  Find all
          of the answers for a fun congratulatory demonstration.
        </p>
        <hr className="my-4"/>
        <p>
          Questions? Comments? Feedback? Feel free to whisper @mistaeksweremade
          on Twitch.
        </p>
      </div>
    );
  }

  const puzzle = state.puzzle;
  const status = state.status;

  let total_num_words = puzzle.num_official_answers;
  if (settings.allow_unofficial_answers) {
    total_num_words += puzzle.num_unofficial_answers;
  }

  const max_score = !settings.allow_unofficial_answers
    ? puzzle.max_official_score
    : puzzle.max_unofficial_score;
  const isGenius = state.score >= Math.round(max_score * 0.7);
  const isQueenBee = state.score === max_score;

  return (
    <div id="spellingbee" className={status === "selected" || status === "paused" ? "blur" : ""}>
      <div id="genius-banner" className="banner">
        GENIUS!<span role="img" aria-label="genius">&nbsp;&#x1f393;</span>
      </div>
      <div className="puzzle">
        <Header
          date={puzzle.published}
          score={state.score}
          isGenius={isGenius}
          isQueenBee={isQueenBee}
          last_start_time={state.last_start_time}
          total_solve_duration={state.total_solve_duration}
        />
        <Grid center={puzzle.center} letters={state.letters}/>
        <Footer/>
      </div>
      <WordsList
        show_placeholders={settings.show_answer_placeholders}
        font_size={settings.font_size}
        view={view}
        words={state.words}
        total={total_num_words}
      />
      { view === "player" && <TwitchChat channel={channel}/> }
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

  let tag;
  if (props.isQueenBee) {
    tag = (<span role="img" aria-label="queen bee">&nbsp;&#x1f41d;</span>);
  } else if (props.isGenius) {
    tag = (<span role="img" aria-label="genius">&nbsp;&#x1f393;</span>);
  }

  return (
    <div className="header">
      <div className="date">{formatDate(props.date)}</div>
      <div className="score">{props.score} points {tag}</div>
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

function WordsList({show_placeholders, font_size, view, words, total}) {
  const isProgress = view === "progress";
  const className = isProgress ? "word filled" : "word";

  // Helper to obscure answers when we're showing only progress.
  const mask = (word, seed) => {
    const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";

    let masked = "";
    for (let j = 0; j < word.length; j++) {
      const index = Math.floor(Math.abs(Math.sin(seed + j) * 10000));
      masked += alphabet[index % alphabet.length];
    }

    return masked;
  };

  // Create a word entry for each provided word, indexed by the position in the
  // list the word belongs at.  This will only insert entries for the words that
  // have actually been provided.  Other indices will be missing.
  const entries = {};
  for (const [word, index] of Object.entries(words)) {
    entries[index] = (
      <div className={className} key={index}>
        {!isProgress ? word : mask(word, index)}
      </div>
    );
  }

  // If we're showing all words, not just the ones that have been provided then
  // we insert placeholder entries into the map.
  if (show_placeholders) {
    for (let index = 0; index < total; index++) {
      if (!entries[index]) {
        entries[index] = (
          <div className={className} key={index}>&nbsp;</div>
        );
      }
    }
  }

  return (
    <div className="word-list" data-font-size={font_size}>
      <div className="header">Found <b>{Object.keys(words).length}</b> out of {total} words</div>
      <div className="words">
        {
          // We have to make sure to iterate the entries in index order.
          [...Array(total).keys()].map(index => entries[index])
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
      <div>Provide an answer: <code>!country</code></div>
      <div>Shuffle the letters: <code>!shuffle</code></div>
    </div>
  );
}
