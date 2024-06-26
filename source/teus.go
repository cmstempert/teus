package main

import "os"
import "fmt"
import "bufio"
import "strconv"
import "strings"
import "github.com/nsf/termbox-go"
import "github.com/mattn/go-runewidth"

var mode int
var ROWS, COLS int
var offset_col, offset_row int
var current_col, current_row int
var text_buffer = [][]rune{}
var undo_buffer = [][]rune{}
var copy_buffer = []rune{}
var source_file string
var modified bool

func read_file(filename string) {
    file, err := os.Open(filename)

    if err != nil {
        source_file = filename
        text_buffer = append(text_buffer, []rune{}); return
    }

    defer file.Close()
    scanner := bufio.NewScanner(file)
    lineNumber := 0

    for scanner.Scan() {
        line := scanner.Text()
        text_buffer = append(text_buffer, []rune{})

        for i := 0; i < len(line); i++ {
            text_buffer[lineNumber] = append(text_buffer[lineNumber], rune(line[i]))
        }
        lineNumber++
    }
    if lineNumber == 0 {
        text_buffer = append(text_buffer, []rune{})
    }

}

func write_file(filename string) {
    file, err := os.Create(filename)
    if err != nil { fmt.Println(err) }
    defer file.Close()

    writer := bufio.NewWriter(file)
    for row, line := range text_buffer {
        new_line := "\n"
        if row == len(text_buffer) - 1 { new_line = "" }
        write_line := string(line) + new_line
        _, err = writer.WriteString(write_line)
        if err != nil { fmt.Println("Error: ", err) }
    }; writer.Flush(); modified = false
}

func insert_rune(event termbox.Event) {
    insert_rune := make([]rune, len(text_buffer[current_row]) + 1)
    copy(insert_rune[:current_col], text_buffer[current_row][:current_col])

    if event.Key == termbox.KeySpace { insert_rune[current_col] = rune(' ')
    } else if event.Key == termbox.KeyTab { insert_rune[current_col] = rune(' ')
    } else { insert_rune[current_col] = rune(event.Ch) }

    copy(insert_rune[current_col + 1:], text_buffer[current_row][current_col:])
    text_buffer[current_row] = insert_rune
    current_col++
}

func delete_rune() {
    if current_col > 0 {
        current_col--
        delete_line := make([]rune, len(text_buffer[current_row]) - 1)
        copy(delete_line[:current_col], text_buffer[current_row][:current_col])
        copy(delete_line[current_col:], text_buffer[current_row][current_col + 1:])
        text_buffer[current_row] = delete_line
    } else if current_row > 0 {
        append_line := make([]rune, len(text_buffer[current_row]))
        copy(append_line, text_buffer[current_row][current_col:])
        new_text_buffer := make([][]rune, len(text_buffer) - 1)
        copy(new_text_buffer[:current_row], text_buffer[:current_row])
        copy(new_text_buffer[current_row:], text_buffer[current_row + 1:])
        text_buffer = new_text_buffer; current_row--
        current_col = len(text_buffer[current_row])
        insert_line := make([]rune, len(text_buffer[current_row]) + len(append_line))
        copy(insert_line[:len(text_buffer[current_row])], text_buffer[current_row])
        copy(insert_line[len(text_buffer[current_row]):], append_line)
        text_buffer[current_row] = insert_line
    }
}

func insert_line() {
    right_line := make([]rune, len(text_buffer[current_row][current_col:]))
    copy(right_line, text_buffer[current_row][current_col:])
    left_line := make([]rune, len(text_buffer[current_row][:current_col]))
    copy(left_line, text_buffer[current_row][:current_col])
    text_buffer[current_row] = left_line
    current_row++; current_col = 0
    new_text_buffer := make([][]rune, len(text_buffer) + 1)
    copy(new_text_buffer, text_buffer[:current_row])
    new_text_buffer[current_row] = right_line
    copy(new_text_buffer[current_row + 1:], text_buffer[current_row:])
    text_buffer = new_text_buffer
}

func copy_line() {
    copy_line := make([]rune, len(text_buffer[current_row]))
    copy(copy_line, text_buffer[current_row])
    copy_buffer = copy_line
}

func cut_line() {
    copy_line()
    if current_row >= len(text_buffer) || len(text_buffer) < 2 { return }
    new_text_buffer := make([][]rune, len(text_buffer) - 1)
    copy(new_text_buffer[:current_row], text_buffer[:current_row])
    copy(new_text_buffer[current_row:], text_buffer[current_row + 1:])
    text_buffer = new_text_buffer
    if current_row > 0 { current_row--; current_col = 0 }
}

func paste_line() {
    if len(copy_buffer) == 0 { current_row++; current_col = 0 }
    new_text_buffer := make([][]rune, len(text_buffer) + 1)
    copy(new_text_buffer[:current_row], text_buffer[:current_row])
    new_text_buffer[current_row] = copy_buffer
    copy(new_text_buffer[current_row + 1:], text_buffer[current_row + 1:])
    text_buffer = new_text_buffer
}

func push_buffer() {
    copy_undo_buffer := make([][]rune, len(text_buffer))
    copy(copy_undo_buffer, text_buffer)
    undo_buffer = copy_undo_buffer
}

func pull_buffer() {
    if len(undo_buffer) == 0 { return }
    text_buffer = undo_buffer
}

func scroll_text_buffer() {
    if current_row < offset_row { offset_row = current_row }
    if current_col < offset_col { offset_col = current_col }
    if current_row >= offset_row + ROWS { offset_row = current_row - ROWS + 1 }
    if current_col >= offset_col + COLS { offset_col = current_col - COLS + 1 }
}

func display_text_buffer() {
    var row, col int
    for row = 0; row < ROWS; row++ {
        text_buffer_row := row + offset_row
        for col = 0; col < COLS; col ++ {
            text_buffer_col := col + offset_col
            if text_buffer_row >= 0 && text_buffer_row < len(text_buffer) && text_buffer_col < len(text_buffer[text_buffer_row]) {
                if text_buffer[text_buffer_row][text_buffer_col] != '\t' {
                    termbox.SetChar(col, row, text_buffer[text_buffer_row][text_buffer_col])
                } else { termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorGreen) }
            } else if row + offset_row > len(text_buffer)-1 {
                  termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault) }
        }
              termbox.SetChar(col, row, rune('\n'))
    }
}

func display_status_bar() {
    var mode_status string
    var copy_status string
    var undo_status string
    var file_status string
    var cursor_status string

    if mode > 0 { mode_status = "EDIT: "
    } else { mode_status = "VIEW: " }

    filename_length := len(source_file)
    if filename_length > 12 { filename_length = 12 }

    file_status = source_file[:filename_length] + " - " + strconv.Itoa(len(text_buffer)) + " lines"

    if modified { file_status += " modified"
    } else { file_status += " saved" }

    cursor_status = " Row " + strconv.Itoa(current_col+1) + " Column " + strconv.Itoa(current_row+1) + " "

    if len(copy_buffer) > 0 { copy_status = " [Copy] " }
    if len(undo_buffer) > 0 { undo_status = " [Undo] " }

    used_space := len(mode_status) + len(file_status) + len(cursor_status) + len(copy_status) + len(undo_status)
    filler_spaces := strings.Repeat(" ", COLS - used_space)
    message := mode_status + file_status + copy_status + undo_status + filler_spaces + cursor_status
    print_message(0, ROWS, termbox.ColorBlack, termbox.ColorWhite, message)
}

func print_message(col, row int, fg, bg termbox.Attribute, message string) {
    for _, ch := range message {
        termbox.SetCell(col, row, ch, fg, bg)
        col += runewidth.RuneWidth(ch)
    }
}

func get_key() termbox.Event {
    var key_event termbox.Event
    switch event := termbox.PollEvent(); event.Type {
        case termbox.EventKey: key_event = event
        case termbox.EventError: panic(event.Err)
    }; return key_event
}

func process_key_press() {
    key_event := get_key()
    if key_event.Key == termbox.KeyEsc || key_event.Key == termbox.KeyCtrlJ { mode = 0
    } else if key_event.Ch != 0 {
        if mode == 1 { insert_rune(key_event); modified = true
        } else {
            switch key_event.Ch {
                case 'q': termbox.Close(); os.Exit(0)
                case 'i': mode = 1
                case 'w': write_file(source_file)
                case 'c': copy_line()
                case 'p': paste_line()
                case 'x': cut_line()
                case 'v': push_buffer()
                case 'n': pull_buffer()
                case 'h': if current_col != 0 { current_col-- } else if current_row > 0 {
                              current_row--; current_col = len(text_buffer[current_row]) }
                case 'k': if current_row != 0 { current_row-- }
                case 'j': if current_row < len(text_buffer) - 1 { current_row++ }
                case 'l': if current_col < len(text_buffer[current_row]) { current_col++
                              } else if current_row < len(text_buffer) - 1 { current_row++; current_col = 0 }
                case '0': current_col = 0
                case '$': current_col = len(text_buffer[current_row])
            }
        }
    } else {
        switch key_event.Key {
            case termbox.KeyEnter: if mode == 1 { insert_line(); modified = true }
            case termbox.KeyBackspace: if mode == 1 { delete_rune(); modified = true }
            case termbox.KeyBackspace2: if mode == 1 { delete_rune(); modified = true }
            case termbox.KeyTab:
                if mode == 1 {
                    for i:= 0; i <4; i++ { insert_rune(key_event) }
                    modified = true
                }
            case termbox.KeySpace:
                if mode == 1 { insert_rune(key_event); modified = true }
            case termbox.KeyHome: current_col = 0
            case termbox.KeyEnd: current_col = len(text_buffer[current_row])
            case termbox.KeyPgup: if current_row - int(ROWS/4) > 0 { current_row -= int(ROWS/4) }
            case termbox.KeyPgdn: if current_row + int(ROWS/4) < len(text_buffer) - 1 { current_row += int(ROWS/4) }
            case termbox.KeyArrowUp: if current_row != 0 { current_row-- }
            case termbox.KeyArrowDown: if current_row < len(text_buffer) - 1 { current_row++ }
            case termbox.KeyArrowLeft:
                if current_col != 0 {
                    current_col--
                } else if current_row > 0 {
                    current_row--
                    current_col = len(text_buffer[current_row])
                }
            case termbox.KeyArrowRight:
                if current_col < len(text_buffer[current_row]) {
                    current_col++
                } else if current_row < len(text_buffer) - 1 {
                    current_row++
                    current_col = 0
                }
        }
    }; if current_col > len(text_buffer[current_row]) {
            current_col = len(text_buffer[current_row])
       }
}

func run_editor() {
    err := termbox.Init()
    if err != nil { fmt.Println(err); os.Exit(1) }
    if len(os.Args) > 1 {
        source_file = os.Args[1]
        read_file(source_file)
    } else {
        source_file = "NewFile.txt"
        text_buffer = append(text_buffer, []rune{})
    }

    for {
        COLS, ROWS = termbox.Size(); ROWS--
        if COLS < 80 { COLS  = 80 }
        termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
        scroll_text_buffer()
        display_text_buffer()
        display_status_bar()
        termbox.SetCursor(current_col - offset_col, current_row - offset_row)
        termbox.Flush()
        process_key_press()
    }
}

func main() {
    run_editor()
}
