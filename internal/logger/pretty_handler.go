package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// PrettyConsoleHandler is a slog.Handler that pretty-prints logs to the console with colors and alignment.
//
// This handler is intended for human-friendly development output. It colorizes log levels,
// aligns fields, and supports optional source location display. It is not suitable for machine parsing.
type PrettyConsoleHandler struct {
	out    io.Writer  // Output destination (e.g., os.Stdout)
	level  slog.Level // Minimum log level to output
	source bool       // Whether to print source location
}

// NewPrettyConsoleHandler creates a new PrettyConsoleHandler.
//
// Args:
//
//	w: Output writer (defaults to os.Stdout if nil)
//	opts: slog.HandlerOptions (level and AddSource supported)
//
// Returns:
//
//	*PrettyConsoleHandler instance
//
// Example:
//
//	handler := NewPrettyConsoleHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true})
func NewPrettyConsoleHandler(w io.Writer, opts *slog.HandlerOptions) *PrettyConsoleHandler {
	if w == nil {
		w = os.Stdout
	}
	level := slog.LevelInfo
	if opts != nil && opts.Level != nil {
		if l, ok := opts.Level.(slog.Level); ok {
			level = l
		} else {
			level = opts.Level.Level()
		}
	}
	source := false
	if opts != nil {
		source = opts.AddSource
	}
	return &PrettyConsoleHandler{out: w, level: level, source: source}
}

// Enabled implements slog.Handler.
// Returns true if the log level is enabled for output.
func (h *PrettyConsoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

// Handle implements slog.Handler.
// Formats and writes the log record to the output in a human-friendly way.
//
// Example output:
//
//	2024-06-01 12:34:56.789 INFO  Starting server | port=8080 env=dev
func (h *PrettyConsoleHandler) Handle(_ context.Context, r slog.Record) error {
	b := &strings.Builder{}
	// Timestamp
	timestamp := r.Time.Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(b, "%s ", timestamp)
	// Level (colorized)
	fmt.Fprintf(b, "%s%-5s%s ", levelColor(r.Level), r.Level.String(), resetColor())
	// Message
	fmt.Fprintf(b, "%s", r.Message)
	// Key-value pairs
	if r.NumAttrs() > 0 {
		b.WriteString(" | ")
		r.Attrs(func(a slog.Attr) bool {
			fmt.Fprintf(b, "%s=%v ", a.Key, a.Value)
			return true
		})
	}
	// Source (prints PC as hex if enabled, since slog.Frame is not public)
	if h.source && r.PC != 0 {
		fmt.Fprintf(b, " (pc=0x%x)", r.PC)
	}
	b.WriteString("\n")
	_, err := h.out.Write([]byte(b.String()))
	return err
}

// WithAttrs implements slog.Handler.
// Returns a handler with additional attributes (not supported in this implementation).
func (h *PrettyConsoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// TODO(2024-06-01): Implement attribute grouping if needed for advanced use cases.
	return h
}

// WithGroup implements slog.Handler.
// Returns a handler with a group name (not supported in this implementation).
func (h *PrettyConsoleHandler) WithGroup(name string) slog.Handler {
	// TODO(2024-06-01): Implement group support if needed for advanced use cases.
	return h
}

// levelColor returns the ANSI color code for a given slog.Level.
func levelColor(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return "\033[36m" // Cyan
	case slog.LevelInfo:
		return "\033[32m" // Green
	case slog.LevelWarn:
		return "\033[33m" // Yellow
	case slog.LevelError:
		return "\033[31m" // Red
	default:
		return ""
	}
}

// resetColor returns the ANSI reset code to clear color formatting.
func resetColor() string {
	return "\033[0m"
}
