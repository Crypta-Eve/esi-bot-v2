package slack

import "errors"

var errCommandUndetermined = errors.New("unable to determine triggered command")
var errCommandWithInvalidArgs = errors.New("unable process invalid args")
var errCommandWithInvalidArgValue = errors.New("unable process invalid arg value")

var knownErrs = map[error]bool{
	errCommandUndetermined:        true,
	errCommandWithInvalidArgValue: true,
	errCommandWithInvalidArgs:     true,
}
