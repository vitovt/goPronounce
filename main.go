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
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

const audioFileName = "goPronounceRecording.wav"

var audioFile = filepath.Join(os.TempDir(), audioFileName)

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

func (w *TimeInputWidget) SetEnabled(enabled bool) {
	for _, o := range w.container.Objects {
		if d, ok := o.(fyne.Disableable); ok {
			if enabled {
				d.Enable()
			} else {
				d.Disable()
			}
		}
	}
	w.container.Refresh()
}

type GoPronounce struct {
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

func NewGoPronounce(window fyne.Window) *GoPronounce {
	gp := &GoPronounce{
		window:        window,
		statusLabel:   widget.NewLabel("Ready to record"),
		durationLabel: widget.NewLabel("No reference file loaded"),
	}

	gp.openBtn = widget.NewButton("📁 Browse", gp.openNativePicker)

	// Add file path entry
	gp.filePathEntry = widget.NewEntry()
	gp.filePathEntry.SetPlaceHolder("Enter audio file path or use Browse button")
	gp.filePathEntry.Disable()
	gp.filePathEntry.OnChanged = func(path string) {
		if path != "" && gp.fileExists(path) {
			gp.referenceFile = path
			gp.statusLabel.SetText("Loading reference file...")
			go gp.getAudioDuration()
		}
	}

	// Initialize time sliders
	gp.startSlider = widget.NewSlider(0, 100)
	gp.startSlider.SetValue(0)
	gp.startSlider.OnChanged = func(value float64) {
		if gp.audioDuration > 0 {
			seconds := (value / 100) * gp.audioDuration
			gp.startTimeInput.SetTime(gp.formatTime(seconds))
		}
	}

	gp.endSlider = widget.NewSlider(0, 100)
	gp.endSlider.SetValue(100)
	gp.endSlider.OnChanged = func(value float64) {
		if gp.audioDuration > 0 {
			seconds := (value / 100) * gp.audioDuration
			gp.endTimeInput.SetTime(gp.formatTime(seconds))
		}
	}

	// Initialize custom time input widgets
	gp.startTimeInput = NewTimeInputWidget()
	gp.startTimeInput.SetEnabled(false)
	gp.startSlider.Disable()
	gp.startTimeInput.SetOnChanged(func(timeStr string) {
		if gp.audioDuration > 0 {
			seconds := gp.parseTime(timeStr)
			if seconds >= 0 && seconds <= gp.audioDuration {
				percentage := (seconds / gp.audioDuration) * 100
				gp.startSlider.SetValue(percentage)
			}
		}
	})

	gp.endTimeInput = NewTimeInputWidget()
	gp.endTimeInput.SetEnabled(false)
	gp.endSlider.Disable()
	gp.endTimeInput.SetOnChanged(func(timeStr string) {
		if gp.audioDuration > 0 {
			seconds := gp.parseTime(timeStr)
			if seconds >= 0 && seconds <= gp.audioDuration {
				percentage := (seconds / gp.audioDuration) * 100
				gp.endSlider.SetValue(percentage)
			}
		}
	})

	gp.playRefBtn = widget.NewButton("▶️ Play Reference", gp.toggleReferencePlayback)
	gp.playRefBtn.Disable()

	gp.recordBtn = widget.NewButton("🎤 Record", gp.toggleRecording)
	gp.playBtn = widget.NewButton("▶️ Play Recording", gp.toggleRecordingPlayback)

	// Initially disable play button if no recording exists
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		gp.playBtn.Disable()
	}

	return gp
}

func (gp *GoPronounce) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (gp *GoPronounce) openNativePicker() {
	filePath, err := dialog.
		File().
		Title("Select audio file").
		Filter("Audio", "wav", "mp3", "flac", "ogg").
		Load()
	if err != nil {
		if err != dialog.Cancelled {
			dialog.Message("%v", err).Title("Error").Error() // optional GUI error
		}
		return
	}
	// `path` is an absolute path string.
	fmt.Println("Selected:", filePath)
	gp.filePathEntry.SetText(filePath)
	gp.referenceFile = filePath
	gp.statusLabel.SetText("Loading reference file...")
	// Load the audio duration in a goroutine
	//go gp.getAudioDuration()
}

func (gp *GoPronounce) getAudioDuration() {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-show_entries",
		"format=duration", "-of", "csv=p=0", gp.referenceFile)

	output, err := cmd.Output()
	if err != nil {
		fyne.Do(func() {
			gp.statusLabel.SetText("Error reading audio file. Make sure ffmpeg is installed.")
		})
		return
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		fyne.Do(func() {
			gp.statusLabel.SetText("Error parsing audio duration")
		})
		return
	}

	gp.audioDuration = duration

	fyne.Do(func() {
		gp.durationLabel.SetText(fmt.Sprintf("Duration: %s", gp.formatTime(duration)))
		gp.endTimeInput.SetTime(gp.formatTime(duration))
		gp.endTimeInput.SetMaxDuration(duration)
		gp.startTimeInput.SetMaxDuration(duration)
		gp.endSlider.SetValue(100)
		gp.startTimeInput.SetEnabled(true)
		gp.endTimeInput.SetEnabled(true)
		gp.startSlider.Enable()
		gp.endSlider.Enable()
		gp.playRefBtn.Enable()
		gp.statusLabel.SetText(fmt.Sprintf("Reference loaded: %s", filepath.Base(gp.referenceFile)))
	})
}

func (gp *GoPronounce) formatTime(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func (gp *GoPronounce) parseTime(timeStr string) float64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}

	mins, _ := strconv.Atoi(parts[0])
	secs, _ := strconv.Atoi(parts[1])
	return float64(mins*60 + secs)
}

func (gp *GoPronounce) toggleReferencePlayback() {
	if gp.isPlayingRef {
		gp.stopReferencePlayback()
	} else {
		gp.playReference()
	}
}

func (gp *GoPronounce) playReference() {
	if gp.referenceFile == "" {
		gp.statusLabel.SetText("No reference file loaded")
		return
	}

	startTime := gp.parseTime(gp.startTimeInput.GetTime())
	endTime := gp.parseTime(gp.endTimeInput.GetTime())

	if endTime <= startTime {
		gp.statusLabel.SetText("End time must be greater than start time")
		return
	}

	if endTime > gp.audioDuration {
		endTime = gp.audioDuration
		gp.endTimeInput.SetTime(gp.formatTime(endTime))
	}

	duration := endTime - startTime

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ffplay", "-ss", fmt.Sprintf("%.2f", startTime),
			"-t", fmt.Sprintf("%.2f", duration), "-nodisp", "-autoexit", gp.referenceFile)
	case "darwin":
		// Create a temporary file with the segment for macOS
		tempFile := "temp_segment.wav"
		extractCmd := exec.Command("ffmpeg", "-ss", fmt.Sprintf("%.2f", startTime),
			"-t", fmt.Sprintf("%.2f", duration), "-i", gp.referenceFile, "-y", tempFile)
		if err := extractCmd.Run(); err != nil {
			gp.statusLabel.SetText("Error extracting audio segment")
			return
		}
		cmd = exec.Command("afplay", tempFile)
		defer os.Remove(tempFile) // Clean up temp file
	case "linux":
		cmd = exec.Command("ffplay", "-ss", fmt.Sprintf("%.2f", startTime),
			"-t", fmt.Sprintf("%.2f", duration), "-nodisp", "-autoexit", gp.referenceFile)
	default:
		gp.statusLabel.SetText("Unsupported operating system for reference playback")
		return
	}

	if err := cmd.Start(); err != nil {
		gp.statusLabel.SetText("Error starting reference playback")
		return
	}

	gp.playRefCmd = cmd
	gp.isPlayingRef = true
	gp.playRefBtn.SetText("⏹️ Stop Reference")
	gp.statusLabel.SetText(fmt.Sprintf("Playing reference (%.1fs)", duration))

	go func() {
		cmd.Wait()
		fyne.Do(func() {
			gp.isPlayingRef = false
			gp.playRefBtn.SetText("▶️ Play Reference")
			gp.statusLabel.SetText("Reference playback finished")
		})
		gp.playRefCmd = nil
	}()
}

func (gp *GoPronounce) stopReferencePlayback() {
	if gp.playRefCmd != nil {
		gp.playRefCmd.Process.Kill()
		gp.playRefCmd.Wait()
		gp.playRefCmd = nil
	}

	gp.isPlayingRef = false
	gp.playRefBtn.SetText("▶️ Play Reference")
	gp.statusLabel.SetText("Reference playback stopped")
}

func (gp *GoPronounce) toggleRecording() {
	if gp.isRecording {
		gp.stopRecording()
	} else {
		gp.startRecording()
	}
}

func (gp *GoPronounce) startRecording() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ffmpeg", "-f", "dshow", "-i", "audio=default", "-y", audioFile)
	case "darwin":
		cmd = exec.Command("sox", "-t", "coreaudio", "default", audioFile)
	case "linux":
		cmd = exec.Command("arecord", "-f", "cd", "-t", "wav", audioFile)
	default:
		gp.statusLabel.SetText("Unsupported operating system")
		return
	}

	if err := cmd.Start(); err != nil {
		gp.statusLabel.SetText(fmt.Sprintf("Error starting recording: %v", err))
		return
	}

	gp.recordCmd = cmd
	gp.isRecording = true
	gp.recordBtn.SetText("⏹️ Stop")
	gp.playBtn.Disable()
	gp.statusLabel.SetText("Recording...")
}

func (gp *GoPronounce) stopRecording() {
	if gp.recordCmd != nil {
		gp.recordCmd.Process.Kill()
		gp.recordCmd.Wait()
		gp.recordCmd = nil
	}

	gp.isRecording = false
	gp.recordBtn.SetText("🎤 Record")
	gp.playBtn.Enable()
	gp.statusLabel.SetText("Recording saved to " + audioFile)
}

func (gp *GoPronounce) toggleRecordingPlayback() {
	if gp.isPlayingRec {
		gp.stopRecordingPlayback()
	} else {
		gp.playRecording()
	}
}

func (gp *GoPronounce) playRecording() {
	if _, err := os.Stat(audioFile); os.IsNotExist(err) {
		gp.statusLabel.SetText("No recording found")
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
		gp.statusLabel.SetText("Unsupported operating system for playback")
		return
	}

	if err := cmd.Start(); err != nil {
		gp.statusLabel.SetText("Error starting recording playback")
		return
	}

	gp.playRecCmd = cmd
	gp.isPlayingRec = true
	gp.playBtn.SetText("⏹️ Stop Playing")
	gp.statusLabel.SetText("Playing recording...")

	go func() {
		cmd.Wait()
		fyne.Do(func() {
			gp.isPlayingRec = false
			gp.playBtn.SetText("▶️ Play Recording")
			gp.statusLabel.SetText("Recording playback finished")
		})
		gp.playRecCmd = nil
	}()
}

func (gp *GoPronounce) stopRecordingPlayback() {
	if gp.playRecCmd != nil {
		gp.playRecCmd.Process.Kill()
		gp.playRecCmd.Wait()
		gp.playRecCmd = nil
	}

	gp.isPlayingRec = false
	gp.playBtn.SetText("▶️ Play Recording")
	gp.statusLabel.SetText("Recording playback stopped")
}

func main() {
	myApp := app.NewWithID("com.audiorecorder.app")
	myWindow := myApp.NewWindow("Audio Recorder with Reference")
	myWindow.Resize(fyne.NewSize(600, 450))

	recorder := NewGoPronounce(myWindow)

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
