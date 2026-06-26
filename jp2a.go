package main

import (
	"fmt"
	"os/exec"
)

// emojiToASCII shells out to jp2a to convert the emoji PNG into ASCII art,
// matching the Python script's flags (colored when requested, fixed 80x38
// canvas, background-tuned ramp, optional horizontal mirror).
func emojiToASCII(fileName string, opts options) (string, error) {
	args := []string{
		"--background=" + opts.jp2aBackground,
		"--size=80x38",
	}
	if opts.color {
		args = append(args, "--colors")
	}
	if opts.mirror {
		args = append(args, "-x")
	}
	args = append(args, fileName)

	out, err := exec.Command("jp2a", args...).Output()
	if err != nil {
		return "", fmt.Errorf("jp2a failed: %w", err)
	}
	return string(out), nil
}
