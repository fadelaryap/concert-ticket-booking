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

	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func LogInfo(format string, v ...interface{}) {
	infoLogger.Printf(format+"\n", v...)
}

func LogWarning(format string, v ...interface{}) {
	warnLogger.Printf(format+"\n", v...)
}

func LogError(format string, v ...interface{}) {
	errorLogger.Printf(format+"\n", v...)
}

func LogFatal(format string, v ...interface{}) {
	errorLogger.Fatalf(format+"\n", v...)
}

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
