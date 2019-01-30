import crosswords.api

puzzle = crosswords.api.load_puzzle("2/15/1942")

# Render the solved puzzle in ASCII.
for row in range(puzzle.rows):
    if row == 0:
        # Top row
        print("┌───", end="")
        for col in range(puzzle.cols - 1):
            print("┬───", end="")
        print("┐")
    else:
        # Bottom of previous row
        for col in range(puzzle.cols):
            if col == 0:
                print("├───", end="")
            else:
                print("┼───", end="")
        print("┤")

    # Top row of contents for this row (contains clue number)
    for col in range(puzzle.cols):
        num = puzzle.cell_clue_numbers[row][col]
        s = f"│{num:<3}" if num != 0 else "│   "
        print(s if puzzle.cells[row][col] is not None else "│███", end="")
    print("│")

    # Middle row of contents for this row (always empty)
    for col in range(puzzle.cols):
        print("│   " if puzzle.cells[row][col] is not None else "│███", end="")
    print("│")

    # Bottom row of contents for this row (contains answer)
    for col in range(puzzle.cols):
        ans = puzzle.cells[row][col]
        s = f"│{ans:^3}" if ans is not None else "│███"
        print(s, end="")
    print("│")

for col in range(puzzle.cols):
    if col == 0:
        print("└───", end="")
    else:
        print("┴───", end="")
print("┘")
print()

# Render the clues
print("Across:")
for num, clue in sorted(puzzle.across_clues.items()):
    print(f"{num:>3}. {clue}")
print()

print("Down:")
for num, clue in sorted(puzzle.down_clues.items()):
    print(f"{num:>3}. {clue}")
print()
