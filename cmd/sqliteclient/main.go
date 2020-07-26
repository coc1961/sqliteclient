package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/coc1961/sqliteclient/internal/gui"
	"github.com/coc1961/sqliteclient/internal/gui/keyboard"
	"github.com/coc1961/sqliteclient/internal/sqlite3client"
)

func main() {
	var db = flag.String("d", "", "sqlite database path")
	var file = flag.String("f", "", "sql file")
	var query = flag.String("q", "", "sql to execute (use instead of -f)")
	var maxRecords = flag.Int("r", 100, "max records")

	flag.Parse()

	if *db == "" || (*file == "" && *query == "") {
		flag.Usage()
		os.Exit(1)
	}
	sql := ""
	if *query != "" {
		sql = *query
	} else {
		b, err := ioutil.ReadFile(*file)
		if err != nil {
			fmt.Println(err)
			return
		}
		sql = string(b)
	}

	sq, err := sqlite3client.New("file:" + (*db))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sq.Close()
	rows := make([][]interface{}, 0)

	max := *maxRecords
	err = sq.QueryCallback(func(columns []string, row []interface{}) error {
		if len(rows) == 0 {
			tmp := make([]interface{}, 0, len(columns))
			for _, v := range columns {
				tmp = append(tmp, v)
			}
			rows = append(rows, tmp)
		}
		rows = append(rows, row)
		max--
		if max < 0 {
			return fmt.Errorf("the maximum number of records has been exceeded (max=%d)", *maxRecords)
		}
		return nil
	}, sql)

	if err != nil {
		fmt.Println(err)
		return
	}

	txt := convertToArray(rows)
	a := gui.NewTerminal()
	defer a.Close()

	exit := make(chan int)
	a.Cls().Blue()
	w := gui.NewWindow(a, 1, 1, a.Row-1, a.Col-2)
	w.Print()
	save := a.SaveScreen()

	if len(txt) < a.Row {
		for i := len(txt); i < a.Row; i++ {
			txt = append(txt, strings.Repeat(" ", a.Col))
		}
	}

	f, c := 1, 0
	printTxt(a, txt, f, c)
	a.Keyboard(func(event keyboard.KeyEvent) {
		if event.Key == keyboard.KeyEsc && event.Err == nil {
			a.Cls().Flush()
			exit <- 1
			return
		} else if event.Key == keyboard.KeyArrowRight {
			c += 2
			if event.Ctrl {
				c += 3
			}
		} else if event.Key == keyboard.KeyArrowLeft {
			c -= 2
			if event.Ctrl {
				c -= 3
			}
		} else if event.Key == keyboard.KeyArrowDown {
			f++
			if event.Ctrl {
				f += 3
			}
		} else if event.Key == keyboard.KeyArrowUp {
			f--
			if event.Ctrl {
				f -= 3
			}
		} else if event.Key == keyboard.KeyPgup {
			f -= 5
		} else if event.Key == keyboard.KeyPgdn {
			f += 5
		} else if event.Key == keyboard.KeyEnter {
			/*
				go func() {
					fmt.Println(a.CursorPos())
				}()
			*/
		} else {
			return
		}

		if f+a.Row-6 > len(txt) {
			f = len(txt) - a.Row + 6
		}

		if c < 0 {
			c = 0
		}
		if f < 1 {
			f = 1
		}

		save.Restore()
		printTxt(a, txt, f, c)
	})
	a.Flush()
	<-exit

}

func convertToArray(t [][]interface{}) []string {
	ret := make([]string, 0, len(t))
	cols := make([]int, 500)
	for i := 0; i < len(t); i++ {
		for j := 0; j < len(t[i]); j++ {
			s := fmt.Sprintf("%v", t[i][j])
			if cols[j] < len(s) {
				cols[j] = len(s)
			}
		}
	}
	for i := 0; i < len(t); i++ {
		b := bytes.Buffer{}
		for j := 0; j < len(t[i]); j++ {
			s := fmt.Sprintf("%v", t[i][j])
			if len(s) < cols[j] {
				s += strings.Repeat(" ", cols[j]-len(s))
			}
			b.WriteString(" | ")
			b.WriteString(s)
		}
		ret = append(ret, b.String())
	}

	return ret
}

func printTxt(a *gui.Terminal, txt []string, f, c int) {
	a.Goto(0, a.Col-15).Print("[", f, "/", c, "]")
	if len(txt) == 0 {
		a.Flush()
		return
	}
	maxRow := a.Row - 3
	maxCol := a.Col - 3
	if f == 0 {
		f++
	}
	l := f + maxRow - 1
	if l > len(txt) {
		l = len(txt)
	}
	tmp := txt[f:l]
	sa := make([]string, 0, len(tmp)+1)
	sa = append(sa, txt[0])
	sa = append(sa, tmp...)
	color := gui.Green
	intensity := gui.Bold
	j := 0

	bar := func(s string) string {
		return strings.ReplaceAll(s, "|", "\u2502")
	}

	hor := strings.Repeat("\u2501", maxCol)
	a.Goto(3, 2).SetIntensity(intensity).SetColor(color).SetBackground(gui.CyanBackground).Print(hor)
	for i := 3; i < maxRow+1; i++ {
		r := i - 3
		if r > len(sa)-1 {
			return
		}
		str := sa[r]
		j = i
		if i == 3 {
			j--
		}
		line := ""
		if c > len(str)-1 {
			line = strings.Repeat(" ", maxCol)
		} else if len(str[c:]) < maxCol {
			line = bar(string(str[c:])) + strings.Repeat(" ", maxCol-len(str[c:]))
		} else {
			line = string(str[c : c+maxCol])
		}
		a.Goto(j, 2).SetIntensity(intensity).SetColor(color).Print(line)
		color = gui.Black + 1
		intensity = gui.Faint
	}
	for j++; j < maxRow+3; j++ {
		a.Goto(j, 2).SetIntensity(intensity).SetColor(color).Print(strings.Repeat(" ", maxCol))
	}
	a.Reset()
	a.Flush()
}
