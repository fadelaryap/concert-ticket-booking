package utils

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

func init() {
	// Initialize loggers to stdout and stderr
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// LogInfo logs informational messages.
func LogInfo(format string, v ...interface{}) {
	infoLogger.Printf(format+"\n", v...)
}

// LogWarning logs warning messages.
func LogWarning(format string, v ...interface{}) {
	warnLogger.Printf(format+"\n", v...)
}

// LogError logs error messages.
func LogError(format string, v ...interface{}) {
	errorLogger.Printf(format+"\n", v...)
}

// FormatValidationErrors converts validator.ValidationErrors into a readable string.
func FormatValidationErrors(errs validator.ValidationErrors) string {
	var sb strings.Builder
	for i, e := range errs {
		sb.WriteString(fmt.Sprintf("Field '%s' failed on '%s' validation (value '%v')",
			e.Field(), e.Tag(), e.Value()))
		if i < len(errs)-1 {
			sb.WriteString("; ")
		}
	}
	return sb.String()
}
