```
██╗░░░░░░█████╗░░██████╗░░██████╗░███████╗██████╗░
██║░░░░░██╔══██╗██╔════╝░██╔════╝░██╔════╝██╔══██╗
██║░░░░░██║░░██║██║░░██╗░██║░░██╗░█████╗░░██████╔╝
██║░░░░░██║░░██║██║░░╚██╗██║░░╚██╗██╔══╝░░██╔══██╗
███████╗╚█████╔╝╚██████╔╝╚██████╔╝███████╗██║░░██║
╚══════╝░╚════╝░░╚═════╝░░╚═════╝░╚══════╝╚═╝░░╚═╝
```

# Logger: Context-Aware Logging for Better Insights

`logger` is an attempt to improve log visualization by embedding parent/child relationships within the logging context, making it easier to trace log flows. It is uses `log/slog` package under the hood.

## Getting Started

### 1. Install the Logger Binary

To install the logger binary, run the following command:

```bash
go install ella.to/logger/cmd/logger-server@latest
```

This binary includes both the UI and server components. You can start the logger server by simply running:

```bash
logger-server
```

By default, the server runs on address `localhost:2022`. To change the port, specify it as an argument:

```bash
logger-server localhost:2021
```

You can then access the UI via your browser at:  
`http://localhost:2022`

### 2. Integrating Logger into Your Project

First, include the logger library in your project by running:

```bash
go get ella.to/logger@latest
```

Next, add the following code to the `main` function of your project to set up the logger:

```go
package main

import (
    // other imports
    "log/slog"
    "os"

    "ella.to/logger"
)

func main() {
    slog.SetDefault(
        slog.New(
            logger.NewHttpExporter(
                "http://localhost:2022", // logger server address
                slog.NewJSONHandler(os.Stdout, nil),
            ),
        ),
    )

    slog.SetLoggerLevel(slog.LevelDebug)

    // your application logic
}
```

### 3. Logging Usage

The logger provides four logging functions—`Info`, `Debug`, `Warn`, and `Error`—which work similarly to `slog.InfoContext`.

Example:

```go
ctx := context.Background()

logger.Info(ctx, "first log message", "request_id", 1)
```

Each of these functions returns a new context, allowing you to pass it to subsequent logs, creating parent/child relationships between logs.

### 4. HTTP Middleware

This library also includes an HTTP middleware compatible with Go's `http.Handler` signature:

```go
func(http.Handler) http.Handler
```

You can use this middleware to either retrieve the parent log ID from the incoming request headers or generate a new parent ID, ensuring all logs within an HTTP request are grouped together for better traceability.
