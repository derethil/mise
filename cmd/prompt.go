package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var errNotInteractive = errors.New("cannot ask for confirmation: stdin is not a terminal")

func confirm(question string) (bool, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, errNotInteractive
	}

	fmt.Printf("%s [y/N] ", question)

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if errors.Is(err, io.EOF) {
		fmt.Println()
		return false, errNotInteractive
	}
	if err != nil {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func autoConfirm(string) (bool, error) {
	return true, nil
}

type progress struct {
	Label     string
	Status    string
	Total     int64
	Completed int64
}

type progressFunc func(progress) error

func printProgress() progressFunc {
	lastStatus := ""

	return func(p progress) error {
		if lastStatus != "" && p.Status != lastStatus {
			fmt.Println()
		}
		lastStatus = p.Status

		if p.Total > 0 {
			pct := float64(p.Completed) / float64(p.Total) * 100
			fmt.Printf("\r%s: %s %.1f%%", p.Label, p.Status, pct)
		} else {
			fmt.Printf("\r%s: %s", p.Label, p.Status)
		}

		return nil
	}
}
