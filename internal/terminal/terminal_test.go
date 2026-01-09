package terminal

import (
	"os"
	"strings"
	"testing"
)

var testTerminal *Terminal

func TestMain(m *testing.M) {
	var err error
	testTerminal, err = New("/bin/sh", 24, 80)
	if err != nil {
		os.Exit(1)
	}

	err = testTerminal.Start()
	if err != nil {
		os.Exit(1)
	}

	testTerminal.WaitForStable(2000, 100)

	code := m.Run()

	testTerminal.Close()
	os.Exit(code)
}

// resetTerminal clears the terminal state between tests.
// ctrl+c cancels any running command; clear removes previous output.
func resetTerminal(t *testing.T) {
	t.Helper()
	if err := testTerminal.SendKey("ctrl+c"); err != nil {
		t.Fatalf("resetTerminal: SendKey(ctrl+c) failed: %v", err)
	}
	if err := testTerminal.Type("clear"); err != nil {
		t.Fatalf("resetTerminal: Type(clear) failed: %v", err)
	}
	if err := testTerminal.SendKey("enter"); err != nil {
		t.Fatalf("resetTerminal: SendKey(enter) failed: %v", err)
	}
	testTerminal.WaitForStable(500, 100)
}

func TestTerminal(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"TypeAfterCommandExecution", testTypeAfterCommandExecution},
		{"TypeUnicode", testTypeUnicode},
		{"SendKeyAfterSendKey", testSendKeyAfterSendKey},
		{"SendKeyCharacters", testSendKeyCharacters},
		{"SendKeyAliases", testSendKeyAliases},
		{"SendKeyErrors", testSendKeyErrors},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resetTerminal(t)
			tc.fn(t)
		})
	}
}

// testTypeAfterCommandExecution verifies Type() works after SendKey().
// Type() must succeed regardless of prior SendKey() calls, since agents
// commonly alternate between typing text and pressing control keys.
func testTypeAfterCommandExecution(t *testing.T) {
	if err := testTerminal.Type("echo first"); err != nil {
		t.Fatalf("Type(echo first) failed: %v", err)
	}
	if err := testTerminal.SendKey("enter"); err != nil {
		t.Fatalf("SendKey(enter) failed: %v", err)
	}
	testTerminal.WaitForStable(1000, 100)

	if err := testTerminal.Type("echo second"); err != nil {
		t.Fatalf("Type(echo second) failed: %v", err)
	}
	if err := testTerminal.SendKey("enter"); err != nil {
		t.Fatalf("SendKey(enter) failed: %v", err)
	}
	testTerminal.WaitForStable(1000, 100)

	screen, err := testTerminal.GetText()
	if err != nil {
		t.Fatalf("GetText() failed: %v", err)
	}
	if !strings.Contains(screen, "second") {
		t.Errorf("Second command output not found. Screen:\n%s", screen)
	}
}

// testTypeUnicode verifies Type() handles Unicode correctly.
// The term.input() API must preserve multi-byte characters.
func testTypeUnicode(t *testing.T) {
	if err := testTerminal.Type("echo '哎呀屌你好打死你'"); err != nil {
		t.Fatalf("Type(unicode) failed: %v", err)
	}
	if err := testTerminal.SendKey("enter"); err != nil {
		t.Fatalf("SendKey(enter) failed: %v", err)
	}
	testTerminal.WaitForStable(1000, 100)

	screen, err := testTerminal.GetText()
	if err != nil {
		t.Fatalf("GetText() failed: %v", err)
	}
	if !strings.Contains(screen, "哎呀屌你好打死你") {
		t.Errorf("Unicode output not found. Screen:\n%s", screen)
	}
}

// testSendKeyAfterSendKey verifies consecutive SendKey() calls work.
// Each SendKey() must leave xterm.js in a clean state for the next input.
func testSendKeyAfterSendKey(t *testing.T) {
	// Type a command, press enter, then use arrow-up to recall it
	if err := testTerminal.Type("echo sendkey_test"); err != nil {
		t.Fatalf("Type() failed: %v", err)
	}
	if err := testTerminal.SendKey("enter"); err != nil {
		t.Fatalf("SendKey(enter) failed: %v", err)
	}
	testTerminal.WaitForStable(1000, 100)

	// Arrow up should recall the previous command
	if err := testTerminal.SendKey("up"); err != nil {
		t.Fatalf("SendKey(up) failed: %v", err)
	}
	testTerminal.WaitForStable(500, 100)

	// Execute the recalled command
	if err := testTerminal.SendKey("enter"); err != nil {
		t.Fatalf("SendKey(enter) after up failed: %v", err)
	}
	testTerminal.WaitForStable(1000, 100)

	screen, err := testTerminal.GetText()
	if err != nil {
		t.Fatalf("GetText() failed: %v", err)
	}

	// Should see "sendkey_test" output twice (original + recalled)
	count := strings.Count(screen, "sendkey_test")
	if count < 2 {
		t.Errorf("Expected 'sendkey_test' at least twice, found %d times. Screen:\n%s", count, screen)
	}
}

// testSendKeyCharacters verifies SendKey() accepts various character types:
// special keys, letters, digits, punctuation, and Unicode (including grapheme clusters).
func testSendKeyCharacters(t *testing.T) {
	samples := []struct {
		key     string
		wantErr bool
	}{
		// Special keys (regression)
		{"enter", false},
		{"tab", false},
		{"escape", false},
		{"up", false},
		{"down", false},
		{"backspace", false},
		{"space", false},
		// Letters (regression)
		{"a", false},
		{"z", false},
		{"A", false},
		// Digits (new)
		{"0", false},
		{"5", false},
		{"9", false},
		// Punctuation (new)
		{"/", false},
		{".", false},
		{",", false},
		{"+", false},
		{"[", false},
		{"]", false},
		{"-", false},
		// Unicode (new)
		{"中", false},
		{"é", false},
		// Literal space (should work like "space")
		{" ", false},
	}

	for _, s := range samples {
		err := testTerminal.SendKey(s.key)
		if (err != nil) != s.wantErr {
			t.Errorf("SendKey(%q): got error %v, wantErr %v", s.key, err, s.wantErr)
		}
	}

	flagUS := "\U0001F1FA\U0001F1F8"
	outputSamples := []string{"A", "5", ".", "+", "中", "é", flagUS}
	for _, key := range outputSamples {
		resetTerminal(t)
		if err := testTerminal.Type("printf '%s\\n' "); err != nil {
			t.Fatalf("Type(printf) failed: %v", err)
		}
		if err := testTerminal.SendKey(key); err != nil {
			t.Fatalf("SendKey(%q) failed: %v", key, err)
		}
		if err := testTerminal.SendKey("enter"); err != nil {
			t.Fatalf("SendKey(enter) failed: %v", err)
		}
		testTerminal.WaitForStable(1000, 100)
		assertOutputLine(t, key)
	}
}

// testSendKeyAliases verifies literal control characters behave like their named keys.
func testSendKeyAliases(t *testing.T) {
	if err := testTerminal.SendKey("\t"); err != nil {
		t.Fatalf("SendKey(\\t) failed: %v", err)
	}

	if err := testTerminal.Type("printf '%s\\n' '"); err != nil {
		t.Fatalf("Type(printf) failed: %v", err)
	}
	if err := testTerminal.SendKey("A"); err != nil {
		t.Fatalf("SendKey(A) failed: %v", err)
	}
	if err := testTerminal.SendKey(" "); err != nil {
		t.Fatalf("SendKey(space) failed: %v", err)
	}
	if err := testTerminal.SendKey("B"); err != nil {
		t.Fatalf("SendKey(B) failed: %v", err)
	}
	if err := testTerminal.SendKey("'"); err != nil {
		t.Fatalf("SendKey(') failed: %v", err)
	}
	if err := testTerminal.SendKey("\n"); err != nil {
		t.Fatalf("SendKey(\\n) failed: %v", err)
	}
	testTerminal.WaitForStable(1000, 100)
	assertOutputLine(t, "A B")
}

// testSendKeyErrors verifies SendKey() returns errors for invalid input.
func testSendKeyErrors(t *testing.T) {
	errorCases := []string{
		"",       // empty string
		"foobar", // unknown key name
		"ctrl+",  // incomplete modifier
		"++a",    // malformed
		"\x00",   // non-printable
	}

	for _, key := range errorCases {
		err := testTerminal.SendKey(key)
		if err == nil {
			t.Errorf("SendKey(%q): expected error, got nil", key)
		}
	}
}

func assertOutputLine(t *testing.T, expected string) {
	t.Helper()

	screen, err := testTerminal.GetText()
	if err != nil {
		t.Fatalf("GetText() failed: %v", err)
	}

	for _, line := range strings.Split(screen, "\n") {
		if line == expected {
			return
		}
	}

	t.Errorf("Expected output line %q not found. Screen:\n%s", expected, screen)
}
