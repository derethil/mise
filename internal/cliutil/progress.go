package cliutil

import "fmt"

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
