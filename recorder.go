package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

type Recorder struct {
	api         *TwitchAPI
	username    string
	quality     string
	tempPath    string
	finalPath   string
	webAPIToken string
	refresh     time.Duration
}

func NewRecorder(
	api *TwitchAPI,
	username, quality, tempPath, finalPath, webAPIToken string,
	refresh time.Duration,
) *Recorder {
	tempPath = path.Join(tempPath, username)
	finalPath = path.Join(finalPath, username)

	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		if err := os.MkdirAll(tempPath, 0o755); err != nil {
			log.Fatalf("Failed to create temp directory: %s", err)
		}
	}
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		if err := os.MkdirAll(finalPath, 0o755); err != nil {
			log.Fatalf("Failed to create final directory: %s", err)
		}
	}

	return &Recorder{
		api:         api,
		username:    username,
		quality:     quality,
		tempPath:    tempPath,
		finalPath:   finalPath,
		webAPIToken: webAPIToken,
		refresh:     refresh,
	}
}

func (r *Recorder) recordStream(ctx context.Context, filePath string) error {
	cmd := exec.CommandContext(ctx,
		"streamlink",
		"--loglevel", "error",
		"--twitch-api-header", fmt.Sprintf("Authorization=OAuth %s", r.webAPIToken),
		fmt.Sprintf("twitch.tv/%s", r.username),
		r.quality,
		"-f", "-o", filePath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (*Recorder) fixStream(
	tempFilePath, finalFilePath string,
) error {
	cmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-err_detect", "ignore_err",
		"-i", tempFilePath,
		"-c", "copy",
		finalFilePath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (r *Recorder) cleanup() error {
	files, err := os.ReadDir(r.tempPath)
	if err != nil {
		return fmt.Errorf("failed to read temp directory: %w", err)
	}

	for _, file := range files {
		err = os.Remove(path.Join(r.tempPath, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to remove temp file %s: %w", file.Name(), err)
		}
	}

	return nil
}

func (r *Recorder) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down recorder...")
			err := r.cleanup()
			if err != nil {
				return fmt.Errorf("failed to cleanup: %w", err)
			}
			return nil
		case <-ticker.C:
			if ctx.Err() != nil {
				return nil
			}

			err := r.checkAndRecord(ctx)
			if err != nil {
				return fmt.Errorf("failed to check and record: %w", err)
			}
		}
	}
}

func (r *Recorder) checkAndRecord(ctx context.Context) error {
	accessToken, err := r.api.fetchAccessToken()
	if err != nil {
		return fmt.Errorf("failed to fetch access token: %w", err)
	}

	isOnline, err := r.api.isStreamOnline(r.username, accessToken)
	if err != nil {
		return fmt.Errorf("failed to check stream status: %w", err)
	}

	if !isOnline {
		log.Printf(
			"%s offline, checking again in %s\n",
			r.username,
			r.refresh.String(),
		)
		return nil
	}

	tempFileName := time.Now().Format("2006-01-02_15-04-05") + ".mp4"
	tempFilePath := path.Join(r.tempPath, tempFileName)

	log.Printf("Starting recording of %s\n", r.username)
	err = r.recordStream(ctx, tempFilePath)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("failed to record stream: %w", err)
	}

	finalFileName := fmt.Sprintf(
		"%s__%s.mp4",
		strings.TrimSuffix(tempFileName, ".mp4"),
		time.Now().Format("2006-01-02_15-04-05"),
	)
	finalFilePath := path.Join(r.finalPath, finalFileName)

	log.Println("Recording finished, fixing stream...")
	err = r.fixStream(tempFilePath, finalFilePath)
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("failed to fix stream: %w", err)
	}

	log.Printf("Stream saved to: %s\n", finalFilePath)
	return r.cleanup()
}
