package sync

import (
	"errors"
	"fmt"
	"time"

	"github.com/rigter/devitri/backend/internal/config"
)

// ErrBulkDeleteBlocked is returned when a delete would exceed configured safety thresholds.
var ErrBulkDeleteBlocked = errors.New("bulk delete blocked")

// BulkDeleteCheckResult represents the result of a bulk delete check
type BulkDeleteCheckResult struct {
	IsSafe     bool
	Reason     string
	FileCount  int
	Threshold  int
	Percentage float64
}

// CheckBulkDeleteSafety checks if a bulk delete operation is safe
func CheckBulkDeleteSafety(totalFiles, deleteCount int) *BulkDeleteCheckResult {
	cfg := config.Current
	result := &BulkDeleteCheckResult{
		FileCount: deleteCount,
		Threshold: cfg.DeleteThresholdCount,
	}

	if deleteCount > cfg.DeleteThresholdCount {
		result.IsSafe = false
		result.Reason = fmt.Sprintf("Delete count (%d) exceeds absolute threshold (%d)", deleteCount, cfg.DeleteThresholdCount)
		return result
	}

	if totalFiles > 0 {
		percentage := float64(deleteCount) / float64(totalFiles) * 100
		result.Percentage = percentage

		if percentage > cfg.DeleteThresholdPercent {
			result.IsSafe = false
			result.Reason = fmt.Sprintf("Delete percentage (%.2f%%) exceeds threshold (%.0f%%) of total files (%d)", percentage, cfg.DeleteThresholdPercent, totalFiles)
			return result
		}
	}

	result.IsSafe = true
	result.Reason = "Delete operation is within safe limits"
	return result
}

// RecentDeletionWindow is how far back to count deletions for per-request API enforcement.
const RecentDeletionWindow = time.Hour
