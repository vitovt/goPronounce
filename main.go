package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const audioFile = "recording.wav"

type AudioRecorder struct {
	isRecording bool
	recordCmd   *exec.Cmd
	recordBtn   *widget.Button
	playBtn     *widget.Button
	statusLabel *widget.Label
}

func NewAudioRecorder() *AudioRecorder {
	ar := &AudioRecorder{
		statusLabel: widget.NewLabel("Ready to record"),
	}

	ar.recordBtn = widget.NewButton("🎤 Record", ar.toggleRecording)
	ar.playBtn = widget.NewButton("▶️ Play", ar.playRecording)

	// Initially disable play button if no recording exists
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		ar.playBtn.Disable()
	}

	return ar
}

func (ar *AudioRecorder) toggleRecording() {
	if ar.isRecording {
		ar.stopRecording()
	} else {
		ar.startRecording()
	}
}

func (ar *AudioRecorder) startRecording() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Use ffmpeg on Windows (requires ffmpeg to be installed)
		cmd = exec.Command("ffmpeg", "-f", "dshow", "-i", "audio=default", "-y", audioFile)
	case "darwin":
		// Use sox on macOS (requires sox to be installed)
		cmd = exec.Command("sox", "-t", "coreaudio", "default", audioFile)
	case "linux":
		// Use arecord on Linux (part of alsa-utils)
		cmd = exec.Command("arecord", "-f", "cd", "-t", "wav", audioFile)
	default:
		ar.statusLabel.SetText("Unsupported operating system")
		return
	}

	if err := cmd.Start(); err != nil {
		ar.statusLabel.SetText(fmt.Sprintf("Error starting recording: %v", err))
		return
	}

	ar.recordCmd = cmd
	ar.isRecording = true
	ar.recordBtn.SetText("⏹️ Stop")
	ar.playBtn.Disable()
	ar.statusLabel.SetText("Recording...")
}

func (ar *AudioRecorder) stopRecording() {
	if ar.recordCmd != nil {
		ar.recordCmd.Process.Kill()
		ar.recordCmd.Wait()
		ar.recordCmd = nil
	}

	ar.isRecording = false
	ar.recordBtn.SetText("🎤 Record")
	ar.playBtn.Enable()
	ar.statusLabel.SetText("Recording saved to " + audioFile)
}

func (ar *AudioRecorder) playRecording() {
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		ar.statusLabel.SetText("No recording found")
		return
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Use default Windows media player
		cmd = exec.Command("cmd", "/c", "start", audioFile)
	case "darwin":
		// Use afplay on macOS
		cmd = exec.Command("afplay", audioFile)
	case "linux":
		// Use aplay on Linux
		cmd = exec.Command("aplay", audioFile)
	default:
		ar.statusLabel.SetText("Unsupported operating system for playback")
		return
	}

	ar.statusLabel.SetText("Playing...")

	go func() {
		if err := cmd.Run(); err != nil {
			fyne.Do(func() {
				ar.statusLabel.SetText(fmt.Sprintf("Error playing: %v", err))
			})
		} else {
			fyne.Do(func() {
				ar.statusLabel.SetText("Playback finished")
			})
		}
	}()
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Simple Audio Recorder")
	myWindow.Resize(fyne.NewSize(300, 150))

	recorder := NewAudioRecorder()

	content := container.NewVBox(
		recorder.statusLabel,
		container.NewHBox(
			recorder.recordBtn,
			recorder.playBtn,
		),
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
