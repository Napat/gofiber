package testHelper

// Note: Show the basic concept of how to implement the assertion functions.
// Acctually, there are some of the ready test framework like,
// - https://github.com/stretchr/testify

import (
	"path/filepath"
	"runtime"
	"testing"
)

func AssertCorrectMessage(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func AssertCorrectHttpStatusCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("got %v want %v", got, want)
	}
}

func AssertError(t testing.TB, got error, want error) {
	t.Helper()
	if got == nil {
		t.Fatal("didn't get an error but wanted one")
	}

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func AssertJustError(t testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1) // similar with t.Helper()
		t.Fatalf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
	}
}

func AssertNoError(t testing.TB, got error) {
	t.Helper()
	if got != nil {
		t.Fatal("got an error but didn't want one")
	}
}
