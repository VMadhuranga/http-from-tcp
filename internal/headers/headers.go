package headers

import (
	"errors"
	"strings"
)

const crlf = "\r\n"

var validHeaderKeyChars = map[rune]struct{}{
	'a': {}, 'b': {}, 'c': {}, 'd': {}, 'e': {}, 'f': {}, 'g': {}, 'h': {}, 'i': {}, 'j': {}, 'k': {}, 'l': {}, 'm': {}, 'n': {}, 'o': {}, 'p': {}, 'q': {}, 'r': {}, 's': {}, 't': {}, 'u': {}, 'v': {}, 'w': {}, 'x': {}, 'y': {}, 'z': {},
	'0': {}, '1': {}, '2': {}, '3': {}, '4': {}, '5': {}, '6': {}, '7': {}, '8': {}, '9': {},
	'!': {}, '#': {}, '$': {}, '%': {}, '&': {}, '\'': {}, '*': {}, '+': {}, '-': {}, '.': {}, '^': {}, '_': {}, '`': {}, '|': {}, '~': {},
}

func checkHeaderKeyValidity(headerKey string) bool {
	if len(headerKey) < 1 {
		return false
	}

	if strings.Count(headerKey, " ") > 0 {
		return false
	}

	for _, v := range headerKey {
		_, ok := validHeaderKeyChars[v]
		if !ok {
			return false
		}
	}

	return true
}

type Headers map[string]string

func (h Headers) Parse(rawHeader string) (n int, done bool, err error) {
	crlfIDX := strings.Index(rawHeader, crlf)
	if crlfIDX == -1 {
		return 0, false, nil
	}
	if crlfIDX == 0 {
		return len(crlf), true, nil
	}

	header := rawHeader[:crlfIDX]
	headerKey, headerVal, ok := strings.Cut(strings.TrimSpace(header), ":")
	if !ok {
		return 0, false, errors.New("invalid header")
	}

	headerKey = strings.ToLower(headerKey)
	if !checkHeaderKeyValidity(headerKey) {
		return 0, false, errors.New("invalid header key")
	}

	if strings.Count(headerVal, " ") > 1 {
		return 0, false, errors.New("invalid header value")
	}

	val, ok := h[headerKey]
	headerVal = strings.TrimSpace(headerVal)
	if ok {
		if val == headerVal {
			h[headerKey] = headerVal
		} else {
			h[headerKey] = val + ", " + headerVal
		}
	} else {
		h[headerKey] = headerVal
	}

	return len(header + crlf), false, nil
}

func (h Headers) Get(headerKey string) string {
	return h[strings.ToLower(headerKey)]
}

func NewHeaders() Headers {
	return make(Headers)
}
