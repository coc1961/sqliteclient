package gui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"
	"unsafe"

	"github.com/coc1961/sqliteclient/internal/gui/keyboard"
	"github.com/pborman/ansi"
)

const (
	Normal = iota
	Bold   // bold or increased intensity
	Faint  // faint, decreased intensity or second colour
	Italics
	Underline
	Blink
	FastBlink
	Inverse
	Hidden
	Strikeout
	PrimaryFont
	AltFont1
	AltFont2
	AltFont3
	AltFont4
	AltFont5
	AltFont6
	AltFont7
	AltFont8
	AltFont9
	Gothic // fraktur
	DoubleUnderline
	NormalColor // normal colour or normal intensity (neither bold nor faint)
	NotItalics  // not italicized, not fraktur
	NotUnderlined
	Steady     // not Blink or FastBlink
	Reserved26 // reserved for proportional spacing as specified in CCITT Recommendation T.61
	NotInverse // Positive
	NotHidden  // Revealed
	NotStrikeout
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Reserved38 // intended for setting character foreground colour as specified in ISO 8613-6 [CCITT Recommendation T.416]
	Default    // default display colour (implementation-defined)
	BlackBackground
	RedBackground
	GreenBackground
	YellowBackground
	BlueBackground
	MagentaBackground
	CyanBackground
	WhiteBackground
	Reserved48        // reserved for future standardization; intended for setting character background colour as specified in ISO 8613-6 [CCITT Recommendation T.416]
	DefaultBackground // default background colour (implementation-defined)
	Reserved50        // reserved for cancelling the effect of the rendering aspect established by parameter value 26
	Framed
	Encircled
	Overlined
	NotFramed // NotEncircled
	NotOverlined
	Reserved56
	Reserved57
	Reserved58
	Reserved59
	IdeogramUnderline       // ideogram underline or right side line
	IdeogramDoubleUnderline // ideogram double underline or double line on the right side
	IdeogramOverline        // ideogram overline or left side line
	IdeogramDoubleOverline  // ideogram double overline or double line on the left side
	IdeogramStress          // ideogram stress marking
	IdeogramCancel          // cancels the effect of the rendition aspects established by parameter values IdeogramUnderline to IdeogramStress
	Reserved66              // This should be 66
)

func NewTerminal() *Terminal {
	t := &Terminal{}
	r, c, err := t.TerminalSize()
	if err == nil {
		t.Row = r
		t.Col = c
	} else {
		t.Row = 24
		t.Col = 80
	}
	t.buf = bytes.Buffer{}
	t.w = ansi.NewWriter(&t.buf)

	t.keyboardLoop(keyboard.GetKeys(10))

	t.readyCursorPos = make(chan int, 1)

	return t
}

type SaveScreen struct {
	b bytes.Buffer
	t *Terminal
}

func newSaveScreen(t *Terminal) *SaveScreen {
	return &SaveScreen{
		t: t,
		b: bytes.Buffer{},
	}
}
func (s *SaveScreen) Restore() *SaveScreen {
	s.t.buf.Reset()
	s.t.Cls()
	s.t.buf.Write(s.b.Bytes())
	return s
}
func (s *SaveScreen) Save() *SaveScreen {
	s.b.Reset()
	s.b.Write(s.t.buf.Bytes())
	return s
}

type Terminal struct {
	Row int
	Col int

	buf bytes.Buffer
	w   *ansi.Writer

	cursorRow      int
	cursorCol      int
	readyCursorPos chan int

	keyboardFunc func(key keyboard.KeyEvent)
}

func (a *Terminal) Flush() *Terminal                                  { fmt.Print(a.buf.String()); a.buf.Reset(); a.w.Reset(); return a }
func (a *Terminal) Cls() *Terminal                                    { a.buf.Reset(); a.w.Print("\033[2J"); a.w.Print("\033[H"); return a }
func (a *Terminal) Reset() *Terminal                                  { a.w = a.w.Reset(); return a }
func (a *Terminal) ForceReset() *Terminal                             { a.w.ForceReset(); return a }
func (a *Terminal) Black() *Terminal                                  { a.w = a.w.Black(); return a }
func (a *Terminal) Red() *Terminal                                    { a.w = a.w.Red(); return a }
func (a *Terminal) Green() *Terminal                                  { a.w = a.w.Green(); return a }
func (a *Terminal) Yellow() *Terminal                                 { a.w = a.w.Yellow(); return a }
func (a *Terminal) Blue() *Terminal                                   { a.w = a.w.Blue(); return a }
func (a *Terminal) Magenta() *Terminal                                { a.w = a.w.Magenta(); return a }
func (a *Terminal) Cyan() *Terminal                                   { a.w = a.w.Cyan(); return a }
func (a *Terminal) White() *Terminal                                  { a.w = a.w.White(); return a }
func (a *Terminal) Default() *Terminal                                { a.w = a.w.Default(); return a }
func (a *Terminal) Bold() *Terminal                                   { a.w = a.w.Bold(); return a }
func (a *Terminal) Faint() *Terminal                                  { a.w = a.w.Faint(); return a }
func (a *Terminal) Normal() *Terminal                                 { a.w = a.w.Normal(); return a }
func (a *Terminal) Goto(r, c int) *Terminal                           { a.w.Print(fmt.Sprintf("\033[%d;%dH", r, c)); return a }
func (a *Terminal) Print(aa ...interface{}) (int, error)              { return a.w.Print(aa...) }
func (a *Terminal) Printf(fmt string, aa ...interface{}) (int, error) { return a.w.Printf(fmt, aa...) }
func (a *Terminal) Println(aa ...interface{}) (int, error)            { return a.w.Println(aa...) }
func (a *Terminal) SetIntensity(intensity int) *Terminal              { a.w = a.w.SetIntensity(intensity); return a }
func (a *Terminal) SetBackground(bg int) *Terminal                    { a.w = a.w.SetBackground(bg); return a }
func (a *Terminal) SetColor(fg int) *Terminal                         { a.w = a.w.SetColor(fg); return a }
func (a *Terminal) ForceSet() (int, error)                            { return a.w.ForceSet() }
func (a *Terminal) Set() (int, error)                                 { return a.w.Set() }
func (a *Terminal) WriteString(in string) (int, error)                { return a.w.WriteString(in) }
func (a *Terminal) Write(in []byte) (int, error)                      { return a.w.Write(in) }
func (a *Terminal) NoColor() *Terminal                                { a.w.NoColor(); return a }
func (a *Terminal) SaveScreen() *SaveScreen                           { return newSaveScreen(a).Save() }
func (a *Terminal) CursorPos() (int, int) {
	go func() {
		_, _ = io.WriteString(os.Stdout, "\x1b[6n")
	}()
	select {
	case <-a.readyCursorPos:
	case <-time.After(time.Second):
	}
	return a.cursorRow, a.cursorCol
}

func (a *Terminal) TerminalSize() (int, int, error) {
	type winsize struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}

	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return -1, -1, errno
	}
	return int(ws.Row), int(ws.Col), nil
}

func (a *Terminal) Close() {
	_ = keyboard.Close()
	defer func() {
		cmd := exec.Command("reset")
		_ = cmd.Start()
		_ = cmd.Wait()
	}()
}

func (a *Terminal) Keyboard(fn func(key keyboard.KeyEvent)) *Terminal {
	a.keyboardFunc = a.makeKeyFunc(fn)
	return a
}

func (a *Terminal) makeKeyFunc(fn func(key keyboard.KeyEvent)) func(key keyboard.KeyEvent) {

	fn1 := func(event keyboard.KeyEvent) {
		str := ""
		for _, c := range event.Data {
			if c == 27 {
				str += "^["
				continue
			}
			str += fmt.Sprintf("%c", c)
		}
		rx, _ := regexp.Compile(`\^\[\[([0-9]*?)\;([0-9]*?)R`)
		if rx.Match([]byte(str)) {
			arr := rx.FindStringSubmatch(str)
			row, _ := strconv.Atoi(arr[1])
			col, _ := strconv.Atoi(arr[2])
			a.cursorRow = row
			a.cursorCol = col
			a.readyCursorPos <- 1
			return
		}
		fn(event)
	}
	return fn1
}

func (a *Terminal) keyboardLoop(keysEvents <-chan keyboard.KeyEvent, err error) {
	if err != nil {
		return
	}
	fn := func(key keyboard.KeyEvent) {
		if a.keyboardFunc != nil {
			a.keyboardFunc(key)
		}
	}
	go func() {
		for {
			event := <-keysEvents
			fn(event)
		}
	}()

}
