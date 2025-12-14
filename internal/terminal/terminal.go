package terminal

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Terminal manages a real terminal session via ttyd and headless Chrome.
type Terminal struct {
	mu      sync.RWMutex
	browser *rod.Browser
	page    *rod.Page
	cmd     *exec.Cmd
	port    int
	shell   string
	rows    int
	cols    int
}

// keyMap maps key names to go-rod input.Key constants
var keyMap = map[string]input.Key{
	"enter":     input.Enter,
	"backspace": input.Backspace,
	"tab":       input.Tab,
	"escape":    input.Escape,
	"up":        input.ArrowUp,
	"down":      input.ArrowDown,
	"left":      input.ArrowLeft,
	"right":     input.ArrowRight,
	"space":     input.Space,
	"delete":    input.Delete,
	"insert":    input.Insert,
	"home":      input.Home,
	"end":       input.End,
	"pageup":    input.PageUp,
	"pagedown":  input.PageDown,
	"f1":        input.F1,
	"f2":        input.F2,
	"f3":        input.F3,
	"f4":        input.F4,
	"f5":        input.F5,
	"f6":        input.F6,
	"f7":        input.F7,
	"f8":        input.F8,
	"f9":        input.F9,
	"f10":       input.F10,
	"f11":       input.F11,
	"f12":       input.F12,
}

// letterKeyMap maps single letters to their input.Key constants
var letterKeyMap = map[rune]input.Key{
	'a': input.KeyA, 'b': input.KeyB, 'c': input.KeyC, 'd': input.KeyD,
	'e': input.KeyE, 'f': input.KeyF, 'g': input.KeyG, 'h': input.KeyH,
	'i': input.KeyI, 'j': input.KeyJ, 'k': input.KeyK, 'l': input.KeyL,
	'm': input.KeyM, 'n': input.KeyN, 'o': input.KeyO, 'p': input.KeyP,
	'q': input.KeyQ, 'r': input.KeyR, 's': input.KeyS, 't': input.KeyT,
	'u': input.KeyU, 'v': input.KeyV, 'w': input.KeyW, 'x': input.KeyX,
	'y': input.KeyY, 'z': input.KeyZ,
}

// New creates a new Terminal instance.
func New(shell string, rows, cols int) (*Terminal, error) {
	port, err := findFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to find free port: %w", err)
	}

	return &Terminal{
		port:  port,
		shell: shell,
		rows:  rows,
		cols:  cols,
	}, nil
}

// findFreePort returns an available TCP port.
func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// Start launches the terminal session.
func (t *Terminal) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.startUnlocked()
}

// startUnlocked launches the terminal session without acquiring the lock.
// Caller must hold the lock.
func (t *Terminal) startUnlocked() error {
	// Start ttyd process with login interactive shell
	t.cmd = exec.Command("ttyd",
		"--port", fmt.Sprintf("%d", t.port),
		"--interface", "127.0.0.1",
		"--writable",
		t.shell, "-l", "-i",
	)

	err := t.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ttyd: %w", err)
	}

	// Wait for ttyd to be ready
	time.Sleep(500 * time.Millisecond)

	// Launch headless browser
	url := launcher.New().Headless(true).MustLaunch()
	t.browser = rod.New().ControlURL(url).MustConnect()

	// Navigate to ttyd
	t.page = t.browser.MustPage(fmt.Sprintf("http://127.0.0.1:%d", t.port))

	// Wait for terminal to initialize
	t.page.MustWaitStable()

	return nil
}

// SendKey sends a keystroke to the terminal.
func (t *Terminal) SendKey(key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.page == nil {
		return fmt.Errorf("terminal not ready")
	}

	return t.sendKeyUnlocked(key)
}

// sendKeyUnlocked sends a keystroke without acquiring the lock.
// Caller must hold the lock.
func (t *Terminal) sendKeyUnlocked(key string) error {
	key = strings.ToLower(key)

	// Handle modifier combinations like "ctrl+c"
	parts := strings.Split(key, "+")
	if len(parts) == 2 {
		modifier := parts[0]
		mainKey := parts[1]

		switch modifier {
		case "ctrl":
			return t.sendCtrlKey(mainKey)
		case "alt":
			return t.sendAltKey(mainKey)
		case "shift":
			return t.sendShiftKey(mainKey)
		default:
			return fmt.Errorf("unknown modifier: %s", modifier)
		}
	}

	// Single key press - check special keys first
	if k, ok := keyMap[key]; ok {
		return t.page.Keyboard.Press(k)
	}

	// Single letter key
	if len(key) == 1 {
		char := rune(key[0])
		if k, ok := letterKeyMap[char]; ok {
			return t.page.Keyboard.Press(k)
		}
	}

	return fmt.Errorf("unknown key: %s", key)
}

// SendKeys sends multiple keystrokes to the terminal.
// Fails fast on the first error, returning which key failed.
func (t *Terminal) SendKeys(keys []string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.page == nil {
		return fmt.Errorf("terminal not ready")
	}

	for i, key := range keys {
		if err := t.sendKeyUnlocked(key); err != nil {
			return fmt.Errorf("key %d (%s): %w", i, key, err)
		}
	}
	return nil
}

// sendCtrlKey sends a Ctrl+key combination
func (t *Terminal) sendCtrlKey(key string) error {
	var targetKey input.Key

	if k, ok := keyMap[key]; ok {
		targetKey = k
	} else if len(key) == 1 {
		char := rune(key[0])
		if k, ok := letterKeyMap[char]; ok {
			targetKey = k
		} else {
			return fmt.Errorf("unknown key: %s", key)
		}
	} else {
		return fmt.Errorf("unknown key: %s", key)
	}

	// Use KeyActions for modifier combinations
	return t.page.KeyActions().Press(input.ControlLeft).Type(targetKey).Do()
}

// sendAltKey sends an Alt+key combination
func (t *Terminal) sendAltKey(key string) error {
	var targetKey input.Key

	if k, ok := keyMap[key]; ok {
		targetKey = k
	} else if len(key) == 1 {
		char := rune(key[0])
		if k, ok := letterKeyMap[char]; ok {
			targetKey = k
		} else {
			return fmt.Errorf("unknown key: %s", key)
		}
	} else {
		return fmt.Errorf("unknown key: %s", key)
	}

	// Use KeyActions for modifier combinations
	return t.page.KeyActions().Press(input.AltLeft).Type(targetKey).Do()
}

// sendShiftKey sends a Shift+key combination
func (t *Terminal) sendShiftKey(key string) error {
	var targetKey input.Key

	if k, ok := keyMap[key]; ok {
		targetKey = k
	} else if len(key) == 1 {
		char := rune(key[0])
		if k, ok := letterKeyMap[char]; ok {
			targetKey = k
		} else {
			return fmt.Errorf("unknown key: %s", key)
		}
	} else {
		return fmt.Errorf("unknown key: %s", key)
	}

	// Use KeyActions for modifier combinations
	return t.page.KeyActions().Press(input.ShiftLeft).Type(targetKey).Do()
}

// Type types a string of characters.
func (t *Terminal) Type(text string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.page == nil {
		return fmt.Errorf("terminal not ready")
	}

	// Find the terminal's textarea element and input text
	textarea := t.page.MustElement("textarea")
	return textarea.Input(text)
}

// Screenshot captures the current screen as JPEG with the specified quality (0-100).
func (t *Terminal) Screenshot(quality int) ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.page == nil {
		return nil, fmt.Errorf("terminal not ready")
	}

	return t.page.Screenshot(false, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: &quality,
	})
}

// GetText returns the current screen as plain text.
func (t *Terminal) GetText() (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.page == nil {
		return "", fmt.Errorf("terminal not ready")
	}

	// Use xterm.js buffer API to get terminal content
	result, err := t.page.Eval(`() => {
		const term = window.term;
		if (!term) return "";

		const buffer = term.buffer.active;
		const lines = [];
		for (let i = 0; i < buffer.length; i++) {
			lines.push(buffer.getLine(i).translateToString().trimEnd());
		}
		return lines.join("\n");
	}`)

	if err != nil {
		return "", fmt.Errorf("failed to get terminal text: %w", err)
	}

	return result.Value.String(), nil
}

// WaitForText polls until the specified text appears on screen or timeout.
// Returns elapsed time in ms, whether text was found, and any error.
func (t *Terminal) WaitForText(text string, timeoutMs int) (elapsedMs int, found bool, err error) {
	pollInterval := 100 * time.Millisecond
	startTime := time.Now()
	timeoutDuration := time.Duration(timeoutMs) * time.Millisecond

	for {
		screenText, err := t.GetText()
		if err != nil {
			return int(time.Since(startTime).Milliseconds()), false, fmt.Errorf("failed to get screen text: %w", err)
		}

		if strings.Contains(screenText, text) {
			return int(time.Since(startTime).Milliseconds()), true, nil
		}

		elapsed := time.Since(startTime)
		if elapsed >= timeoutDuration {
			return int(elapsed.Milliseconds()), false, nil
		}

		time.Sleep(pollInterval)
	}
}

// WaitForStable polls until the screen stops changing for stableMs duration or timeout.
// Returns elapsed time in ms, whether stability was achieved, and any error.
func (t *Terminal) WaitForStable(timeoutMs, stableMs int) (elapsedMs int, stable bool, err error) {
	pollInterval := 100 * time.Millisecond
	timeout := time.Duration(timeoutMs) * time.Millisecond
	stableDuration := time.Duration(stableMs) * time.Millisecond

	startTime := time.Now()
	var lastText string
	var lastChangeTime time.Time

	lastText, err = t.GetText()
	if err != nil {
		return 0, false, err
	}
	lastChangeTime = startTime

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			currentText, err := t.GetText()
			if err != nil {
				elapsed := int(time.Since(startTime).Milliseconds())
				return elapsed, false, err
			}

			if currentText != lastText {
				lastText = currentText
				lastChangeTime = time.Now()
			}

			timeSinceChange := time.Since(lastChangeTime)
			if timeSinceChange >= stableDuration {
				elapsed := int(time.Since(startTime).Milliseconds())
				return elapsed, true, nil
			}

			if time.Since(startTime) >= timeout {
				elapsed := int(time.Since(startTime).Milliseconds())
				return elapsed, false, nil
			}
		}
	}
}

// Resize changes the terminal dimensions.
func (t *Terminal) Resize(rows, cols int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.page == nil {
		return fmt.Errorf("terminal not ready")
	}

	// Use xterm.js resize API
	_, err := t.page.Eval(fmt.Sprintf(`() => {
		const term = window.term;
		if (term) {
			term.resize(%d, %d);
		}
	}`, cols, rows))

	if err != nil {
		return fmt.Errorf("failed to resize terminal: %w", err)
	}

	t.rows = rows
	t.cols = cols

	return nil
}

// Close terminates the terminal session.
func (t *Terminal) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Close browser
	if t.browser != nil {
		t.browser.MustClose()
	}

	// Kill ttyd process
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
	}

	return nil
}

// Restart closes and restarts the terminal, optionally with a new command.
// If command is empty, uses the existing shell command.
func (t *Terminal) Restart(command string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Close existing browser and ttyd
	if t.browser != nil {
		t.browser.MustClose()
		t.browser = nil
	}
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Kill()
		t.cmd.Wait()
		t.cmd = nil
	}
	t.page = nil

	// Update command if provided
	if command != "" {
		t.shell = command
	}

	// Find new port (old one may still be releasing)
	port, err := findFreePort()
	if err != nil {
		return fmt.Errorf("failed to find free port: %w", err)
	}
	t.port = port

	return t.startUnlocked()
}

// Status returns terminal status information.
func (t *Terminal) Status() (rows, cols int, ready bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	ready = t.page != nil
	return t.rows, t.cols, ready
}
