function render_crossword(puzzle) {
  var crossword = document.querySelector("#crossword");
  crossword.setAttribute("data-size", puzzle["rows"]);

  var title = document.querySelector("#crossword #title");
  title.innerText = puzzle["title"];

  var date = document.querySelector("#crossword #date");
  date.innerText = puzzle["published"];

  render_grid(puzzle);

  var across = document.querySelector("#crossword #across-clues");
  render_clues(puzzle.across_clues, across, "a");
  var down = document.querySelector("#crossword #down-clues");
  render_clues(puzzle.down_clues, down, "d");
}

function render_grid(puzzle) {
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
        contentDiv.innerText = cellText;

        // Add a data attribute that indicates the number of characters
        // present in the cell.
        contentDiv.setAttribute("data-length", cellText.length);
      }
    }
  }

  // Clear the old grid and swap the new one into place.
  var gridDiv = document.querySelector("#crossword #grid");
  clear(gridDiv);
  gridDiv.appendChild(table);
}

function render_clues(clues, root, side) {
  /*
    We're going to model the clues as an ol of items with each clue being a
    single li within the list.  An individual clue will be comprised of the
    number of the clue followed by the text of the clue each in their
    respective span.
  */

  // First sort the numbers to make sure the clues show in the correct order.
  var numbers = Object.keys(clues);
  numbers.sort(function (a, b) {
    return parseInt(a) < parseInt(b);
  });

  var list = document.createElement("ol");
  for (var i = 0; i < numbers.length; i++) {
    var li = document.createElement("li");
    li.setAttribute("id", numbers[i] + side);
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

function clear(elem) {
  while (elem.firstChild) {
    elem.removeChild(elem.firstChild);
  }
}