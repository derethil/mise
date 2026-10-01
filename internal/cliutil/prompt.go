package cliutil

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"github.com/manifoldco/promptui"
)

func AutoConfirm(string) (bool, error) {
	return true, nil
}

func Confirm(question string) (bool, error) {
	return requestConfirmation(question, false)
}

func ConfirmWithDefault(question string, defaultYes bool) (bool, error) {
	return requestConfirmation(question, defaultYes)
}

func ConfirmOrDie(question string) {
	confirmed, err := requestConfirmation(question, false)
	if err == nil && confirmed {
		return
	}

	slog.Info("Aborted.", "question", question, "error", err)
	os.Exit(1)
}

func SelectOption(question string, options []string) (string, error) {
	return SelectOptionWithDefault(question, options, options[0])
}

func SelectOptionWithDefault(question string, options []string, defaultOption string) (string, error) {
	cursor := slices.Index(options, defaultOption)
	if cursor < 0 {
		cursor = 0
	}

	prompt := promptui.Select{
		Label:     question,
		Items:     options,
		CursorPos: cursor,
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

func requestConfirmation(question string, defaultYes bool) (bool, error) {
	def := ""
	if defaultYes {
		def = "y"
	}

	prompt := promptui.Prompt{
		Label:     question,
		IsConfirm: true,
		Default:   def,
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
