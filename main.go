package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	fynedialog "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nativedialog "github.com/sqweek/dialog"
)

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
	settingsBtn    *widget.Button
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

	gp.settingsBtn = widget.NewButtonWithIcon("⚙ Settings…", theme.SettingsIcon(), func() {
		gp.showInputSettings()
	})

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
	filePath, err := nativedialog.
		File().
		Title("Select audio file").
		Filter("Audio", "wav", "mp3", "flac", "ogg").
		Load()
	if err != nil {
		// nativedialog.Message("%v", err).Title("Error").Error()
		fynedialog.ShowError(err, gp.window)
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

func (gp *GoPronounce) showInputSettings() {
	devs, err := listInputDevices()
	if err != nil || len(devs) == 0 {
		fynedialog.ShowError(fmt.Errorf("no capture devices found"), gp.window)
		return
	}

	const defLabel = "<system default>"
	devs = append([]string{defLabel}, devs...)

	selectBox := widget.NewSelect(devs, func(sel string) {
		if sel == defLabel {
			recordDevice = "" // reset to default
		} else {
			recordDevice = sel // explicit device
		}
	})
	if recordDevice == "" {
		selectBox.SetSelected(defLabel)
	} else {
		selectBox.SetSelected(recordDevice)
	}
	log.Printf("recordDevice: %v", recordDevice)

	w := fyne.CurrentApp().NewWindow("Input device")
	w.SetContent(container.NewVBox(
		widget.NewLabel("Choose recording source:"),
		selectBox,
		widget.NewButton("OK", func() { w.Close() }),
	))
	w.Resize(fyne.NewSize(380, 160))
	w.CenterOnScreen()
	w.Show()
}

func main() {
	myApp := app.NewWithID("com.gopronounce.app")
	myWindow := myApp.NewWindow("Trainer for sharpening pronunciation")
	myWindow.Resize(fyne.NewSize(600, 450))

	recorder := NewGoPronounce(myWindow)

	// Reference audio section
	referenceSection := container.NewVBox(
		container.NewGridWithColumns(2,
			widget.NewLabel("Reference Audio:"),
			recorder.settingsBtn,
		),
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
