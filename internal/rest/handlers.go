package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
)

// KeystrokesRequest represents the JSON body for /keystrokes endpoint.
type KeystrokesRequest struct {
	Keys []string `json:"keys"`
}

// TypeRequest represents the JSON body for /type endpoint.
type TypeRequest struct {
	Text string `json:"text"`
}

// ResizeRequest represents the JSON body for /resize endpoint.
type ResizeRequest struct {
	Rows int `json:"rows"`
	Cols int `json:"cols"`
}

// RestartRequest represents the JSON body for /restart endpoint.
type RestartRequest struct {
	Command string `json:"command,omitempty"`
}

// WaitForTextRequest represents the JSON body for /wait/text endpoint.
type WaitForTextRequest struct {
	Text      string `json:"text"`
	TimeoutMs int    `json:"timeout_ms,omitempty"`
}

// WaitForTextResponse represents the response from /wait/text endpoint.
type WaitForTextResponse struct {
	Found     bool `json:"found"`
	ElapsedMs int  `json:"elapsed_ms"`
}

// WaitStableRequest represents the JSON body for /wait/stable endpoint.
type WaitStableRequest struct {
	TimeoutMs int `json:"timeout_ms,omitempty"`
	StableMs  int `json:"stable_ms,omitempty"`
}

// WaitStableResponse represents the response from /wait/stable endpoint.
type WaitStableResponse struct {
	Stable    bool `json:"stable"`
	ElapsedMs int  `json:"elapsed_ms"`
}

// SuccessResponse represents a successful operation response.
type SuccessResponse struct {
	Success bool `json:"success"`
}

// StatusResponse represents the terminal status response.
type StatusResponse struct {
	Rows  int  `json:"rows"`
	Cols  int  `json:"cols"`
	Ready bool `json:"ready"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// handleKeystrokes handles POST /keystrokes - sends multiple keys to the terminal.
func (s *Server) handleKeystrokes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req KeystrokesRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid JSON"})
		return
	}

	if len(req.Keys) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "keys field is required and must not be empty"})
		return
	}

	if err := s.term.SendKeys(req.Keys); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Success: true})
}

// handleType handles POST /type - types a string of characters.
func (s *Server) handleType(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req TypeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid JSON"})
		return
	}

	if req.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "text field is required"})
		return
	}

	if err := s.term.Type(req.Text); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Success: true})
}

// handleScreen handles GET /screen - returns current screen as JPEG.
func (s *Server) handleScreen(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	quality := 70 // default
	if q := r.URL.Query().Get("quality"); q != "" {
		if parsed, err := strconv.Atoi(q); err == nil && parsed >= 0 && parsed <= 100 {
			quality = parsed
		}
	}

	screenshot, err := s.term.Screenshot(quality)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(screenshot)
}

// handleScreenText handles GET /screen/text - returns current screen as text.
func (s *Server) handleScreenText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	text, err := s.term.GetText()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(text))
}

// handleStatus handles GET /status - returns terminal status information.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	rows, cols, ready := s.term.Status()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(StatusResponse{
		Rows:  rows,
		Cols:  cols,
		Ready: ready,
	})
}

// handleResize handles POST /resize - resizes the terminal.
func (s *Server) handleResize(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req ResizeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid JSON"})
		return
	}

	if req.Rows <= 0 || req.Cols <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "rows and cols must be positive"})
		return
	}

	if err := s.term.Resize(req.Rows, req.Cols); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Success: true})
}

// handleRestart handles POST /restart - restarts the terminal.
func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	var req RestartRequest
	json.NewDecoder(r.Body).Decode(&req)
	defer r.Body.Close()

	if err := s.term.Restart(req.Command); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Success: true})
}

// handleWaitForText handles POST /wait/text - waits for text to appear on screen.
func (s *Server) handleWaitForText(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req WaitForTextRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid JSON"})
		return
	}

	if req.Text == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "text field is required"})
		return
	}

	timeoutMs := req.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}

	elapsedMs, found, err := s.term.WaitForText(req.Text, timeoutMs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(WaitForTextResponse{
		Found:     found,
		ElapsedMs: elapsedMs,
	})
}

// handleWaitForStable handles POST /wait/stable - waits for screen to become stable.
func (s *Server) handleWaitForStable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "failed to read request body"})
		return
	}
	defer r.Body.Close()

	var req WaitStableRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid JSON"})
		return
	}

	timeoutMs := req.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}

	stableMs := req.StableMs
	if stableMs <= 0 {
		stableMs = 500
	}

	elapsedMs, stable, err := s.term.WaitForStable(timeoutMs, stableMs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(WaitStableResponse{
		Stable:    stable,
		ElapsedMs: elapsedMs,
	})
}
