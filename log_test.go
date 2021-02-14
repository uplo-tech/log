package log

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
)

// tempDir joins the provided directories and prefixes them with the testing
// directory.
func tempDir(dirs ...string) string {
	path := filepath.Join(os.TempDir(), "LogTesting", filepath.Join(dirs...))
	err := os.RemoveAll(path) // remove old test data
	if err != nil {
		panic(err)
	}
	return path
}

// TestLogger checks that the basic functions of the file logger work as
// designed.
func TestLogger(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a folder for the log file.
	testdir := tempDir(t.Name())
	err := os.MkdirAll(testdir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Create the logger.
	logFilename := filepath.Join(testdir, "test.log")
	options := Options{
		Debug:   true,
		Release: Testing,
		Version: "0.0.1",
	}
	fl, err := NewFileLogger(logFilename, options)
	if err != nil {
		t.Fatal(err)
	}

	// Write an example statement, and then close the logger.
	fl.Println("TEST: this should get written to the logfile")
	err = fl.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Check that data was written to the log file. There should be three
	// lines, one for startup, the example line, and one to close the logger.
	expectedSubstring := []string{"STARTUP", "TEST", "SHUTDOWN", ""} // file ends with a newline
	fileData, err := ioutil.ReadFile(path.Clean(logFilename))
	if err != nil {
		t.Fatal(err)
	}
	fileLines := strings.Split(string(fileData), "\n")
	for i, line := range fileLines {
		if !strings.Contains(line, expectedSubstring[i]) {
			t.Error("did not find the expected message in the logger")
		}
	}
	if len(fileLines) != 4 { // file ends with a newline
		t.Error("logger did not create the correct number of lines:", len(fileLines))
	}
}

// TestLoggerCriticalPanic tests printing a critical message from the logger and
// verifies that a panic occurs.
func TestLoggerCriticalPanic(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a folder for the log file.
	testdir := tempDir(t.Name())
	err := os.MkdirAll(testdir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Create the logger.
	logFilename := filepath.Join(testdir, "test.log")
	options := Options{
		BinaryName: "test",
		Debug:      true,
		Release:    Testing, // Suppress printing the stacktrace.
		Version:    "0.0.1",
	}
	fl, err := NewFileLogger(logFilename, options)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := fl.Close(); err != nil {
			t.Error(err)
		}
	}()

	// Write a catch for a panic that should trigger when logger.Critical is
	// called.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("critical message was not thrown in a panic")
		}
		s := "Critical error: (test v0.0.1, Release: testing) a critical message"
		if !strings.Contains(r.(string), s) {
			t.Errorf("expected panic message %v, was %v", s, r)
		}
	}()
	fl.Critical("a critical message")
}

// TestLoggerCriticalNoPanic tests printing a critical message from the logger
// and verifies that no panic occurs.
func TestLoggerCriticalNoPanic(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a folder for the log file.
	testdir := tempDir(t.Name())
	err := os.MkdirAll(testdir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	// Create the logger.
	logFilename := filepath.Join(testdir, "test.log")
	options := Options{
		Debug:   false,
		Release: Testing,
		Version: "0.0.1",
	}
	fl, err := NewFileLogger(logFilename, options)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := fl.Close(); err != nil {
			t.Error(err)
		}
	}()

	// Write a catch that ensures no panic is triggered when logger.Critical is
	// called.
	defer func() {
		r := recover()
		if r != nil {
			t.Error("critical message was thrown in a panic")
		}
	}()
	fl.Critical("a critical message")
}

// TestLoggerDiscard tests that the DiscardLogger does induce panics.
func TestLoggerDiscard(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Write a catch that ensures a panic is triggered when logger.Critical is
	// called.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("critical message was not thrown in a panic")
		}
		s := "Critical error: (discard v0, Release: testing) a critical message"
		if !strings.Contains(r.(string), s) {
			t.Errorf("expected panic message %v, was %v", s, r)
		}
	}()
	DiscardLogger.Critical("a critical message")
}

// TestLoggerUninitializedReleaseType tests creating a logger with an
// uninitialized release type.
func TestLoggerUninitializedReleaseType(t *testing.T) {
	var buf bytes.Buffer

	// Write a catch for a panic that should trigger when initializing the
	// logger.
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected to panic")
		}
		s := "uninitialized release type"
		if !strings.Contains(r.(string), s) {
			t.Errorf("expected to get panic message %v, was %v", s, r)
		}
	}()

	// Initializing the logger without a release type should fail.
	options := Options{
		Debug:   false,
		Version: "0.0.1",
	}
	_, _ = NewLogger(&buf, options) // Don't check outputs, this should panic.
}

// TestLoggerWrites tests printing Errorf and Errorln to the log.
func TestLoggerWrites(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Create a folder for the log files.
	testdir := tempDir(t.Name())
	err := os.MkdirAll(testdir, 0700)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("TestErrorf", func(t *testing.T) {
		fl, path := newTestingFileLogger(t, testdir)
		fl.Errorf("%v - %v", "test", "message")
		testLogContainsMessage(t, path, "[ERROR] test - message")
	})

	t.Run("TestErrorln", func(t *testing.T) {
		fl, path := newTestingFileLogger(t, testdir)
		fl.Errorln("test", "message")
		testLogContainsMessage(t, path, "[ERROR] test message")
	})
}

// TestLoggerInitializedReleaseType tests creating a logger with the different
// release types.
func TestLoggerInitializedReleaseType(t *testing.T) {
	var buf bytes.Buffer

	options := Options{
		Debug:   false,
		Version: "0.0.1",
	}

	// Check initializing with a valid release type.
	for _, releaseType := range []ReleaseType{Release, Testing, Dev} {
		options.Release = releaseType
		_, err := NewLogger(&buf, options)
		if err != nil {
			t.Fatal(err)
		}
	}
}

// newTestingFileLogger returns a new file logger for a subtest and its
// filepath
func newTestingFileLogger(t *testing.T, testdir string) (*Logger, string) {
	subtestName := filepath.Base(t.Name())
	logFilepath := filepath.Join(testdir, subtestName+".log")

	options := Options{
		Release: Testing,
		Version: "0.0.1",
	}

	fl, err := NewFileLogger(logFilepath, options)
	if err != nil {
		t.Fatal(err)
	}
	return fl, logFilepath
}

// testLogContainsMessage tests that log file contains expected log message.
func testLogContainsMessage(t *testing.T, logFilepath, message string) {
	fileData, err := ioutil.ReadFile(path.Clean(logFilepath))
	if err != nil {
		t.Fatal(err)
	}
	fileLines := strings.Split(string(fileData), "\n")
	for _, line := range fileLines {
		if strings.HasSuffix(line, message) {
			return
		}
	}
	t.Error("did not find the expected message in the logger")
}
