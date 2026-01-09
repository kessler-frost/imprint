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
