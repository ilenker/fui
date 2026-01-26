package fui

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"github.com/ilenker/fui/internal/calc"
	tc "github.com/gdamore/tcell/v3"
	color "github.com/gdamore/tcell/v3/color"
)

type mdFormatMask uint8
const (
	BOLD    mdFormatMask = 1<<0
	ITALIC  mdFormatMask = 1<<1
	LINK    mdFormatMask = 1<<2
	HEADING mdFormatMask = 1<<3
	LIST    mdFormatMask = 1<<4
	HILITE  mdFormatMask = 1<<7
)

func (f *mdFormatMask) String() string {
	str := ""
	if BOLD & *f == BOLD {
		str += "BOLD "
	}
	if ITALIC & *f == ITALIC {
		str += "ITALIC "
	}
	if LINK & *f == LINK {
		str += "LINK "
	}
	if HEADING & *f == HEADING {
		str += "HEADING "
	}
	if LIST & *f == LIST {
		str += "LIST "
	}
	if HILITE & *f ==  HILITE {
		str += "HILITE "
	}
	return str
}

func WriteMarkdown(id int, termData *TerminalData, inputBuf []byte) {
	if len(inputBuf) == 0 {
		return
	}

	////////////////////////////////////
	// Declarations, Helper definitions
	destBuf := &termData.Buffers[id]
	formats := &termData.AttrRanges[id]
	var linkURL string
	var mask mdFormatMask
	var indentLevel int
	var strictCommonMark bool
	var appendNewlineDouble = func(entering bool) {
		if !entering {
			*destBuf = append(*destBuf, '\n', '\n')
		}
	}
	var appendNewline = func(entering bool) {
		if !entering {
			*destBuf = append(*destBuf, '\n')
		}
	}
	var checkLineBreak = func(node *ast.Text) {
		if node.SoftLineBreak() {
			if strictCommonMark {
				*destBuf = append(*destBuf, ' ')
			} else {
				// Maintain indentation level when forcing a break
				*destBuf = append(*destBuf, '\n')
				for range indentLevel {
					*destBuf = append(*destBuf, []byte("\t")...)
				}
			}
		}
		if node.HardLineBreak() {
			*destBuf = append(*destBuf, '\n', '\n')
		}
	}
	var appendAndAddRange = func(node *ast.Text, mask mdFormatMask) {
		start := len(*destBuf)
		*destBuf = append(*destBuf, node.Segment.Value(inputBuf)...)
		if mask == 0 {
			checkLineBreak(node)
			return
		}
		end := len(*destBuf)
		termData.AppendAttrRange(id, start, end, mask)
		if mask & LINK != 0 {
			formats.Links[formats.idx] = string(linkURL)
			return
		}
		checkLineBreak(node)
	}

	///////////////////////
	// Goldmark AST setup
	md := goldmark.New()
	reader := text.NewReader(inputBuf)
	doc := md.Parser().Parse(reader)

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		switch n := n.(type) {
		case *ast.Heading:
			mask ^= HEADING
			appendNewlineDouble(entering)
		case *ast.TextBlock, *ast.Paragraph: //TextBlock
			// Indent when entering a new paragraph (line of actual text with follow in next nodes)
			if entering {
				for range indentLevel {
					*destBuf = append(*destBuf, []byte("\t")...)
				}
			}
			if indentLevel == 0 {
				appendNewlineDouble(entering)
			} else {
				appendNewline(entering)
			}
		case *ast.List:
			// TODO(Johan): replace this with normal code
			indentLevel += (calc.FB2i(entering)*2)-1
			if indentLevel == 0 {
				appendNewline(entering)
			}
		case *ast.Link:
			mask ^= LINK
			linkURL = string(n.Destination)
		case *ast.Emphasis:
			switch n.Level {
			case 1:
				mask ^= ITALIC
			case 2:
				mask ^= BOLD
			}
		case *ast.ListItem:
		case *ast.Text:
			if !entering {
				return ast.WalkContinue, nil
			}
			appendAndAddRange(n, mask)
		}
		return ast.WalkContinue, nil
	})
}

func ProcessAttr(mask mdFormatMask) (mdFormatMask, bool) {
	if mask & LINK != 0 {
		return mask ^ LINK, true
	}
	if mask & HEADING != 0 {
		return mask ^ HEADING, true
	}
	return mask, false
}

const COLOR_LINK = 0xC20E68
const COLOR_HEADING = 0xE36840
const COLOR_LIST = 0x76E3A9

func AttrToStyle(mask mdFormatMask) (tc.Style) {
	styleMask := tc.AttrMask(0)
	col := color.White
	underline := tc.UnderlineStyle(0)
	if mask & BOLD != 0 {
		styleMask |= tc.AttrBold
	}
	if mask & ITALIC != 0 {
		styleMask |= tc.AttrItalic
	}
	if mask & LINK != 0 {
		col = color.NewHexColor(0xC20E68)
		underline |= tc.UnderlineStyleSolid
	}
	if mask & HEADING != 0 {
		col = color.NewHexColor(0xE36840)
		styleMask |= tc.AttrBold
	}
	if mask & LIST != 0 {
		col = color.NewHexColor(0x76E3A9)
	}
	if mask & HILITE != 0 {
		styleMask |= tc.AttrReverse
	}
	return tc.StyleDefault.Foreground(col).Attributes(styleMask).Underline(underline)
}
