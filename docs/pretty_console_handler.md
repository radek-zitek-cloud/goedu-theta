# Pretty Console Handler for slog.Logger

This document describes the architecture, usage, and integration of the custom `PrettyConsoleHandler` for Go's `slog` logging package, as implemented in the `internal/logger` package.

---

## Overview

The `PrettyConsoleHandler` is a custom implementation of `slog.Handler` designed to provide human-friendly, colorized, and aligned log output for console use. It is intended for development and debugging scenarios where readability is prioritized over machine parsing.

- **Location:** `internal/logger/pretty_handler.go`
- **Integration:** Used via `logger.ConfigureLogger` or `logger.NewPrettyLogger`
- **Format:** Timestamp, colorized level, message, key-value pairs, and optional PC (program counter) for source

---

## Features

- Colorized log levels (debug, info, warn, error)
- Aligned and readable output
- Optional source (PC) display
- Compatible with Go's `slog` API
- No external dependencies

---

## Usage

### 1. Programmatic Usage

```
import (
    "log/slog"
    "github.com/radek-zitek-cloud/goedu-theta/internal/logger"
)

func main() {
    log := logger.NewPrettyLogger(slog.LevelDebug, false)
    log.Info("Server started", "port", 8080)
}
```

### 2. Via Configuration

Set your config (e.g., `config.json`):

```
{
  "logger": {
    "level": "debug",
    "format": "pretty",
    "add_source": false
  }
}
```

Then call:

```
import (
    "github.com/radek-zitek-cloud/goedu-theta/internal/config"
    "github.com/radek-zitek-cloud/goedu-theta/internal/logger"
)

func main() {
    cfg := config.Load() // or your config loading logic
    logger.ConfigureLogger(cfg.Logger)
    log := logger.GetLogger()
    log.Info("Pretty log enabled", "foo", "bar")
}
```

---

## Example Output

```
2024-06-01 12:34:56.789 INFO  Server started | port=8080 env=dev
2024-06-01 12:34:56.790 WARN  Disk space low | path=/var/data free=512MB
2024-06-01 12:34:56.791 ERROR Failed to connect | err=timeout
```

---

## Testing Suggestions

- Unit test that log output contains color codes for each level
- Test that key-value pairs are rendered and aligned
- Test with and without `add_source` enabled
- Test with different log levels and formats

---

## Troubleshooting

- **No color in output:** Ensure your terminal supports ANSI colors.
- **Not seeing pretty logs:** Set `format` to `pretty` in your config or use `NewPrettyLogger`.
- **PC value is not a file/line:** Go's `slog.Record.PC` is a program counter, not a file/line. For file/line, use `AddSource` with the default handler.

---

## Potential Improvements

- Add support for grouping and attribute inheritance
- Support for file/line source display (requires reflection or custom logic)
- Configurable color schemes
- Option to disable colors for non-TTY output

---

## References

- [Go slog package](https://pkg.go.dev/log/slog)
- [Go ANSI color codes](https://en.wikipedia.org/wiki/ANSI_escape_code)
