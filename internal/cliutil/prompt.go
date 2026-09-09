package cliutil

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/manifoldco/promptui"
)

func AutoConfirm(string) (bool, error) {
	return true, nil
}

func Confirm(question string) (bool, error) {
	return requestConfirmation(question)
}

func ConfirmOrDie(question string) {
	confirmed, err := requestConfirmation(question)
	if err == nil && confirmed {
		return
	}

	slog.Info("Aborted.", "question", question, "error", err)
	os.Exit(1)
}

func SelectOption(question string, options []string) (string, error) {
	prompt := promptui.Select{
		Label: question,
		Items: options,
	}

	_, result, err := prompt.Run()
	if err != nil {
		slog.Error("select option failed", "error", err)
		return "", err
	}

	return result, nil
}

func requestConfirmation(question string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     question,
		IsConfirm: true,
	}

	_, err := prompt.Run()
	switch {
	case err == nil:
		return true, nil
	case errors.Is(err, promptui.ErrAbort):
		return false, nil
	case errors.Is(err, promptui.ErrEOF):
		return false, fmt.Errorf("cannot request user input: stdin is not an interactive terminal: %w", err)
	default:
		return false, err
	}
}
