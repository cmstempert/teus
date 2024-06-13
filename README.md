# teus - terminal editor, ultra-simple

teus is an ultra-simple terminal-based text editor written in Go.

## Features
* Bimodal (View/Edit)
* Status bar
* Key-based navigation
* Write and delete text
* Copy/paste lines
* Undo/redo

## Key Bindings
```
     # Mode Controls
Ctrl+J: enter VIEW mode
   ESC: enter VIEW mode
     i: enter EDIT mode

     # Navigation
     h: move cursor left
     j: move cursor down
     k: move cursor up
     l: move cursor right
     g: jump to first row of file
     G: jump to last row of file
     0: jump to start of current line
     $: jump to end of current line
     Arrows: move the cursor

     # Actions
     w: save file
     q: exit the program (be sure to save your work first!)
     c: copy current line to copy buffer
     p: paste line from copy buffer
     x: cut current line to copy buffer
     v: push text buffer to undo buffer
     n: pull text buffer from undo buffer
```

## Usage
```
$ teus           # open editor with 'NewFile.txt' source file
$ teus file.txt  # opens "file.txt" in editor, if it exists. if
                 # it does not exist, opens a new file by that name
```
