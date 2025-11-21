package job

import (
	"io"
	"os"
	"path/filepath"

	"x-ui/logger"
	"x-ui/xray"
)

type ClearLogsJob struct{}

func NewClearLogsJob() *ClearLogsJob {
	return new(ClearLogsJob)
}

// ensureFileExists creates the necessary directories and file if they don't exist
func ensureFileExists(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	file.Close()
	return nil
}

// Here Run is an interface method of the Job interface
func (j *ClearLogsJob) Run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("ClearLogsJob panic recovered:", r)
		}
	}()

	logger.Info("ClearLogsJob: Starting log rotation and cleanup")
	
	logFiles := []string{xray.GetIPLimitLogPath(), xray.GetIPLimitBannedLogPath(), xray.GetAccessPersistentLogPath()}
	logFilesPrev := []string{xray.GetIPLimitBannedPrevLogPath(), xray.GetAccessPersistentPrevLogPath()}

	// Ensure all log files and their paths exist
	for _, path := range append(logFiles, logFilesPrev...) {
		if err := ensureFileExists(path); err != nil {
			logger.Error("ClearLogsJob: Failed to ensure log file exists:", path, "-", err)
		}
	}

	successCount := 0
	failCount := 0

	// Clear log files and copy to previous logs
	for i := 0; i < len(logFiles); i++ {
		if i > 0 {
			// Copy to previous logs
			logFilePrev, err := os.OpenFile(logFilesPrev[i-1], os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				logger.Error("ClearLogsJob: Failed to open previous log file for writing:", logFilesPrev[i-1], "-", err)
				failCount++
				continue
			}

			logFile, err := os.OpenFile(logFiles[i], os.O_RDONLY, 0644)
			if err != nil {
				logger.Error("ClearLogsJob: Failed to open current log file for reading:", logFiles[i], "-", err)
				logFilePrev.Close()
				failCount++
				continue
			}

			bytesWritten, err := io.Copy(logFilePrev, logFile)
			if err != nil {
				logger.Error("ClearLogsJob: Failed to copy log file:", logFiles[i], "to", logFilesPrev[i-1], "-", err)
				failCount++
			} else {
				logger.Debugf("ClearLogsJob: Copied %d bytes from %s to %s", bytesWritten, logFiles[i], logFilesPrev[i-1])
			}

			logFile.Close()
			logFilePrev.Close()
		}

		err := os.Truncate(logFiles[i], 0)
		if err != nil {
			logger.Error("ClearLogsJob: Failed to truncate log file:", logFiles[i], "-", err)
			failCount++
		} else {
			logger.Debug("ClearLogsJob: Truncated log file:", logFiles[i])
			successCount++
		}
	}

	logger.Infof("ClearLogsJob: Completed log rotation (success: %d, failed: %d)", successCount, failCount)
}
