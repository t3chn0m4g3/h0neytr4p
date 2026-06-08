package h0neytr4p

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ensureRuntimeDir(path string) error {
	if err := os.MkdirAll(path, RuntimeDirMode); err != nil {
		return err
	}
	return os.Chmod(path, RuntimeDirMode)
}

// Initialize log file
func InitLogFile(filename string, verbose bool) error {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if dir := filepath.Dir(filename); dir != "." && dir != "" {
		if err := ensureRuntimeDir(dir); err != nil {
			return fmt.Errorf("create log directory: %w", err)
		}
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, RuntimeFileMode)
	Verbose = verbose
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	if err := file.Chmod(RuntimeFileMode); err != nil {
		_ = file.Close()
		return fmt.Errorf("set log file mode: %w", err)
	}
	if logFile != nil {
		_ = logFile.Close()
	}
	logFile = file
	fmt.Println("Logging is configured and ready.")
	return nil
}

// Initialize payload folder
func InitPayloadFolder(folder string) error {
	payloadFolder = folder
	if err := ensureRuntimeDir(folder); err != nil {
		return fmt.Errorf("create payload folder: %w", err)
	}
	fmt.Println("Payload folder is configured and ready.")
	return nil
}

func CloseLogFile() error {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if logFile == nil {
		return nil
	}
	err := logFile.Close()
	logFile = nil
	return err
}

func LogEntry(details map[string]string) error {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if logFile == nil {
		return fmt.Errorf("log file is not configured")
	}

	entryJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}
	if _, err := fmt.Fprintln(logFile, string(entryJSON)); err != nil {
		return fmt.Errorf("write log entry: %w", err)
	}

	if Verbose {
		fmt.Printf("[%s] [Path: %s] [Trapped: %s]\n", details["timestamp"], details["request_uri"], details["trapped"])
	}

	if err := logFile.Sync(); err != nil {
		return fmt.Errorf("sync log entry: %w", err)
	}
	return nil
}
