package parser

import (
	"fmt"
	"os"
	"strings"
)

type TokenType int

const (
	TokenWord TokenType = iota
	TokenRedirectOut
	TokenRedirectOutAppend
	TokenRedirectErrAppend
	TokenRedirectIn
	TokenRedirectErr
)

type Token struct {
	Type  TokenType
	Value string
}
type Redirect struct {
	FD   int
	Path string
	Mode int
}
type Command struct {
	Args      []string
	Redirects []Redirect
}

func ParseCommand(input string) (*Command, error) {
	tokens, err := lex(input)
	if err != nil {
		return nil, err
	}

	command, err := parser(tokens)
	if err != nil {
		return nil, err
	}

	return command, nil
}

func lex(input string) ([]Token, error) {
	var tokens []Token
	var arg strings.Builder

	var quote rune
	flushArg := func() {
		if arg.Len() == 0 {
			return
		}

		tokens = append(tokens, Token{
			Type:  TokenWord,
			Value: arg.String(),
		})

		arg.Reset()
	}

	runes := []rune(input)

	for i := 0; i < len(runes); {
		r := runes[i]

		if quote == '\'' {
			if r == '\'' {
				quote = 0
			} else {
				arg.WriteRune(r)
			}

			i++
			continue
		}

		if quote == '"' {
			if r == '"' {
				quote = 0
				i++
				continue
			}

			if r == '\\' {
				if i+1 >= len(runes) {
					return nil, fmt.Errorf(
						"syntax error: trailing '\\'",
					)
				}

				next := runes[i+1]

				switch next {
				case '\\', '"':
					arg.WriteRune(next)

				default:
					arg.WriteRune('\\')
					arg.WriteRune(next)
				}

				i += 2
				continue
			}

			arg.WriteRune(r)
			i++
			continue
		}

		if r >= '0' && r <= '9' &&
			i+1 < len(runes) &&
			runes[i+1] == '>' &&
			arg.Len() == 0 {

			fd := int(r - '0')

			if fd > 2 {
				return nil, fmt.Errorf(
					"unsupported file descriptor: %d",
					fd,
				)
			}

			tokenType := TokenRedirectOut
			length := 2

			switch fd {
			case 0:
				tokenType = TokenRedirectIn
			case 1:
				tokenType = TokenRedirectOut
			case 2:
				tokenType = TokenRedirectErr
			}

			if i+2 < len(runes) && runes[i+2] == '>' {
				if fd == 0 {
					return nil, fmt.Errorf("append redirection is invalid for stdin")
				}

				if fd == 1 {
					tokenType = TokenRedirectOutAppend
				} else {
					tokenType = TokenRedirectErrAppend
				}

				length = 3
			}

			tokens = append(tokens, Token{
				Type:  tokenType,
				Value: string(runes[i : i+length]),
			})

			i += length
			continue
		}

		switch r {
		case '\'':
			quote = '\''

		case '"':
			quote = '"'

		case '\\':
			if i+1 >= len(runes) {
				arg.WriteRune('\\')
				i++
				continue
			}

			arg.WriteRune(runes[i+1])
			i += 2
			continue

		case ' ', '\t', '\r', '\n':
			flushArg()

		case '>':
			flushArg()

			if i+1 < len(runes) && runes[i+1] == '>' {
				tokens = append(tokens, Token{
					Type:  TokenRedirectOutAppend,
					Value: ">>",
				})

				i += 2
				continue
			}

			tokens = append(tokens, Token{
				Type:  TokenRedirectOut,
				Value: ">",
			})

		case '<':
			flushArg()

			tokens = append(tokens, Token{
				Type:  TokenRedirectIn,
				Value: "<",
			})

		default:
			if r == '2' &&
				i+1 < len(runes) &&
				runes[i+1] == '>' &&
				arg.Len() == 0 {

				tokens = append(tokens, Token{
					Type:  TokenRedirectErr,
					Value: "2>",
				})

				i += 2
				continue
			}

			arg.WriteRune(r)
		}

		i++
	}

	if quote != 0 {
		return nil, fmt.Errorf("syntax error: unterminated quote")
	}

	flushArg()

	return tokens, nil
}

func parser(tokens []Token) (*Command, error) {
	command := &Command{}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		switch token.Type {
		case TokenWord:
			command.Args = append(command.Args, token.Value)

		case TokenRedirectOut:
			path, next, err := getRedirectPath(tokens, i)
			if err != nil {
				return nil, err
			}

			command.Redirects = append(
				command.Redirects,
				Redirect{
					FD:   1,
					Path: path,
					Mode: os.O_CREATE |
						os.O_WRONLY |
						os.O_TRUNC,
				},
			)

			i = next

		case TokenRedirectOutAppend:
			path, next, err := getRedirectPath(tokens, i)
			if err != nil {
				return nil, err
			}

			command.Redirects = append(
				command.Redirects,
				Redirect{
					FD:   1,
					Path: path,
					Mode: os.O_CREATE |
						os.O_WRONLY |
						os.O_APPEND,
				},
			)

			i = next

		case TokenRedirectIn:
			path, next, err := getRedirectPath(tokens, i)
			if err != nil {
				return nil, err
			}

			command.Redirects = append(
				command.Redirects,
				Redirect{
					FD:   0,
					Path: path,
					Mode: os.O_RDONLY,
				},
			)

			i = next

		case TokenRedirectErr:
			path, next, err := getRedirectPath(tokens, i)
			if err != nil {
				return nil, err
			}

			command.Redirects = append(
				command.Redirects,
				Redirect{
					FD:   2,
					Path: path,
					Mode: os.O_CREATE |
						os.O_WRONLY |
						os.O_TRUNC,
				},
			)

			i = next

		case TokenRedirectErrAppend:
			path, next, err := getRedirectPath(tokens, i)
			if err != nil {
				return nil, err
			}

			command.Redirects = append(
				command.Redirects,
				Redirect{
					FD:   2,
					Path: path,
					Mode: os.O_CREATE |
						os.O_WRONLY |
						os.O_APPEND,
				},
			)

			i = next
		}
	}

	if len(command.Args) == 0 {
		return nil, fmt.Errorf("syntax error: empty command")
	}

	return command, nil
}

func getRedirectPath(tokens []Token, index int) (string, int, error) {
	if index+1 >= len(tokens) {
		return "", index, fmt.Errorf(
			"syntax error: expected file after '%s'",
			tokens[index].Value,
		)
	}

	next := tokens[index+1]

	if next.Type != TokenWord {
		return "", index, fmt.Errorf(
			"syntax error: expected file after '%s'",
			tokens[index].Value,
		)
	}

	return next.Value, index + 1, nil
}
