/*
Copyright © 2025 Abhinav Gupta

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ABHINAVGUPTA02/SnapDB/cmd"
)

func main() {
	if len(os.Args) > 1 {
		if err := cmd.Execute(); err != nil {
			os.Exit(1)
		}
	} else {
		replMode()
	}
}

func replMode() {
	reader := bufio.NewReader(os.Stdin)
	history := make([]string, 0)
	suggestions := []string{
		"snapdb backup",
		"snapdb cleanup",
		"snapdb config init",
		"snapdb config set",
		"snapdb config show",
		"snapdb delete",
		"snapdb help",
		"snapdb list",
		"snapdb restore",
		"snapdb schedule create",
		"snapdb schedule delete",
		"snapdb schedule list",
		"snapdb stats",
		"snapdb verify",
		"snapdb version",
	}

	fmt.Println("SnapDB interactive shell. Type `snapdb help`, `history`, or `exit`.")
	for {
		fmt.Print(">>> ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			break
		}
		if line == "history" {
			printHistory(history)
			continue
		}
		if strings.HasSuffix(line, "\t") {
			printSuggestions(strings.TrimSpace(strings.TrimSuffix(line, "\t")), suggestions)
			continue
		}

		args, err := splitArgs(line)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		if len(args) == 0 {
			continue
		}
		if args[0] != "snapdb" {
			fmt.Println("Error: commands must start with 'snapdb'")
			continue
		}
		history = append(history, line)
		cmd.SetArguments(args[1:])
		if err := cmd.Execute(); err != nil {
			fmt.Println("Command failed:", err)
		}
	}
}

func splitArgs(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuote := rune(0)
	escaped := false

	for _, r := range line {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case inQuote != 0:
			if r == inQuote {
				inQuote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			inQuote = r
		case r == ' ' || r == '\t':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if escaped {
		current.WriteRune('\\')
	}
	if inQuote != 0 {
		return nil, fmt.Errorf("unterminated quote")
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

func printHistory(history []string) {
	if len(history) == 0 {
		fmt.Println("No commands in this session yet.")
		return
	}
	for i, command := range history {
		fmt.Printf("%d  %s\n", i+1, command)
	}
}

func printSuggestions(prefix string, suggestions []string) {
	matches := make([]string, 0)
	for _, suggestion := range suggestions {
		if strings.HasPrefix(suggestion, prefix) {
			matches = append(matches, suggestion)
		}
	}
	sort.Strings(matches)
	for _, match := range matches {
		fmt.Println(match)
	}
	if len(matches) == 0 {
		fmt.Println("No suggestions found.")
	}
}
