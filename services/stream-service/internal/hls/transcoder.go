package hls

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type Transcoder struct {
	outputDir   string
	segmentTime int
	listSize    int
	mu          sync.Mutex
	sessions    map[string]*session
}

type session struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	done  chan struct{}
}

func NewTranscoder(outputDir string, segmentTime, listSize int, _ string) *Transcoder {
	return &Transcoder{
		outputDir:   outputDir,
		segmentTime: segmentTime,
		listSize:    listSize,
		sessions:    make(map[string]*session),
	}
}

// Start launches ffmpeg reading from a pipe (stdin).
// Returns a WriteCloser — the RTMP server writes raw FLV data into it.
func (t *Transcoder) Start(username string) (io.WriteCloser, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, exists := t.sessions[username]; exists {
		return nil, fmt.Errorf("transcoder already running for %s", username)
	}

	streamDir := filepath.Join(t.outputDir, username)
	if err := os.MkdirAll(streamDir, 0755); err != nil {
		return nil, fmt.Errorf("create hls dir: %w", err)
	}

	playlistPath := filepath.Join(streamDir, "index.m3u8")
	segmentPattern := filepath.Join(streamDir, "seg%03d.ts")

	// ffmpeg reads FLV from stdin and writes HLS segments
	args := []string{
		"-f", "flv",
		"-i", "pipe:0",
		// Re-encode video to H.264 Baseline — required for MSE/HLS.js compatibility
		// "copy" passes through High profile which some browsers can't buffer
		"-c:v", "libx264",
		"-profile:v", "baseline",
		"-level", "3.1",
		"-preset", "veryfast", // low CPU overhead
		"-tune", "zerolatency",
		"-c:a", "aac",
		"-ar", "44100",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", t.segmentTime),
		"-hls_list_size", fmt.Sprintf("%d", t.listSize),
		"-hls_flags", "delete_segments+append_list+independent_segments",
		"-hls_segment_type", "mpegts",
		"-start_number", "0",
		"-hls_segment_filename", segmentPattern,
		playlistPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("start ffmpeg: %w", err)
	}

	sess := &session{cmd: cmd, stdin: stdin, done: make(chan struct{})}
	go func() {
		cmd.Wait()
		close(sess.done)
		t.mu.Lock()
		delete(t.sessions, username)
		t.mu.Unlock()
		log.Printf("[HLS] ffmpeg exited for %s", username)
	}()

	t.sessions[username] = sess
	log.Printf("[HLS] transcoder started for %s → %s", username, playlistPath)
	return stdin, nil
}

func (t *Transcoder) Stop(username string) {
	t.mu.Lock()
	sess, exists := t.sessions[username]
	t.mu.Unlock()
	if !exists {
		return
	}
	sess.stdin.Close() // closing stdin signals ffmpeg to finish
	<-sess.done
	log.Printf("[HLS] transcoder stopped for %s", username)
}

func (t *Transcoder) IsRunning(username string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	_, exists := t.sessions[username]
	return exists
}
