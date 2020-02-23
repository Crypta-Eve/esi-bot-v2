package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

// Cors Middleware to Allow for Frontend Consumption of the API
func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "600")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

// NewStructuredLogger is a constructor for creating a request logger middleware
func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

// StructuredLogger holds our application's instance of our logger
type StructuredLogger struct {
	Logger *logrus.Logger
}

// NewLogEntry will return a new log entry scoped to the http.Request
func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{Logger: logrus.NewEntry(l.Logger)}
	logFields := logrus.Fields{}

	logFields["ts"] = time.Now().Format(time.RFC1123)

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		entry.Logger = entry.Logger.WithFields(logFields)
		return entry
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	var fields map[string]interface{}
	json.Unmarshal(body, &fields)

	logFields["data"] = fields

	entry.Logger = entry.Logger.WithFields(logFields)

	return entry
}

// StructuredLoggerEntry holds our FieldLogger entry
type StructuredLoggerEntry struct {
	Logger logrus.FieldLogger
}

// Write will write to logger entry once the http.Request is complete
func (l *StructuredLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elasped_ms": float64(elapsed.Nanoseconds()) / 1000000.0,
	})

	l.Logger.Infoln("request complete")
}

// Panic attaches the panic stack and text to the log entry
func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.Logger = l.Logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})

	l.Logger.Errorln("request panic'd")
}
