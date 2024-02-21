package ircd

import (
	"errors"
	"strings"
)

func parseMessage(line string) (message, error) {
	if len(line) == 0 {
		return message{}, nil
	}

	if len(line) > 512 {
		return message{}, errors.New("message too long")
	}

	message := message{
		Raw:     line,
		Tags:    make(map[string]interface{}),
		Prefix:  "",
		Command: "",
		Params:  make([]string, 0),
	}

	pos := 0
	next := 0

	if line[0] == '@' {
		next = strings.IndexByte(line, ' ')

		if next == -1 {
			return message, errors.New("malformed message")
		}

		rawTags := strings.Split(line[1:next], ";")

		for _, tag := range rawTags {
			pair := strings.SplitN(tag, "=", 2)

			if len(pair) != 2 {
				break
			}

			message.Tags[pair[0]] = pair[1]
			if len(pair) == 1 {
				message.Tags[pair[0]] = true
			}
		}

		pos = next + 1
	}

	for pos < len(line) && line[pos] == ' ' {
		pos++
	}

	if pos < len(line) && line[pos] == ':' {
		next = strings.IndexByte(line[pos:], ' ')

		if next == -1 {
			return message, errors.New("malformed message")
		}

		message.Prefix = line[pos+1 : pos+next]
		pos += next + 1

		for pos < len(line) && line[pos] == ' ' {
			pos++
		}
	}

	if line[pos] == ':' {
		message.Params = append(message.Params, line[pos+1:])
		return message, nil
	}

	next = strings.IndexByte(line[pos:], ' ')

	if next == -1 {
		if len(line) > pos {
			cmd := line[pos:]
			message.Command = strings.ToUpper(cmd)
			return message, nil
		}

		return message, errors.New("malformed message")
	}

	message.Command = line[pos : pos+next]
	pos += next + 1

	for pos < len(line) {
		if line[pos] == ':' {
			message.Params = append(message.Params, line[pos+1:])
			break
		}

		if line[pos] == ' ' {
			pos++
			continue
		}

		next = strings.IndexByte(line[pos:], ' ')

		if next != -1 {
			message.Params = append(message.Params, line[pos:pos+next])
			pos += next + 1

			for pos < len(line) && line[pos] == ' ' {
				pos++
			}

			continue
		}

		if next == -1 {
			message.Params = append(message.Params, line[pos:])
			break
		}
	}

	return message, nil
}
