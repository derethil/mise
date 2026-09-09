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

type PromptValidator func(input string) error

func PromptForInput(question, preset string, validators ...PromptValidator) (string, error) {
	return promptForInput(question, preset, false, validators...)
}

func PromptForHiddenInput(question, preset string, validators ...PromptValidator) (string, error) {
	return promptForInput(question, preset, true, validators...)
}

func promptForInput(question, preset string, hidden bool, validators ...PromptValidator) (string, error) {
	var mask rune
	if hidden {
		mask = '*'
	}

	prompt := promptui.Prompt{
		Default:   preset,
		Label:     question,
		Mask:      mask,
		AllowEdit: true,
		Validate: func(input string) (err error) {
			if len(validators) > 0 {
				for _, validator := range validators {
					err = validator(input)
				}
			}

			return err
		},
	}

	userInput, err := prompt.Run()
	err = handlePromptError(err)

	return userInput, err
}

func requestConfirmation(question string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     question,
		IsConfirm: true,
	}

	_, err := prompt.Run()
	err = handlePromptError(err)

	return err == nil, err
}

func IsUserAbort(err error) bool {
	return errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF) || errors.Is(err, promptui.ErrAbort)
}

func handlePromptError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, promptui.ErrAbort):
		return err
	case errors.Is(err, promptui.ErrEOF):
		return fmt.Errorf("cannot request user input: stdin is not an interactive terminal: %w", err)
	default:
		return err
	}
}
