const pieces = {
  r: "♜",
  n: "♞",
  b: "♝",
  q: "♛",
  k: "♚",
  p: "♟",
  R: "♖",
  N: "♘",
  B: "♗",
  Q: "♕",
  K: "♔",
  P: "♙",
};

let board = [
  ["r", "n", "b", "q", "k", "b", "n", "r"],
  ["p", "p", "p", "p", "p", "p", "p", "p"],
  ["", "", "", "", "", "", "", ""],
  ["", "", "", "", "", "", "", ""],
  ["", "", "", "", "", "", "", ""],
  ["", "", "", "", "", "", "", ""],
  ["P", "P", "P", "P", "P", "P", "P", "P"],
  ["R", "N", "B", "Q", "K", "B", "N", "R"],
];

let selected = null;

function drawBoard() {
  const b = document.getElementById("board");
  b.innerHTML = "";
  for (let y = 0; y < 8; y++)
    for (let x = 0; x < 8; x++) {
      const cell = document.createElement("div");
      cell.className =
        `cell ${(x + y) % 2 ? "black" : "white"}` +
        (selected && selected.x === x && selected.y === y ? " selected" : "");
      cell.textContent = pieces[board[y][x]] || "";
      cell.onclick = () => handleClick(x, y);
      b.appendChild(cell);
    }
}

function handleClick(x, y) {
  if (selected) {
    board[y][x] = board[selected.y][selected.x];
    board[selected.y][selected.x] = "";
    selected = null;
  } else if (board[y][x]) {
    selected = { x, y };
  }
  drawBoard();
}

drawBoard();
