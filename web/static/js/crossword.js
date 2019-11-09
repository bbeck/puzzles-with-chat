function set_play_pause(state) {
  var crossword = document.querySelector("#crossword");
  if (crossword === null) {
    return;
  }

  var button = document.querySelector("#start-pause");
  // It's okay if the button isn't present as non-streamer views won't have it,
  // but their puzzle should still be blurred.

  if (state === "created") {
    if (button !== null) {
      button.classList.remove("invisible");
      button.classList.remove("btn-warning");
      button.classList.add("btn-success")
      button.innerText = "Start";
      button.disabled = false;
    }

    crossword.classList.add("blur");
  } else if (state === "paused") {
    if (button !== null) {
      button.classList.remove("invisible");
      button.classList.remove("btn-warning");
      button.classList.add("btn-success")
      button.innerText = "Unpause";
      button.disabled = false;
    }

    crossword.classList.add("blur");
  } else if (state === "playing") {
    if (button !== null) {
      button.classList.remove("invisible");
      button.classList.add("btn-warning");
      button.classList.remove("btn-success")
      button.innerText = "Pause";
      button.disabled = false;
    }

    crossword.classList.remove("blur");
  } else if (state === "complete") {
    if (button !== null) {
      button.classList.remove("btn-warning");
      button.classList.add("invisible");
      button.classList.add("btn-success")
      button.innerText = "Complete";
      button.disabled = true;
    }

    crossword.classList.remove("blur");
  }
}

function set_date(date) {
  var selector = document.querySelector("#date-selector");
  if (selector !== null) {
    selector.value = date;
  }
}

function set_settings(settings) {
  var checkbox = document.querySelector("#correct-answers-only");
  if (checkbox !== null) {
    checkbox.checked = settings["only_allow_correct_answers"];
  }

  var on = function (button) {
    if (button !== null) {
      button.classList.add("btn-success");
      button.classList.remove("btn-dark");
    }
  };
  var off = function (button) {
    if (button !== null) {
      button.classList.add("btn-dark");
      button.classList.remove("btn-success");
    }
  };
  var show = function (div) {
    if (div !== null) {
      div.style.display = null;
    }
  }
  var hide = function (div) {
    if (div !== null) {
      div.style.display = "none";
    }
  }

  var showAllClues = document.querySelector("#show-all-clues");
  var hideAcrossClues = document.querySelector("#hide-across-clues");
  var hideDownClues = document.querySelector("#hide-down-clues");
  var acrossClues = document.querySelector("#crossword .clues .across");
  var downClues = document.querySelector("#crossword .clues .down");
  if (settings["hide_clues"] == "none") {
    on(showAllClues);
    off(hideAcrossClues);
    off(hideDownClues);
    show(acrossClues);
    show(downClues);
  } else if (settings["hide_clues"] == "across") {
    off(showAllClues);
    on(hideAcrossClues);
    off(hideDownClues);
    hide(acrossClues);
    show(downClues);
  } else if (settings["hide_clues"] == "down") {
    off(showAllClues);
    off(hideAcrossClues);
    on(hideDownClues);
    hide(downClues);
    show(acrossClues);
  }

  var fontSize = function (div, size) {
    if (div !== null) {
      div.setAttribute("data-font-size", size);
    }
  }

  var clues = document.querySelector("#crossword .clues");
  var clueFontSizeNormal = document.querySelector("#font-size-clues-normal");
  var clueFontSizeLarge = document.querySelector("#font-size-clues-large");
  var clueFontSizeXLarge = document.querySelector("#font-size-clues-xlarge");
  if (settings["clue_font_size"] == "normal") {
    on(clueFontSizeNormal);
    off(clueFontSizeLarge);
    off(clueFontSizeXLarge);
    fontSize(clues, "normal");
  } else if (settings["clue_font_size"] == "large") {
    off(clueFontSizeNormal);
    on(clueFontSizeLarge);
    off(clueFontSizeXLarge);
    fontSize(clues, "large");
  } else if (settings["clue_font_size"] == "xlarge") {
    off(clueFontSizeNormal);
    off(clueFontSizeLarge);
    on(clueFontSizeXLarge);
    fontSize(clues, "xlarge");
  }
}

// This global variable holds the id of the timer that updates how long the
// solve has been going for.  When there's not an active timer this will be
// null.
var timer_id = null;

function set_timer(state, last_start_time, total_time_secs) {
  // If there's a previous timer, then cancel it since we're about to create a
  // new one.
  if (timer_id !== null) {
    clearInterval(timer_id);
    timer_id = null;
  }

  last_start_time = Date.parse(last_start_time) / 1000;

  // Regardless of what the state is show the time
  render_timer(last_start_time, total_time_secs);

  if (state === "playing") {
    // If we're in the playing mode then we need to update the timer
    // periodically.  We'll automatically do it on every state update, but
    // towards the end of a puzzle we might not be getting state updates that
    // often.  So have a background timer that keeps it up to date.
    timer_id = setInterval(render_timer, 1000, last_start_time, total_time_secs);
  }
}

function render_crossword(state, show_only_progress) {
  var puzzle = state.puzzle;
  var across_clues_filled = state.across_clues_filled;
  var down_clues_filled = state.down_clues_filled;

  var crossword = document.querySelector("#crossword");
  crossword.setAttribute("data-size", Math.max(puzzle["cols"], puzzle["rows"]));

  var title = document.querySelector("#crossword #title");
  title.innerText = puzzle["title"];

  var author = document.querySelector("#crossword #author");
  author.innerText = puzzle["author"];

  var date = document.querySelector("#crossword #date");
  date.innerText = puzzle["published"];

  render_grid(puzzle, show_only_progress);

  var across = document.querySelector("#crossword #across-clues");
  render_clues(puzzle.across_clues, across_clues_filled, across, "a");
  var down = document.querySelector("#crossword #down-clues");
  render_clues(puzzle.down_clues, down_clues_filled, down, "d");

  var notes = document.querySelector("#crossword #clue-notes");
  if (notes !== null && puzzle.notes !== null) {
    notes.innerHTML = puzzle.notes;
  }

  // If the puzzle is unsolved thus far then show 1a and 1d.
  var unsolved = function (filled) { return filled === false; };
  if (Object.values(across_clues_filled).every(unsolved) &&
    Object.values(down_clues_filled).every(unsolved)) {
    var clueElem = document.getElementById("1a");
    if (clueElem !== null) {
      clueElem.scrollIntoView();
    }
    var clueElem = document.getElementById("1d");
    if (clueElem !== null) {
      clueElem.scrollIntoView();
    }
  }
}

function render_grid(puzzle, show_only_progress) {
  /*
    We're going to model the grid as a table with one td element per cell of
    the puzzle.  Each td in the grid will be broken in half horizontally.  The
    top part will contain the cell number if one is needed, and the bottom half
    will contain the text of that cell.
  */
  var table = document.createElement("table");
  for (var y = 0; y < puzzle["rows"]; y++) {
    var tr = document.createElement("tr");
    table.append(tr)

    for (var x = 0; x < puzzle["cols"]; x++) {
      var td = document.createElement("td");
      tr.append(td);

      var outerDiv = document.createElement("div");
      outerDiv.classList.add("cell");
      td.append(outerDiv);

      var numberDiv = document.createElement("div");
      numberDiv.classList.add("number");
      outerDiv.append(numberDiv);

      var contentDiv = document.createElement("div");
      contentDiv.classList.add("content");
      outerDiv.append(contentDiv);

      var cellText = puzzle["cells"][y][x];
      var cellNumber = puzzle["cell_clue_numbers"][y][x];
      var cellShaded = puzzle["cell_circles"][y][x];

      // Style the outer div
      if (cellText === null) {
        outerDiv.classList.add("block");
      } else if (cellShaded) {
        outerDiv.classList.add("shaded");
      }

      // Populate the cell number
      if (cellNumber !== 0) {
        numberDiv.innerText = cellNumber;
      }

      // Populate the cell content
      if (cellText !== null && cellText.length > 0) {
        if (show_only_progress) {
          // In the progress view we only indicate that the cell has been
          // filled with an answer.
          outerDiv.classList.add("filled");
        } else {
          // In the other views we fill the cell with the correct value.
          contentDiv.innerText = cellText;

          // Add a data attribute that indicates the number of characters
          // present in the cell.
          contentDiv.setAttribute("data-length", cellText.length);
        }
      }
    }
  }

  // Clear the old grid and swap the new one into place.
  var gridDiv = document.querySelector("#crossword #grid");
  clear(gridDiv);
  gridDiv.appendChild(table);
}

function render_clues(clues, filled, root, side) {
  /*
    We're going to model the clues as an ol of items with each clue being a
    single li within the list.  An individual clue will be comprised of the
    number of the clue followed by the text of the clue each in their
    respective span.
  */

  // First sort the numbers to make sure the clues show in the correct order.
  var numbers = Object.keys(clues);
  numbers.sort(function (a, b) {
    var ia = parseInt(a);
    var ib = parseInt(b);
    return (ia < ib) ? -1 : (ia == ib) ? 0 : 1;
  });

  var list = document.createElement("ul");
  for (var i = 0; i < numbers.length; i++) {
    var li = document.createElement("li");
    li.setAttribute("id", numbers[i] + side);
    if (filled[numbers[i]]) {
      li.classList.add("filled");
    }
    list.appendChild(li);

    var numberSpan = document.createElement("span");
    numberSpan.classList.add("number");
    numberSpan.innerText = numbers[i];
    li.appendChild(numberSpan);

    var textSpan = document.createElement("span");
    textSpan.classList.add("clue");
    textSpan.innerText = clues[numbers[i]];
    li.appendChild(textSpan);
  }

  clear(root);
  root.appendChild(list);
}

function render_timer(last_start_time, total_time_secs) {
  var now = new Date();

  // Compute the duration, it's comprised of 2 parts.  The accumulated time
  // which will always be a number and the time that we last entered the
  // playing state.  The last start time may be NaN when the puzzle is
  // currently paused and in that case should be ignored.
  var duration = total_time_secs;
  if (!isNaN(last_start_time)) {
    duration += (now.getTime() + now.getTimezoneOffset() * 60000) / 1000 - last_start_time;
  }

  var hours = parseInt(duration / 3600),
    mins = parseInt((duration - hours * 3600) / 60),
    secs = parseInt((duration - hours * 3600 - mins * 60));

  if (mins < 10) {
    mins = "0" + mins;
  }
  if (secs < 10) {
    secs = "0" + secs;
  }

  // Clear the old timer and swap the new one into place.
  var timerDiv = document.querySelector("#crossword #timer");
  timerDiv.innerText = hours + "h " + mins + "m " + secs + "s";
}

function clear(elem) {
  while (elem.firstChild) {
    elem.removeChild(elem.firstChild);
  }
}