package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ilenker/fui"
	"github.com/mattn/go-runewidth"
)

var term *fui.Box

func main() {
	fui.Init()

	// Calculator setup
	term = fui.Terminal("Simple Calculator")
	for i := range 10 {
		fui.Button(string(byte(i+'0')), func(b *fui.Box) {
			calculator(rune(i + '0'))
		})
	}
	fui.Button("＋", func(b *fui.Box) {
		calculator('＋')
	})
	fui.Button("－", func(b *fui.Box) {
		calculator('－')
	})
	fui.Button("＊", func(b *fui.Box) {
		calculator('＊')
	})
	fui.Button("／", func(b *fui.Box) {
		calculator('／')
	})
	fui.Button(".", func(b *fui.Box) {
		calculator('.')
	})
	fui.Button("＝", func(b *fui.Box) {
		calculator('＝')
	})
	fui.Button("Run Test", func(b *fui.Box) {
		runNextTest()
	})

	// Spawn some watchers
	fui.Watcher("result", &result)
	fui.Watcher("op", &op)
	fui.Watcher("lastOp", &lastOp)
	fui.Watcher("lastRhs", &lastRhs)
	fui.Watcher("period", &period)

	go fui.Start()

	<-fui.ExitSig
}

// Calculator ---------------------------------------------------------------------------------
var result = math.Inf(1)
var op rune
var lastOp rune
var lastRhs float64
var period bool

func calculator(r rune) {
	var lhs, rhs float64
	switch {
	// Numbers
	case r >= '0' && r <= '9':
		// Reset case (no op and no result)
		if result != math.Inf(1) && op == 0 {
			term.Write("\n")
			result = math.Inf(1)
		}
		term.Write(string(r))
		return
	case r == '.':
		if period {
			return
		}
		if result != math.Inf(1) && op == 0 {
			term.Write("\n")
			result = math.Inf(1)
		}
		period = true
		term.Write(string(r))
		return
	case r == '＝':
		if result != math.Inf(1) && op == 0 {
			op = lastOp
			rhs = lastRhs
			term.Write(string(lastOp))
			term.Write(fmt.Sprintf("%.5g", rhs))
		}
	default:
		// New op should replace the previous op
		if op != 0 {
			op = r
			term.Backspace()
			term.Write(string(r))
			return
		}
		op = r
		period = false
		term.Write(string(r))
		return
	}

	// TODO: fix +Inf parsing
	line := term.Line(-1)
	i := strings.IndexAny(line, "＋－＊／")
	offset := runewidth.RuneWidth('＋')
	if i == -1 {
		return
	}
	if result == math.Inf(1) {
		lhs, _ = strconv.ParseFloat(line[:i], 64)
		result = lhs
	}
	if rhs == 0 {
		rhs, _ = strconv.ParseFloat(line[i+1+offset:], 64)
	}
	if op == 0 {
		op, _ = utf8.DecodeRuneInString(line[i:])
	}

	switch op {
	case '＋':
		result = result + rhs
	case '－':
		result = result - rhs
	case '＊':
		result = result * rhs
	case '／':
		result = result / rhs
	default:
	}
	term.Write(fmt.Sprintf("\n＝%.5g", result))
	lastOp = op
	lastRhs = rhs
	period = false
	op = 0
}
