package reload

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

// TestNewFileWatcherNotifier tests the creation of a new FileWatcherNotifier.
func TestNewFileWatcherNotifier(t *testing.T) {
	filePath := "test.txt"
	reloadDelay := 100 * time.Millisecond

	notifier, err := NewFileWatcherNotifier(filePath, reloadDelay)
	assert.NoError(t, err)
	assert.NotNil(t, notifier)
	assert.Equal(t, filePath, notifier.filePath)
	assert.Equal(t, reloadDelay, notifier.reloadDelay)
}

// TestWatchForChange tests the WatchForChange method.
func TestWatchForChange(t *testing.T) {
	file, _ := os.CreateTemp("", "test.txt")
	filePath := file.Name()
	defer os.Remove(filePath)

	reloadDelay := 100 * time.Millisecond
	notifier, _ := NewFileWatcherNotifier(filePath, reloadDelay)

	// Create a context that will be canceled after a short duration
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Variables to track reload success and the filePath passed
	reloadSuccessful := false
	var receivedFilePath string

	// Mock the onReload function
	onReload := func(filePath string) error {
		reloadSuccessful = true     // Mark reload as successful
		receivedFilePath = filePath // Capture the received filePath
		return nil
	}

	// Start watching for changes in a separate goroutine
	go func() {
		_ = notifier.WatchForChange(ctx, onReload)
	}()

	// Simulate a file change event
	notifier.watcher.Events <- fsnotify.Event{Op: fsnotify.Write, Name: filePath}

	// Wait for a bit to allow the reload to be triggered
	time.Sleep(150 * time.Millisecond)

	// Assert that the reload was successful
	assert.True(t, reloadSuccessful, "Expected reload to be successful")
	// Assert that the correct filePath was passed to the onReload function
	assert.Equal(t, filePath, receivedFilePath, "Expected filePath to match")

	// Clean up
	_ = notifier.Close()
}