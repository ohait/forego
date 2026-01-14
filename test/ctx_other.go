//go:build !linux && !darwin
// +build !linux,!darwin

package test

var isTerminal = func() bool {
	return true
}()
