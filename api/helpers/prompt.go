package helpers

import (
	"github.com/manifoldco/promptui"
	"strconv"
	"strings"
)

func Prompt(message string) bool {
	p := promptui.Prompt{
		Label:     message,
		IsConfirm: true,
	}

	result, _ := p.Run()

	switch strings.TrimSpace(strings.ToLower(result)) {
	case "y", "yes":
		return true
	default:
		b, _ := strconv.ParseBool(result)
		return b
	}
}
