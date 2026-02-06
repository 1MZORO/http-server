package request

import (
	"bytes"
	"fmt"
	"http-server/internal/headers"
	"io"
	"strconv"
)

type parserState string

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	Header      *headers.Headers
	Body        string
	state       parserState
}

const (
	StateInit   parserState = "init"
	StateHeader parserState = "header"
	StateBody   parserState = "body"
	StateDone   parserState = "done"
	StateError  parserState = "error"
)

var ERROR_REQUEST_LINE = fmt.Errorf("melformed request-line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsupported http-version")
var ERROR_REQUEST_IN_ERROR_STATE = fmt.Errorf("request in error state")
var SEPARATOR = []byte("\r\n")

func newRequest() *Request {
	return &Request{
		state:  StateInit,
		Header: headers.NewHeader(),
		Body:   "",
	}
}

func getInt(header *headers.Headers, name string, defaultValue int) int {
	valueStr, exits := header.Get(name)

	if !exits {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)

	if err != nil {
		return defaultValue
	}

	return value
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		currentData := data[read:]
		if len(currentData) == 0 {
			break outer
		}
		switch r.state {
		case StateError:
			return 0, ERROR_REQUEST_IN_ERROR_STATE
		case StateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				r.state = StateError
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateHeader
		case StateHeader:
			n, _, err := r.Header.Parse(currentData)

			if err != nil {
				r.state = StateError
				return 0, err
			}

			if n == 0 {
				// Check if we're at the blank line separator
				if len(currentData) >= 2 && bytes.HasPrefix(currentData, SEPARATOR) {
					read += 2
					length := getInt(r.Header, "content-length", 0)
					if length == 0 {
						r.state = StateDone
					} else {
						r.state = StateBody
					}
				} else {
					break outer
				}
			} else {
				read += n
			}

		case StateBody:
			length := getInt(r.Header, "content-length", 0)

			if length == 0 {
				r.state = StateDone
				break outer
			}

			remaining := min(length-len(r.Body), len(currentData))
			r.Body += string(currentData[:remaining])
			read += remaining

			if len(r.Body) == length {
				r.state = StateDone
			}

		case StateDone:
			break outer
		default:
			panic("State Default called!!!")
		}
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone
}

func (r *Request) error() bool {
	return r.state == StateError
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPARATOR)

	if idx == -1 {
		return nil, 0, nil
	}

	startLine := b[:idx]
	read := idx + len(SEPARATOR)

	parts := bytes.Split(startLine, []byte(" "))

	if len(parts) != 3 {
		return nil, 0, ERROR_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, 0, ERROR_REQUEST_LINE
	}
	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, read, nil

}
func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for !request.done() && !request.error() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil && err != io.EOF {
			return nil, err
		}

		if n == 0 && err == io.EOF {
			break
		}

		bufLen += n

		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
		bufLen -= readN
	}

	if request.state == StateBody {
		return nil, fmt.Errorf("unexpected EOF while reading body")
	}

	return request, nil
}
