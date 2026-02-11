package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var (
	logger  *log.Logger
	once    sync.Once
	logFile *os.File
)

func SendGetRequest(url string) (map[string]interface{}, error) {
	fmt.Printf("[INFO] Sending GET request: %s\n", url)

	// --- 1️⃣ Send request ---
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("[ERROR] Request failed for URL=%s, err=%v\n", url, err)
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	defer resp.Body.Close()
	fmt.Printf("[INFO] Response status for URL=%s: %s\n", url, resp.Status)

	// --- 2️⃣ Read body ---
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[ERROR] Failed to read response body for URL=%s, err=%v\n", url, err)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	fmt.Printf("[DEBUG] Response body length=%d bytes\n", len(body))

	// --- 3️⃣ Decode JSON ---
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		fmt.Printf("[ERROR] JSON unmarshal failed for URL=%s, err=%v\nResponse body: %s\n", url, err, string(body))
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	fmt.Printf("[INFO] Successfully decoded JSON from URL=%s\n", url)
	return result, nil
}

func initLogger() {
	var err error
	logFile, err = os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	logger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
}

// Log writes formatted logs to both console and log file
func Log(format string, v ...interface{}) {
	once.Do(initLogger)

	log.Printf(format, v...)    // Print to console (stderr)
	logger.Printf(format, v...) // Print to file with file/line info
}

// CloseLog should be called on program exit to close the log file
func CloseLog() {
	if logFile != nil {
		logFile.Close()
	}
}

var (
	klogger    *log.Logger
	loggerOnce sync.Once
)

// ErrorLogger returns a singleton error logger
func ErrorLogger() *log.Logger {
	loggerOnce.Do(func() {
		file, err := os.OpenFile("error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Failed to open error.log:", err)
		}
		logger = log.New(file, "ERROR: ", log.LstdFlags)
	})
	return logger
}
func Convert[T any](s string) (T, error) {
	var zero T

	switch any(zero).(type) {
	case int:
		v, err := strconv.Atoi(s)
		return any(v).(T), err
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return any(v).(T), err
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		return any(v).(T), err
	case bool:
		v, err := strconv.ParseBool(s)
		return any(v).(T), err
	case string:
		return any(s).(T), nil
	default:
		return zero, fmt.Errorf("unsupported type")
	}
}
