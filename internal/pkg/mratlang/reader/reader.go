package reader

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
)

type trackingRuneScanner struct {
	rs io.RuneScanner

	filename       string
	nextRuneLine   int
	nextRuneColumn int

	// keep track of the last two runes read, most recent last.
	history []ast.Pos
}

func newTrackingRuneScanner(rs io.RuneScanner, filename string) *trackingRuneScanner {
	if filename == "" {
		filename = "<unknown-file>"
	}
	return &trackingRuneScanner{
		rs:             rs,
		filename:       filename,
		nextRuneLine:   1,
		nextRuneColumn: 1,
		history:        make([]ast.Pos, 0, 2),
	}
}

func (r *trackingRuneScanner) ReadRune() (rune, int, error) {
	rn, size, err := r.rs.ReadRune()
	if err != nil {
		return rn, size, err
	}
	if len(r.history) == 2 {
		r.history[0] = r.history[1]
		r.history = r.history[:1]
	}
	r.history = append(r.history, ast.Pos{
		Filename: r.filename,
		Line:     r.nextRuneLine,
		Column:   r.nextRuneColumn,
	})
	if rn == '\n' {
		r.nextRuneLine++
		r.nextRuneColumn = 1
	} else {
		r.nextRuneColumn++
	}
	return rn, size, nil
}

func (r *trackingRuneScanner) UnreadRune() error {
	err := r.rs.UnreadRune()
	if err != nil {
		return err
	}
	if len(r.history) == 0 {
		panic("UnreadRune called when history is empty")
	}
	lastPos := r.history[len(r.history)-1]
	r.history = r.history[:len(r.history)-1]
	r.nextRuneLine = lastPos.Line
	r.nextRuneColumn = lastPos.Column
	return nil
}

// pos returns the position of the next rune that will be read.
func (r *trackingRuneScanner) pos() ast.Pos {
	if len(r.history) == 0 {
		return ast.Pos{
			Filename: r.filename,
			Line:     r.nextRuneLine,
			Column:   r.nextRuneColumn,
		}
	}
	return r.history[len(r.history)-1]
}

type Reader struct {
	rs *trackingRuneScanner

	posStack []ast.Pos
}

type options struct {
	filename string
}

// Option represents an option that can be passed to New.
type Option func(*options)

// WithFilename sets the filename to be associated with the input.
func WithFilename(filename string) Option {
	return func(o *options) {
		o.filename = filename
	}
}

func New(r io.RuneScanner, opts ...Option) *Reader {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &Reader{
		rs: newTrackingRuneScanner(r, o.filename),
	}
}

// Read reads all expressions from the input until a read error occurs
// or io.EOF is reached. A final io.EOF will not be returned if the
// input ends with a valid expression or if it contains no expressions
// at all.
func (r *Reader) ReadAll() ([]ast.Node, error) {
	var nodes []ast.Node
	for {
		_, err := r.next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, r.error("error reading input: %w", err)
		}
		r.rs.UnreadRune()
		node, err := r.readExpr()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// error returns a formatted error that includes the current position
// of the scanner.
func (r *Reader) error(format string, args ...interface{}) error {
	pos := r.rs.pos()
	prefix := fmt.Sprintf("%s:%d:%d: ", pos.Filename, pos.Line, pos.Column)
	return fmt.Errorf(prefix+format, args...)
}

// popSection returns the last section read, ending at the current
// input, and pops it off the stack.
func (r *Reader) popSection() ast.Section {
	sec := ast.Section{
		StartPos: r.posStack[len(r.posStack)-1],
		EndPos:   r.rs.pos(),
	}
	r.posStack = r.posStack[:len(r.posStack)-1]
	return sec
}

// pushSection pushes a new section onto the stack, starting at the
// current input.
func (r *Reader) pushSection() {
	r.posStack = append(r.posStack, r.rs.pos())
}

// next returns the next rune that is not whitespace or a comment.
func (r *Reader) next() (rune, error) {
	for {
		rn, _, err := r.rs.ReadRune()
		if err != nil {
			return 0, r.error("error reading input: %w", err)
		}
		if unicode.IsSpace(rn) {
			continue
		}
		if rn == ';' {
			for {
				rn, _, err := r.rs.ReadRune()
				if err != nil {
					return 0, r.error("error reading input: %w", err)
				}
				if rn == '\n' {
					break
				}
			}
			continue
		}
		return rn, nil
	}
}

func (r *Reader) readExpr() (ast.Node, error) {
	rune, err := r.next()
	if err != nil {
		return nil, err
	}

	r.pushSection()
	switch rune {
	case '(':
		return r.readList()
	case ')':
		return nil, r.error("unexpected ')'")
	case '"':
		return r.readString()
	case '\'':
		return r.readQuote()
	case '`':
		return nil, r.error("quasiquote not implemented")
	case ',':
		return nil, r.error("unquote not implemented")
	case '#':
		return r.readDispatch()
	case ':':
		return r.readKeyword()
	default:
		r.rs.UnreadRune()
		return r.readSymbol()
	}
}

func (r *Reader) readList() (ast.Node, error) {
	var nodes []ast.Node
	for {
		rune, err := r.next()
		if err != nil {
			return nil, err
		}
		if unicode.IsSpace(rune) {
			continue
		}
		if rune == ')' {
			break
		}

		r.rs.UnreadRune()
		node, err := r.readExpr()
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return ast.NewList(nodes, r.popSection()), nil
}

func (r *Reader) readString() (ast.Node, error) {
	var str string
	for {
		rune, _, err := r.rs.ReadRune()
		if err != nil {
			return nil, r.error("error reading string: %w", err)
		}
		// handle escape sequences
		if rune == '\\' {
			rune, _, err = r.rs.ReadRune()
			if err != nil {
				return nil, r.error("error reading string: %w", err)
			}
			switch rune {
			case 'n':
				rune = '\n'
			case 't':
				rune = '\t'
			case 'r':
				rune = '\r'
			case '"':
				rune = '"'
			case '\\':
				rune = '\\'
			default:
				return nil, r.error("invalid escape sequence: \\%c", rune)
			}
		} else if rune == '"' {
			break
		}
		str += string(rune)
	}
	return ast.NewString(str, r.popSection()), nil
}

func (r *Reader) readQuote() (ast.Node, error) {
	node, err := r.readExpr()
	if err != nil {
		return nil, err
	}
	section := r.popSection()
	items := []ast.Node{
		ast.NewSymbol("quote", ast.Section{StartPos: section.StartPos, EndPos: node.Pos()}),
		node,
	}
	return ast.NewList(items, section), nil
}

func (r *Reader) readDispatch() (ast.Node, error) {
	rn, _, err := r.rs.ReadRune()
	if err != nil {
		return nil, r.error("error reading dispatch: %w", err)
	}
	switch rn {
	case '(':
		return nil, r.error("vector dispatch not implemented")
	case 't':
		return ast.NewBool(true, r.popSection()), nil
	case 'f':
		return ast.NewBool(false, r.popSection()), nil
	case '\\':
		return nil, r.error("character dispatch not implemented")
	case '"':
		return nil, r.error("string dispatch not implemented")
	default:
		return nil, r.error("invalid dispatch: #%c", rn)
	}
}

func (r *Reader) readSymbol() (ast.Node, error) {
	var sym string
	for {
		rn, _, err := r.rs.ReadRune()
		if err != nil {
			return nil, r.error("error reading symbol: %w", err)
		}
		if unicode.IsSpace(rn) || rn == ')' {
			r.rs.UnreadRune()
			break
		}
		sym += string(rn)
	}
	// check if symbol is a number
	if num, err := strconv.ParseFloat(sym, 64); err == nil {
		return ast.NewNumber(num, r.popSection()), nil
	}

	return ast.NewSymbol(sym, r.popSection()), nil
}

func (r *Reader) readKeyword() (ast.Node, error) {
	var sym string
	for {
		rn, _, err := r.rs.ReadRune()
		if err != nil {
			return nil, r.error("error reading keyword: %w", err)
		}
		if unicode.IsSpace(rn) || rn == ')' {
			r.rs.UnreadRune()
			break
		}
		sym += string(rn)
	}
	return ast.NewKeyword(sym, r.popSection()), nil
}
