package main

import gc "github.com/rthornton128/goncurses"

//TODO: 改行はできているが、そのあとの画面のリフレッシュがうまく行かない
func (v *View) insertNewLine() {
	str_copy := []byte(v.file.buf[v.cursor.text_y])
	oline := str_copy[0:v.cursor.x]
	nline := str_copy[v.cursor.x:]
	if len(nline) == 0 {
		nline = []byte(" ")
	}

	v.file.buf = append(v.file.buf[:v.cursor.text_y+1], v.file.buf[v.cursor.text_y:]...)
	v.file.buf[v.cursor.text_y] = string(oline)
	v.file.buf[v.cursor.text_y+1] = string(nline)

	v.ScrollWin(true)
	//TODO: insertして新しくしたところから下をrefresh
	v.refresh()
	v.cursor.text_y++
	v.cursor.y++
	v.main_window.Move(v.cursor.y, v.cursor.x)
}

func (v *View) insert(ch gc.Key) {
	text := ""
	pos := v.cursor.x - 1
	if len(v.file.buf[v.cursor.text_y]) != 0 {
		for i, t := range []byte(v.file.buf[v.cursor.text_y]) {
			text += string(t)
			if i == pos {
				text += string(ch)
			}
		}
	} else {
		text += string(ch)
	}
	v.file.buf[v.cursor.text_y] = text
	v.main_window.MovePrint(v.cursor.y, 0, v.file.buf[v.cursor.text_y])
	v.main_window.Refresh()
	v.ScrollWin(false)
	v.cursor.x++
	v.main_window.Move(v.cursor.y, v.cursor.x)
	v.main_window.Refresh()
}

func (v *View) Insert(ch gc.Key) {
	if ch == '\n' {
		v.insertNewLine()
	} else if len(string(ch)) == 1 {
		v.insert(ch)
	}
}

func (v *View) delete() {
	if v.cursor.x > 0 {
		text := []byte(v.file.buf[v.cursor.y])
		text = append(text[:v.cursor.x-1], text[v.cursor.x:]...)
		v.file.buf[v.cursor.y] = string(text)

		v.main_window.MovePrint(v.cursor.y, 0, v.file.buf[v.cursor.text_y])
		v.main_window.Refresh()
		v.cursor.x--
		v.main_window.Move(v.cursor.y, v.cursor.x)
		v.main_window.Refresh()
	}
}

//TODO: 文字の入力
func (v *View) InsertCommand(ch gc.Key) {
	switch ch {
	case ESC_KEY:
		v.mode = Normal
	case gc.KEY_RIGHT, gc.KEY_UP, gc.KEY_DOWN, gc.KEY_LEFT:
		v.CursorMove(ch)
	case DELETE_KEY:
		v.delete()
	//何もしないと一定時間ごとに 0 がgetcharからかえる
	case 0:
	default:
		v.Insert(ch)
	}
}
