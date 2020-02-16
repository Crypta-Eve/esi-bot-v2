package server

import (
	"fmt"
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

	// Force output to CST, the RV Standard
	location, _ := time.LoadLocation("America/Chicago")

	logFields["ts"] = time.Now().In(location).Format(time.RFC1123)

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	logFields["application"] = "agentx-pathfinder-api"
	logFields["http_scheme"] = scheme
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

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
