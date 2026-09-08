package cliutil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

var ErrNotInteractive = errors.New("cannot ask for confirmation: stdin is not a terminal")

func Confirm(question string) (bool, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, ErrNotInteractive
	}

	fmt.Printf("%s [y/N] ", question)

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if errors.Is(err, io.EOF) {
		fmt.Println()
		return false, ErrNotInteractive
	}
	if err != nil {
		return false, err
	}

	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func AutoConfirm(string) (bool, error) {
	return true, nil
}

type Progress struct {
	Label     string
	Status    string
	Total     int64
	Completed int64
}

type ProgressFunc func(Progress) error

func PrintProgress() ProgressFunc {
	lastStatus := ""

	return func(p Progress) error {
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
