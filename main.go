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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const audioFile = "recording.wav"

// TimeInputWidget is a custom widget for time input with increment/decrement controls
type TimeInputWidget struct {
	widget.BaseWidget

	minutes     int
	seconds     int
	maxDuration float64 // Maximum duration in seconds

	minutesEntry *widget.Entry
	secondsEntry *widget.Entry

	onChanged func(string)

	container *fyne.Container
}

// NewTimeInputWidget creates a new time input widget
func NewTimeInputWidget() *TimeInputWidget {
	w := &TimeInputWidget{
		minutes:     0,
		seconds:     0,
		maxDuration: 0,
	}
	w.ExtendBaseWidget(w)
	w.createUI()
	return w
}

// SetMaxDuration sets the maximum allowed duration
func (w *TimeInputWidget) SetMaxDuration(duration float64) {
	w.maxDuration = duration
	w.validateTime()
}

// SetOnChanged sets the callback for when the time changes
func (w *TimeInputWidget) SetOnChanged(callback func(string)) {
	w.onChanged = callback
}

// GetTime returns the current time as a formatted string (MM:SS)
func (w *TimeInputWidget) GetTime() string {
	return fmt.Sprintf("%02d:%02d", w.minutes, w.seconds)
}

// SetTime sets the time from a formatted string (MM:SS)
func (w *TimeInputWidget) SetTime(timeStr string) {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		if mins, err := strconv.Atoi(parts[0]); err == nil {
			w.minutes = mins
		}
		if secs, err := strconv.Atoi(parts[1]); err == nil {
			w.seconds = secs
		}
		w.validateTime()
		w.updateEntries()
		w.triggerChanged()
	}
}

// validateTime ensures the time doesn't exceed maxDuration
func (w *TimeInputWidget) validateTime() {
	if w.maxDuration > 0 {
		totalSeconds := w.minutes*60 + w.seconds
		maxSeconds := int(w.maxDuration)

		if totalSeconds > maxSeconds {
			w.minutes = maxSeconds / 60
			w.seconds = maxSeconds % 60
		}
	}

	// Ensure seconds don't exceed 59
	if w.seconds >= 60 {
		w.minutes += w.seconds / 60
		w.seconds = w.seconds % 60
	}

	// Ensure non-negative values
	if w.minutes < 0 {
		w.minutes = 0
	}
	if w.seconds < 0 {
		w.seconds = 0
	}
}

// updateEntries updates the entry widgets with current values
func (w *TimeInputWidget) updateEntries() {
	w.minutesEntry.SetText(fmt.Sprintf("%02d", w.minutes))
	w.secondsEntry.SetText(fmt.Sprintf("%02d", w.seconds))
}

// triggerChanged calls the onChanged callback if set
func (w *TimeInputWidget) triggerChanged() {
	if w.onChanged != nil {
		w.onChanged(w.GetTime())
	}
}

// createUI creates the UI components
func (w *TimeInputWidget) createUI() {
	// Create entry widgets
	w.minutesEntry = widget.NewEntry()
	w.minutesEntry.SetText("00")
	w.minutesEntry.OnChanged = func(text string) {
		if val, err := strconv.Atoi(text); err == nil && val >= 0 {
			w.minutes = val
			w.validateTime()
			w.updateEntries()
			w.triggerChanged()
		}
	}

	w.secondsEntry = widget.NewEntry()
	w.secondsEntry.SetText("00")
	w.secondsEntry.OnChanged = func(text string) {
		if val, err := strconv.Atoi(text); err == nil && val >= 0 && val < 60 {
			w.seconds = val
			w.validateTime()
			w.updateEntries()
			w.triggerChanged()
		}
	}

	// Create increment/decrement buttons for minutes
	minUpBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		w.minutes++
		w.validateTime()
		w.updateEntries()
		w.triggerChanged()
	})

	minDownBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		if w.minutes > 0 {
			w.minutes--
			w.validateTime()
			w.updateEntries()
			w.triggerChanged()
		}
	})

	// Create increment/decrement buttons for seconds
	secUpBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		w.seconds++
		w.validateTime()
		w.updateEntries()
		w.triggerChanged()
	})

	secDownBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		if w.seconds > 0 {
			w.seconds--
			w.validateTime()
			w.updateEntries()
			w.triggerChanged()
		}
	})

	// Create horizontal layout: ▲mm▼:▲ss▼
	colonLabel := widget.NewLabel(":")

	w.container = container.NewHBox(
		minUpBtn,
		w.minutesEntry,
		minDownBtn,
		colonLabel,
		secUpBtn,
		w.secondsEntry,
		secDownBtn,
	)
}

// CreateRenderer creates the renderer for this widget
func (w *TimeInputWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(w.container)
}

type AudioRecorder struct {
	isRecording  bool
	isPlayingRef bool
	isPlayingRec bool
	recordCmd    *exec.Cmd
	playRefCmd   *exec.Cmd
	playRecCmd   *exec.Cmd

	// Reference audio
	referenceFile string
	audioDuration float64
	filePathEntry *widget.Entry

	// UI elements
	openBtn        *widget.Button
	startSlider    *widget.Slider
	endSlider      *widget.Slider
	startTimeInput *TimeInputWidget
	endTimeInput   *TimeInputWidget
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

	// Initialize time sliders
	ar.startSlider = widget.NewSlider(0, 100)
	ar.startSlider.SetValue(0)
	ar.startSlider.OnChanged = func(value float64) {
		if ar.audioDuration > 0 {
			seconds := (value / 100) * ar.audioDuration
			ar.startTimeInput.SetTime(ar.formatTime(seconds))
		}
	}

	ar.endSlider = widget.NewSlider(0, 100)
	ar.endSlider.SetValue(100)
	ar.endSlider.OnChanged = func(value float64) {
		if ar.audioDuration > 0 {
			seconds := (value / 100) * ar.audioDuration
			ar.endTimeInput.SetTime(ar.formatTime(seconds))
		}
	}

	// Initialize custom time input widgets
	ar.startTimeInput = NewTimeInputWidget()
	ar.startTimeInput.SetOnChanged(func(timeStr string) {
		if ar.audioDuration > 0 {
			seconds := ar.parseTime(timeStr)
			if seconds >= 0 && seconds <= ar.audioDuration {
				percentage := (seconds / ar.audioDuration) * 100
				ar.startSlider.SetValue(percentage)
			}
		}
	})

	ar.endTimeInput = NewTimeInputWidget()
	ar.endTimeInput.SetOnChanged(func(timeStr string) {
		if ar.audioDuration > 0 {
			seconds := ar.parseTime(timeStr)
			if seconds >= 0 && seconds <= ar.audioDuration {
				percentage := (seconds / ar.audioDuration) * 100
				ar.endSlider.SetValue(percentage)
			}
		}
	})

	ar.playRefBtn = widget.NewButton("▶️ Play Reference", ar.toggleReferencePlayback)
	ar.playRefBtn.Disable()

	ar.recordBtn = widget.NewButton("🎤 Record", ar.toggleRecording)
	ar.playBtn = widget.NewButton("▶️ Play Recording", ar.toggleRecordingPlayback)

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
		ar.endTimeInput.SetTime(ar.formatTime(duration))
		ar.endTimeInput.SetMaxDuration(duration)
		ar.startTimeInput.SetMaxDuration(duration)
		ar.endSlider.SetValue(100)
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

func (ar *AudioRecorder) toggleReferencePlayback() {
	if ar.isPlayingRef {
		ar.stopReferencePlayback()
	} else {
		ar.playReference()
	}
}

func (ar *AudioRecorder) playReference() {
	if ar.referenceFile == "" {
		ar.statusLabel.SetText("No reference file loaded")
		return
	}

	startTime := ar.parseTime(ar.startTimeInput.GetTime())
	endTime := ar.parseTime(ar.endTimeInput.GetTime())

	if endTime <= startTime {
		ar.statusLabel.SetText("End time must be greater than start time")
		return
	}

	if endTime > ar.audioDuration {
		endTime = ar.audioDuration
		ar.endTimeInput.SetTime(ar.formatTime(endTime))
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

	if err := cmd.Start(); err != nil {
		ar.statusLabel.SetText("Error starting reference playback")
		return
	}

	ar.playRefCmd = cmd
	ar.isPlayingRef = true
	ar.playRefBtn.SetText("⏹️ Stop Reference")
	ar.statusLabel.SetText(fmt.Sprintf("Playing reference (%.1fs)", duration))

	go func() {
		cmd.Wait()
		fyne.Do(func() {
			ar.isPlayingRef = false
			ar.playRefBtn.SetText("▶️ Play Reference")
			ar.statusLabel.SetText("Reference playback finished")
		})
		ar.playRefCmd = nil
	}()
}

func (ar *AudioRecorder) stopReferencePlayback() {
	if ar.playRefCmd != nil {
		ar.playRefCmd.Process.Kill()
		ar.playRefCmd.Wait()
		ar.playRefCmd = nil
	}

	ar.isPlayingRef = false
	ar.playRefBtn.SetText("▶️ Play Reference")
	ar.statusLabel.SetText("Reference playback stopped")
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

func (ar *AudioRecorder) toggleRecordingPlayback() {
	if ar.isPlayingRec {
		ar.stopRecordingPlayback()
	} else {
		ar.playRecording()
	}
}

func (ar *AudioRecorder) playRecording() {
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		ar.statusLabel.SetText("No recording found")
		return
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ffplay", "-nodisp", "-autoexit", audioFile)
	case "darwin":
		cmd = exec.Command("afplay", audioFile)
	case "linux":
		cmd = exec.Command("aplay", audioFile)
	default:
		ar.statusLabel.SetText("Unsupported operating system for playback")
		return
	}

	if err := cmd.Start(); err != nil {
		ar.statusLabel.SetText("Error starting recording playback")
		return
	}

	ar.playRecCmd = cmd
	ar.isPlayingRec = true
	ar.playBtn.SetText("⏹️ Stop Recording")
	ar.statusLabel.SetText("Playing recording...")

	go func() {
		cmd.Wait()
		fyne.Do(func() {
			ar.isPlayingRec = false
			ar.playBtn.SetText("▶️ Play Recording")
			ar.statusLabel.SetText("Recording playback finished")
		})
		ar.playRecCmd = nil
	}()
}

func (ar *AudioRecorder) stopRecordingPlayback() {
	if ar.playRecCmd != nil {
		ar.playRecCmd.Process.Kill()
		ar.playRecCmd.Wait()
		ar.playRecCmd = nil
	}

	ar.isPlayingRec = false
	ar.playBtn.SetText("▶️ Play Recording")
	ar.statusLabel.SetText("Recording playback stopped")
}

func main() {
	myApp := app.NewWithID("com.audiorecorder.app")
	myWindow := myApp.NewWindow("Audio Recorder with Reference")
	myWindow.Resize(fyne.NewSize(600, 450))

	recorder := NewAudioRecorder(myWindow)

	// Reference audio section
	referenceSection := container.NewVBox(
		widget.NewLabel("Reference Audio:"),
		recorder.filePathEntry,
		recorder.openBtn,
		recorder.durationLabel,
		// Fixed layout: labels in one row, time inputs in next row, sliders in third row
		container.NewGridWithColumns(2,
			widget.NewLabel("Start Time:"),
			widget.NewLabel("End Time:"),
		),
		container.NewGridWithColumns(2,
			recorder.startTimeInput,
			recorder.endTimeInput,
		),
		container.NewGridWithColumns(2,
			recorder.startSlider,
			recorder.endSlider,
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
