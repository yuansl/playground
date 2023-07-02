// -*- mode:go;mode:go-playground -*-
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

//go:generate stringer -type JsonParserStage
type JsonParserStage int

const (
	JsonParserStart JsonParserStage = iota
	InJsonString
	OutJsonString
	InJsonQuoteString
	InJsonObject
	OutJsonObject
	InJsonArray
	OutJsonArray
	InJsonNumber
	InJsonBool
	OutJsonBool
	InJsonNull
	OutJsonNull
	JsonParserStop
)

type State struct {
	stage JsonParserStage
	end   int
}

var JsonNull any = nil

func parseJsonObject(text string, state *State) any {
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
			state0.end++
			kvs := parseJsonKeyValues(text[i+1:], &state0)
			for _, kv := range kvs {
				obj[kv.key] = kv.value
			}
			i += state0.end
		case '}':
			i++
			goto end
		default:
			fatal("json_object: unexpected input:", text[i:])
		}
	}
end:
	state.end += i
	return obj
}

func parseJsonArray(text string, state *State) any {
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
			state0.end++
			v := parseJsonValue(text[i+1:], &state0)
			jarray = append(jarray, v)
		case ']':
			i++
			goto end
		default:
			fatal("json_array: unexpected input:", text[i:])
		}
		i += state0.end
	}
end:
	state.end += i
	return jarray
}

func parseJsonQuoteString(text string, state *State) string {
	if state.stage != InJsonQuoteString {
		return ""
	}

	var builder strings.Builder

	i := 0
	switch text[i] {
	case 'a', 'b', 't', 'v', 'n', '\'', 'f', 'r', 'u', 'x', '0', '1', '2', '3', '4', '5', '6', '7', '"':
		var c rune
		switch text[i] {
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
		case '\'':
			c = '\''
		case 'f':
			c = '\f'
		case 'r':
			c = '\r'
		case 'u':
			if len(text[i+1:]) >= 4 {
				builder.WriteString("\\u")
				builder.WriteString(text[i+1 : i+5])
				i += 4
				goto end
			} else {
				panic("BUG: json_string: invalid unicode literal sequence:" + text[i:])
			}
		case 'x':
			if len(text[i+1:]) >= 2 {

			} else {
				panic("BUG: json_string: invalid hex literal sequence:" + text[i:])
			}
		case '0', '1', '2', '3', '4', '5', '6', '7': // TODO
			switch text[i] {
			case '8', '9', 'A', 'b', 'c', 'd', 'e', 'f': // TODO
			}
		}
		builder.WriteRune(c)
		i++
	default:
		panic("BUG: json_string: unexpected json quote string literal:" + text[i:])
	}
end:
	state.stage = InJsonString
	state.end += i
	return builder.String()
}

func parseJsonString(text string, state *State) string {
	if text == "" || text[0] != '"' {
		fatal("json_string: invalid format:", text)
	}
	var builder strings.Builder
	var state0 State
	i := 0
	for ; i < len(text); i++ {
		switch text[i] {
		case '"':
			if state0.stage == JsonParserStart {
				state0.stage = InJsonString
			} else if state0.stage == InJsonString {
				state0.stage = OutJsonString
				i++
				goto end
			} else if state0.stage == InJsonQuoteString {
				builder.WriteString("\"")
				state0.stage = InJsonString
			} else {
				panic("BUG: json_string: unexpected stage")
			}
		case '\\':
			if state0.stage == InJsonQuoteString {
				state0.stage = InJsonString
				builder.WriteByte('\\')
			} else {
				state0.stage = InJsonQuoteString

				state0.end++
				s0 := parseJsonQuoteString(text[i+1:], &state0)
				builder.WriteString(s0)
				i += state0.end - 1
			}
		default:
			if state0.stage == InJsonString {
				builder.WriteByte(text[i])
			} else {
				panic("BUG: json_string: unexpected stage")
			}
		}
	}
end:
	state.end += i
	if state0.stage != OutJsonString {
		fatal("json.unmarshal error: invalid json string:" + text)
	}
	return builder.String()
}

func parseJsonNumber(text string, state *State) float64 {
	var numberPattern strings.Builder

	i := 0
	for ; i < len(text); i++ {
		i += skipSpace(text[i:])
		if i >= len(text) {
			panic("BUG: json_number: unexpected eof:" + text)
		}
		switch text[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e', '+', '-', 'E':
			numberPattern.WriteByte(text[i])
		case ',', ']', '}':
			goto end
		default:
			panic("BUG: json_number: unexpected input:" + text[i:])
		}
	}
end:
	state.end += i
	value, err := strconv.ParseFloat(numberPattern.String(), 64)
	if err != nil {
		fatal("json_number: can't parse number: ", numberPattern.String())
	}
	return value
}

func parseJsonBool(text string, state *State) bool {
	if text == "" {
		fatal("json_string: invalid format:", text)
	}

	var builder strings.Builder

	i := skipSpace(text)
	if i >= len(text) {
		panic("json_bool: unexpected terminated")
	}
	switch text[i] {
	case 't':
		if len(text[i:]) < 4 || text[i:i+4] != "true" {
			fatal("json_bool: invalid json value 'true':", text)
		}
		builder.WriteString("true")
		i += 4
	case 'f':
		if len(text[i:]) < 5 || text[i:i+5] != "false" {
			fatal("json_bool: invalid json value 'false':", text)
		}
		builder.WriteString("false")
		i += 5
	default:
		fatal("json_bool: invalid boolean:", text)
	}
	state.end += i
	boolean := builder.String()
	if boolean == "true" {
		return true
	}
	return false
}

func parseJsonNull(text string, state *State) any {
	if text == "" {
		fatal("json_null: invalid input:", text)
	}
	i := skipSpace(text)
	if i >= len(text) {
		panic("BUG: json_null: unexpected eof")
	}
	switch text[i] {
	case 'n':
		if len(text[i:]) < 4 || text[i:i+4] != "null" {
			fatal("json_null: invalid json null:", text)
		}
		state.end += i + 4
	default:
		fatal("json_null: unexpected input:", text)
	}
	state.end += i
	return JsonNull
}

func parseJsonValue(text string, state *State) any {
	var value any
	var state0 State

	i := skipSpace(text)
	if i >= len(text) {
		return JsonNull
	}
	switch text[i] {
	case '{':
		state0.stage = InJsonObject
		value = parseJsonObject(text[i:], &state0)
	case '}':
		break
	case '[':
		state0.stage = InJsonArray
		value = parseJsonArray(text[i:], &state0)
	case ']':
		break
	case '"':
		state0.stage = InJsonString
		value = parseJsonString(text[i:], &state0)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e', '+', '-', 'E':
		state0.stage = InJsonNumber
		value = parseJsonNumber(text[i:], &state0)
	case 't', 'f':
		state0.stage = InJsonBool
		value = parseJsonBool(text[i:], &state0)
	case 'n':
		state0.stage = InJsonNull
		value = parseJsonNull(text[i:], &state0)
	default:
		panic("BUG: json_value: unexpected input:" + text[i:])
	}
	i += state0.end
	state.end += i
	return value
}

type keyValue struct {
	key   string
	value any
}

func parseJsonKeyValues(text string, state *State) []keyValue {
	var kvs []keyValue

	key := ""
	i := 0
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
			state0.end += 1
			value := parseJsonValue(text[i+1:], &state0)
			if key != "" {
				kvs = append(kvs, keyValue{key: key, value: value})
				key = ""
			}
		case ',':
			state0.end += 1
			kvs0 := parseJsonKeyValues(text[i+1:], &state0)
			kvs = append(kvs, kvs0...)
		case ']', '}':
			goto end
		default:
			panic("BUG: json_key_values: unexpected input:" + text[i:])
		}
		i += state0.end
	}
end:
	if key != "" {
		panic("BUG: json_key_values: unmatched key:" + key)
	}
	state.end += i
	return kvs
}

func parseJson(text string) any {
	if text == "" {
		return nil
	}

	if text[0] != '{' && text[0] != '[' {
		return nil
	}

	var value any

	state := State{stage: JsonParserStart}

	i := 0
	i += skipSpace(text[i:])
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
		if state.stage == JsonParserStart {
			state.stage = InJsonArray
		}
		value = parseJsonArray(text[i:], &state)
	case ']':
		break
	default:
		fatal("json: unexpected input:", text[i:])
	}
	return value
}

func main() {
	rawjson := `{"name": "lio", "age": 26, "female": false, "extra": null, "points": [1,2,3,4,{"attr": "address", "location": "beijing"}, null, null, null]}`

	fp, err := os.Open("/home/yuansl/.cache/mintinstall/reviews.json")
	if err != nil {
		fatal("os.Open error:", err)
	}
	defer fp.Close()

	data, err := io.ReadAll(io.LimitReader(fp, 30<<20))
	if err != nil {
		fatal("io.ReadAll error:", err)
	}

	s := parseJson(string(data))

	fmt.Printf("parsing json R('%s') :\ns='%v'\n", rawjson, s)
}
