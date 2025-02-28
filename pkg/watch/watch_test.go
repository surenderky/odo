package watch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/fsnotify/fsnotify"
)

func evaluateChangesHandler(events []fsnotify.Event, path string, fileIgnores []string, watcher *fsnotify.Watcher) ([]string, []string) {
	var changedFiles []string
	var deletedPaths []string

	for _, event := range events {
		for _, file := range fileIgnores {
			// if file is in fileIgnores, don't do anything
			if event.Name == file {
				continue
			}
		}
		// this if condition is copied from original implementation in watch.go. Code within the if block is simplified for tests
		if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
			// On remove/rename, stop watching the resource
			_ = watcher.Remove(event.Name)
			// Append the file to list of deleted files
			deletedPaths = append(deletedPaths, event.Name)
			continue
		}
		// this if condition is copied from original implementation in watch.go. Code within the if block is simplified for tests
		if event.Op&fsnotify.Remove != fsnotify.Remove {
			changedFiles = append(changedFiles, event.Name)
		}
	}
	return changedFiles, deletedPaths
}

func processEventsHandler(changedFiles, deletedPaths []string, _ WatchParameters, out io.Writer) {
	fmt.Fprintf(out, "changedFiles %s deletedPaths %s\n", changedFiles, deletedPaths)
}

func cleanupHandler(_ parser.DevfileObj, out io.Writer) error {
	fmt.Fprintf(out, "cleanup done")
	return nil
}

func Test_eventWatcher(t *testing.T) {
	type args struct {
		parameters WatchParameters
	}
	tests := []struct {
		name          string
		args          args
		wantOut       string
		wantErr       bool
		watcherEvents []fsnotify.Event
		watcherError  error
	}{
		{
			name: "Case 1: Multiple events, no errors",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       "changedFiles [file1 file2] deletedPaths []\ncleanup done",
			wantErr:       false,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Create}, {Name: "file2", Op: fsnotify.Write}},
			watcherError:  nil,
		},
		{
			name: "Case 2: Multiple events, one error",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       "",
			wantErr:       true,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Create}, {Name: "file2", Op: fsnotify.Write}},
			watcherError:  fmt.Errorf("error"),
		},
		{
			name: "Case 3: Delete file, no error",
			args: args{
				parameters: WatchParameters{FileIgnores: []string{"file1"}},
			},
			wantOut:       "changedFiles [] deletedPaths [file1 file2]\ncleanup done",
			wantErr:       false,
			watcherEvents: []fsnotify.Event{{Name: "file1", Op: fsnotify.Remove}, {Name: "file2", Op: fsnotify.Rename}},
			watcherError:  nil,
		},
		{
			name: "Case 4: Only errors",
			args: args{
				parameters: WatchParameters{},
			},
			wantOut:       "",
			wantErr:       true,
			watcherEvents: nil,
			watcherError:  fmt.Errorf("error1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, _ := fsnotify.NewWatcher()
			var cancel context.CancelFunc
			ctx, cancel := context.WithCancel(context.Background())
			out := &bytes.Buffer{}

			go func() {
				for _, event := range tt.watcherEvents {
					watcher.Events <- event
				}

				if tt.watcherError != nil {
					watcher.Errors <- tt.watcherError
				}
				<-time.After(500 * time.Millisecond)
				cancel()
			}()

			err := eventWatcher(ctx, watcher, tt.args.parameters, out, evaluateChangesHandler, processEventsHandler, cleanupHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventWatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("eventWatcher() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
