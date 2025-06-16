package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const audioFile = "recording.wav"

type AudioRecorder struct {
	isRecording bool
	recordCmd   *exec.Cmd
	playCmd     *exec.Cmd

	// Reference audio
	referenceFile string
	audioDuration float64
	filePathEntry *widget.Entry

	// UI elements
	openBtn        *widget.Button
	progressBar    *widget.ProgressBar
	startTimeEntry *widget.Entry
	endTimeEntry   *widget.Entry
	playRefBtn     *widget.Button
	recordBtn      *widget.Button
	playBtn        *widget.Button
	statusLabel    *widget.Label
	durationLabel  *widget.Label

	// Main window reference
	window fyne.Window
}

func NewAudioRecorder(window fyne.Window) *AudioRecorder {
	ar := &AudioRecorder{
		window:        window,
		statusLabel:   widget.NewLabel("Ready to record"),
		durationLabel: widget.NewLabel("No reference file loaded"),
	}

	ar.openBtn = widget.NewButton("📁 Browse", ar.openReferenceFile)

	// Add file path entry
	ar.filePathEntry = widget.NewEntry()
	ar.filePathEntry.SetPlaceHolder("Enter audio file path or use Browse button")
	ar.filePathEntry.OnChanged = func(path string) {
		if path != "" && ar.fileExists(path) {
			ar.referenceFile = path
			ar.statusLabel.SetText("Loading reference file...")
			go ar.getAudioDuration()
		}
	}
	ar.progressBar = widget.NewProgressBar()
	ar.progressBar.SetValue(0)

	ar.startTimeEntry = widget.NewEntry()
	ar.startTimeEntry.SetText("00:00")
	ar.startTimeEntry.SetPlaceHolder("00:00")

	ar.endTimeEntry = widget.NewEntry()
	ar.endTimeEntry.SetText("00:00")
	ar.endTimeEntry.SetPlaceHolder("00:00")

	ar.playRefBtn = widget.NewButton("▶️ Play Reference", ar.playReference)
	ar.playRefBtn.Disable()

	ar.recordBtn = widget.NewButton("🎤 Record", ar.toggleRecording)
	ar.playBtn = widget.NewButton("▶️ Play Recording", ar.playRecording)

	// Initially disable play button if no recording exists
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		ar.playBtn.Disable()
	}

	return ar
}

func (ar *AudioRecorder) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (ar *AudioRecorder) openReferenceFile() {
	// Try to use native dialog, but with a timeout fallback
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fyne.Do(func() {
					ar.statusLabel.SetText("File dialog failed. Please type file path manually.")
				})
			}
		}()

		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			filePath := reader.URI().Path()
			ar.filePathEntry.SetText(filePath)
			ar.referenceFile = filePath
			ar.statusLabel.SetText("Loading reference file...")

			go ar.getAudioDuration()

		}, ar.window)
	}()
}

func (ar *AudioRecorder) getAudioDuration() {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries",
		"format=duration", "-of", "csv=p=0", ar.referenceFile)

	output, err := cmd.Output()
	if err != nil {
		fyne.Do(func() {
			ar.statusLabel.SetText("Error reading audio file. Make sure ffmpeg is installed.")
		})
		return
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		fyne.Do(func() {
			ar.statusLabel.SetText("Error parsing audio duration")
		})
		return
	}

	ar.audioDuration = duration

	fyne.Do(func() {
		ar.durationLabel.SetText(fmt.Sprintf("Duration: %s", ar.formatTime(duration)))
		ar.endTimeEntry.SetText(ar.formatTime(duration))
		ar.playRefBtn.Enable()
		ar.statusLabel.SetText(fmt.Sprintf("Reference loaded: %s", filepath.Base(ar.referenceFile)))
	})
}

func (ar *AudioRecorder) formatTime(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func (ar *AudioRecorder) parseTime(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}

	mins, _ := strconv.Atoi(parts[0])
	secs, _ := strconv.Atoi(parts[1])
	return float64(mins*60 + secs)
}

func (ar *AudioRecorder) playReference() {
	if ar.referenceFile == "" {
		ar.statusLabel.SetText("No reference file loaded")
		return
	}

	startTime := ar.parseTime(ar.startTimeEntry.Text)
	endTime := ar.parseTime(ar.endTimeEntry.Text)

	if endTime <= startTime {
		ar.statusLabel.SetText("End time must be greater than start time")
		return
	}

	if endTime > ar.audioDuration {
		endTime = ar.audioDuration
		ar.endTimeEntry.SetText(ar.formatTime(endTime))
	}

	duration := endTime - startTime

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ffplay", "-ss", fmt.Sprintf("%.2f", startTime),
			"-t", fmt.Sprintf("%.2f", duration), "-nodisp", "-autoexit", ar.referenceFile)
	case "darwin":
		// Create a temporary file with the segment for macOS
		tempFile := "temp_segment.wav"
		extractCmd := exec.Command("ffmpeg", "-ss", fmt.Sprintf("%.2f", startTime),
			"-t", fmt.Sprintf("%.2f", duration), "-i", ar.referenceFile, "-y", tempFile)
		if err := extractCmd.Run(); err != nil {
			ar.statusLabel.SetText("Error extracting audio segment")
			return
		}
		cmd = exec.Command("afplay", tempFile)
		defer os.Remove(tempFile) // Clean up temp file
	case "linux":
		cmd = exec.Command("ffplay", "-ss", fmt.Sprintf("%.2f", startTime),
			"-t", fmt.Sprintf("%.2f", duration), "-nodisp", "-autoexit", ar.referenceFile)
	default:
		ar.statusLabel.SetText("Unsupported operating system for reference playback")
		return
	}

	ar.playCmd = cmd
	ar.statusLabel.SetText(fmt.Sprintf("Playing reference (%.1fs)", duration))

	go func() {
		if err := cmd.Run(); err != nil {
			fyne.Do(func() {
				ar.statusLabel.SetText("Error playing reference")
			})
		} else {
			fyne.Do(func() {
				ar.statusLabel.SetText("Reference playback finished")
			})
		}
		ar.playCmd = nil
	}()
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
		cmd = exec.Command("ffmpeg", "-f", "dshow", "-i", "audio=default", "-y", audioFile)
	case "darwin":
		cmd = exec.Command("sox", "-t", "coreaudio", "default", audioFile)
	case "linux":
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
		cmd = exec.Command("cmd", "/c", "start", audioFile)
	case "darwin":
		cmd = exec.Command("afplay", audioFile)
	case "linux":
		cmd = exec.Command("aplay", audioFile)
	default:
		ar.statusLabel.SetText("Unsupported operating system for playback")
		return
	}

	ar.statusLabel.SetText("Playing recording...")

	go func() {
		if err := cmd.Run(); err != nil {
			fyne.Do(func() {
				ar.statusLabel.SetText(fmt.Sprintf("Error playing: %v", err))
			})
		} else {
			fyne.Do(func() {
				ar.statusLabel.SetText("Recording playback finished")
			})
		}
	}()
}

func main() {
	myApp := app.NewWithID("com.audiorecorder.app")
	myWindow := myApp.NewWindow("Audio Recorder with Reference")
	myWindow.Resize(fyne.NewSize(450, 400))

	recorder := NewAudioRecorder(myWindow)

	// Reference audio section
	referenceSection := container.NewVBox(
		widget.NewLabel("Reference Audio:"),
		recorder.filePathEntry,
		recorder.openBtn,
		recorder.durationLabel,
		recorder.progressBar,
		container.NewGridWithColumns(2,
			container.NewVBox(
				widget.NewLabel("Start Time:"),
				recorder.startTimeEntry,
			),
			container.NewVBox(
				widget.NewLabel("End Time:"),
				recorder.endTimeEntry,
			),
		),
		recorder.playRefBtn,
		widget.NewSeparator(),
	)

	// Recording section
	recordingSection := container.NewVBox(
		widget.NewLabel("Recording:"),
		container.NewHBox(
			recorder.recordBtn,
			recorder.playBtn,
		),
	)

	content := container.NewVBox(
		referenceSection,
		recordingSection,
		widget.NewSeparator(),
		recorder.statusLabel,
	)

	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}
