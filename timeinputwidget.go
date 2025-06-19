package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
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
