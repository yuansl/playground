package main

import (
	"strconv"
	"strings"
	"unicode"
)

//go:generate stringer -type JsonParserStage
type JsonParserStage int

const (
	JsonParserStart JsonParserStage = iota
	InJsonString
	InJsonQuoteString
	InJsonObject
	InJsonArray
	InJsonNumber
	InJsonBool
	InJsonNull
	JsonParserStop
)

type State struct {
	stage JsonParserStage
	pos   int
}

var JsonNull any = nil

func parseJsonObject(text []byte, state *State) any {
	if text[0] != '{' {
		return nil
	}

	obj := make(map[string]any)
	i := 0
	for i < len(text) {
		var state0 State

		i += skipSpace(text[i:])
		if i >= len(text) {
			break
		}
		switch text[i] {
		case '{':
			i++
			kvs := parseJsonKeyValues(text[i:], &state0)
			for _, kv := range kvs {
				obj[kv.key] = kv.value
			}
			i += state0.pos
		case '}':
			i++
			goto end
		default:
			fatal("json_object: unexpected input:", text[i:])
		}
	}
end:
	state.pos += i
	return obj
}

func parseJsonArray(text []byte, state *State) any {
	if text[0] != '[' {
		return nil
	}
	var jarray []any
	i := 0

	for i < len(text) {
		var state0 State

		i += skipSpace(text[i:])
		if i >= len(text) {
			panic("BUG: json_array: unexpected eof")
		}
		switch text[i] {
		case '[', ',':
			i++
			v := parseJsonValue(text[i:], &state0)
			jarray = append(jarray, v)
		case ']':
			i++
			goto end
		default:
			panic("BUG: json_array: unexpected input:" + string(text[i:]))
		}
		i += state0.pos
	}
end:
	state.pos += i
	return jarray
}

func parseJsonQuoteString(text []byte, state *State) string {
	var builder strings.Builder
	var i int

	if state.stage != InJsonQuoteString {
		return ""
	}

	i = 0
	switch text[i] {
	case '\\', '\'', '"', 'a', 'b', 't', 'v', 'n', 'f', 'r', 'u', 'x', '0', '1', '2', '3', '4', '5', '6', '7':
		var c rune
		switch text[i] {
		case '\\':
			c = '\\'
		case '\'':
			c = '\''
		case '"':
			i++
			builder.WriteString("\"")
			goto end
		case 'a':
			c = '\a'
		case 'b':
			c = '\b'
		case 't':
			c = '\t'
		case 'v':
			c = '\v'
		case 'n':
			c = '\n'
		case 'f':
			c = '\f'
		case 'r':
			c = '\r'
		case 'u':
			if len(text[i+1:]) >= 4 {
				builder.WriteString("\\u")
				i++
				for count := 0; i < len(text) && unicode.In(rune(text[i]), unicode.ASCII_Hex_Digit); i++ {
					builder.WriteByte(text[i])
					count++
					if count >= 8 {
						break
					}
				}
				goto end
			} else {
				panic("BUG: json_string: invalid unicode literal sequence:" + string(text[i:]))
			}
		case 'x':
			if len(text[i+1:]) >= 2 {

			} else {
				panic("BUG: json_string: invalid hex literal sequence:" + string(text[i:]))
			}
		case '0', '1', '2', '3', '4', '5', '6', '7': // TODO
			switch text[i] {
			case '8', '9', 'A', 'b', 'c', 'd', 'e', 'f': // TODO
			}
		}
		builder.WriteRune(c)
		i++
	default:
		panic("BUG: json_string: unexpected json quote string literal:" + string(text[i:]))
	}
end:
	state.stage = InJsonString
	state.pos += i
	return builder.String()
}

func parseJsonString(text []byte, state *State) string {
	if len(text) == 0 || text[0] != '"' {
		fatal("json_string: invalid format:", text)
	}
	var builder strings.Builder
	i := 0

	for ; i < len(text); i++ {
		switch text[i] {
		case '"':
			if state.stage == JsonParserStart {
				state.stage = InJsonString
			} else if state.stage == InJsonString {
				i++
				goto end
			} else if state.stage == InJsonQuoteString {
				builder.WriteString("\"")
				state.stage = InJsonString
			} else {
				panic("BUG: json_string: unexpected stage")
			}
		case '\\':
			state.stage = InJsonQuoteString

			state0 := State{stage: InJsonQuoteString, pos: 0}
			s0 := parseJsonQuoteString(text[i+1:], &state0)
			builder.WriteString(s0)
			i += state0.pos

			state.stage = InJsonString
		default:
			if state.stage != InJsonString {
				panic("BUG: json_string: unexpected stage")
			}

			builder.WriteByte(text[i])
		}
	}
end:
	state.pos += i
	return builder.String()
}

func parseJsonNumber(text []byte, state *State) float64 {
	var numberPattern strings.Builder

	i := 0
	for ; i < len(text); i++ {
		i += skipSpace(text[i:])
		if i >= len(text) {
			panic("BUG: json_number: unexpected eof:" + string(text))
		}
		switch text[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e', '+', '-', 'E':
			numberPattern.WriteByte(text[i])
		case ',', ']', '}':
			goto end
		default:
			panic("BUG: json_number: unexpected input:" + string(text[i:]))
		}
	}
end:
	state.pos += i
	value, err := strconv.ParseFloat(numberPattern.String(), 64)
	if err != nil {
		fatal("json_number: can't parse number: ", numberPattern.String())
	}
	return value
}

func parseJsonBool(text []byte, state *State) bool {
	if len(text) == 0 {
		fatal("json_string: invalid format:", text)
	}

	var builder strings.Builder

	i := skipSpace(text)
	if i >= len(text) {
		panic("json_bool: unexpected terminated")
	}
	switch text[i] {
	case 't':
		if len(text[i:]) < 4 || string(text[i:i+4]) != "true" {
			fatal("json_bool: invalid json value 'true':", text)
		}
		builder.WriteString("true")
		i += 4
	case 'f':
		if len(text[i:]) < 5 || string(text[i:i+5]) != "false" {
			fatal("json_bool: invalid json value 'false':", text)
		}
		builder.WriteString("false")
		i += 5
	default:
		fatal("json_bool: invalid boolean:", text)
	}
	state.pos += i
	boolean := builder.String()
	if boolean == "true" {
		return true
	}
	return false
}

func parseJsonNull(text []byte, state *State) any {
	if len(text) == 0 {
		fatal("json_null: invalid input:", text)
	}
	i := skipSpace(text)
	if i >= len(text) {
		panic("BUG: json_null: unexpected eof")
	}
	switch text[i] {
	case 'n':
		if len(text[i:]) < 4 || string(text[i:i+4]) != "null" {
			fatal("json_null: invalid json null:", text)
		}
		state.pos += i + 4
	default:
		fatal("json_null: unexpected input:", text)
	}
	state.pos += i
	return JsonNull
}

func parseJsonValue(text []byte, state *State) any {
	var value any
	var state0 State

	i := skipSpace(text)
	if i >= len(text) {
		return JsonNull
	}
	switch text[i] {
	case '{':
		value = parseJsonObject(text[i:], &state0)
	case '}':
		break
	case '[':
		value = parseJsonArray(text[i:], &state0)
	case ']':
		break
	case '"':
		value = parseJsonString(text[i:], &state0)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e', '+', '-', 'E':
		value = parseJsonNumber(text[i:], &state0)
	case 't', 'f':
		value = parseJsonBool(text[i:], &state0)
	case 'n':
		value = parseJsonNull(text[i:], &state0)
	default:
		panic("BUG: json_value: unexpected input:" + string(text[i:]))
	}
	state.pos += i + state0.pos

	return value
}

type keyValue struct {
	key   string
	value any
}

func parseJsonKeyValues(text []byte, state *State) []keyValue {
	var (
		kvs []keyValue
		key string
		i   int
	)

	for i < len(text) {
		var state0 State

		i += skipSpace(text[i:])
		if i >= len(text) {
			break
		}

		switch text[i] {
		case '"':
			key = parseJsonString(text[i:], &state0)
		case ':':
			i++
			value := parseJsonValue(text[i:], &state0)
			if key != "" {
				kvs = append(kvs, keyValue{key: key, value: value})
				key = ""
			}
		case ',':
			i++
			kvs0 := parseJsonKeyValues(text[i:], &state0)
			kvs = append(kvs, kvs0...)
		case '}':
			goto end
		default:
			panic("BUG: json_key_values: unexpected input:" + string(text[i:]))
		}
		i += state0.pos
	}
end:
	if key != "" {
		panic("BUG: json_key_values: unmatched key:" + key)
	}
	state.pos += i
	return kvs
}

func parseJson(text []byte) any {
	var value any
	var state State

	if len(text) == 0 {
		return nil
	}
	if text[0] != '{' && text[0] != '[' {
		return nil
	}
	i := skipSpace(text)
	if i >= len(text) {
		return nil
	}
	switch text[i] {
	case '{':
		state.stage = InJsonObject
		value = parseJsonObject(text[i:], &state)
	case '}':
		break
	case '[':
		state.stage = InJsonArray
		value = parseJsonArray(text[i:], &state)
	case ']':
		break
	default:
		fatal("json: unexpected input:", text[i:])
	}
	return value
}
