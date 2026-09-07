package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

// ---------------------------------------------------------------- raw terminal

type termios struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Line   uint8
	Cc     [32]uint8
	Ispeed uint32
	Ospeed uint32
}

const (
	tcgets  = 0x5401
	tcsets  = 0x5402
	tcgwinsz = 0x5413
)

var savedTermios *termios

func isTerminal(fd int) bool {
	var t termios
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(tcgets), uintptr(unsafe.Pointer(&t)))
	return errno == 0
}

func termRaw(fd int) {
	if savedTermios != nil {
		return
	}
	var old termios
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(tcgets), uintptr(unsafe.Pointer(&old))); errno != 0 {
		return
	}
	savedTermios = &old
	raw := old
	raw.Iflag &^= 0x0001 | 0x0002 | 0x0004 | 0x0008 | 0x0010 | 0x0020 | 0x0040 | 0x0080 | 0x0100 // IGNBRK..IXON
	raw.Oflag &^= 0x0001 // OPOST
	raw.Lflag &^= 0x0001 | 0x0002 | 0x0008 | 0x0040 | 0x8000       // ECHO, ECHONL, ICANON, ISIG, IEXTEN
	raw.Cc[0] = 1                                                 // VMIN
	raw.Cc[1] = 0                                                 // VTIME
	_ = ioctlSet(fd, &raw)
}

func ioctlSet(fd int, t *termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(tcsets), uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return errno
	}
	return nil
}

func termRestore(fd int) {
	if savedTermios != nil {
		_ = ioctlSet(fd, savedTermios)
		savedTermios = nil
	}
}

func termSize(fd int) (w, h int) {
	var ws struct {
		Row, Col       uint16
		Xpixel, Ypixel uint16
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(tcgwinsz), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
}

// ---------------------------------------------------------------- ANSI helpers

const (
	escReset  = "\x1b[0m"
	escBold   = "\x1b[1m"
	escDim    = "\x1b[2m"
	escUnder  = "\x1b[4m"
	escRev    = "\x1b[7m"
	escBlack  = "\x1b[30m"
	escRed    = "\x1b[31m"
	escGreen  = "\x1b[32m"
	escYellow = "\x1b[33m"
	escBlue   = "\x1b[34m"
	escMag    = "\x1b[35m"
	escCyan   = "\x1b[36m"
	escWhite  = "\x1b[37m"
)

func c256(n int) string { return "\x1b[38;5;" + itoa(n) + "m" }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

func ansi(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\x1b':
			// copy escape sequence through
			j := i
			for j < len(s) && (s[j] < 'A' || s[j] > 'z' || s[j] == '[') && s[j] != 'm' {
				j++
				if j < len(s) && s[j] == '[' {
					j++
				}
			}
			if j < len(s) && s[j] == 'm' {
				j++
			}
			sb.WriteString(s[i:j])
			i = j - 1
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

func stripAnsi(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			i++
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

// renderLine clips a line to width, keeping ANSI sequences intact.
func renderLine(s string, width int) string {
	s = stripAnsi(s)
	if len(s) > width {
		s = s[:width]
	}
	return s
}

func pad(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	return s + strings.Repeat(" ", w-len(s))
}

// ---------------------------------------------------------------- keys

type key int

const (
	keyNone key = iota
	keyUp
	keyDown
	keyLeft
	keyRight
	keyPgUp
	keyPgDn
	keyHome
	keyEnd
	keyEnter
	keyEsc
	keyTab
	keyBackspace
	keyRune
)

var stdinReader = bufio.NewReader(os.Stdin)

func readKey() (key, rune) {
	b, err := stdinReader.ReadByte()
	if err != nil {
		return keyNone, 0
	}
	if b == 0x1b {
		b2, err := stdinReader.ReadByte()
		if err != nil {
			return keyEsc, 0
		}
		if b2 != '[' {
			return keyEsc, 0
		}
		b3, err := stdinReader.ReadByte()
		if err != nil {
			return keyEsc, 0
		}
		switch b3 {
		case 'A':
			return keyUp, 0
		case 'B':
			return keyDown, 0
		case 'C':
			return keyRight, 0
		case 'D':
			return keyLeft, 0
		case 'H':
			return keyHome, 0
		case 'F':
			return keyEnd, 0
		case '5':
			stdinReader.ReadByte() // ~
			return keyPgUp, 0
		case '6':
			stdinReader.ReadByte() // ~
			return keyPgDn, 0
		case '1':
			stdinReader.ReadByte() // ~
			return keyHome, 0
		case '4':
			stdinReader.ReadByte() // ~
			return keyEnd, 0
		}
		return keyNone, 0
	}
	switch b {
	case '\r', '\n':
		return keyEnter, 0
	case '\t':
		return keyTab, 0
	case 0x7f, 0x08:
		return keyBackspace, 0
	}
	return keyRune, rune(b)
}

// readLine collects printable runes until Enter or Esc.
func readLine(prompt string) (string, bool) {
	out := os.Stdout
	out.WriteString("\x1b[?25h") // show cursor
	defer out.WriteString("\x1b[?25l")
	var buf []rune
	for {
		k, r := readKey()
		switch k {
		case keyEnter:
			return string(buf), true
		case keyEsc:
			return "", false
		case keyBackspace:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				out.WriteString("\x1b[D\x1b[K")
			}
		case keyRune:
			buf = append(buf, r)
			fmt.Fprint(out, string(r))
		}
	}
}
