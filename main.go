package main

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	fynedialog "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	nativedialog "github.com/sqweek/dialog"
)

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
			gp.startTimeInput.SetTime(formatTime(seconds))
		}
	}

	gp.endSlider = widget.NewSlider(0, 100)
	gp.endSlider.SetValue(100)
	gp.endSlider.OnChanged = func(value float64) {
		if gp.audioDuration > 0 {
			seconds := (value / 100) * gp.audioDuration
			gp.endTimeInput.SetTime(formatTime(seconds))
		}
	}

	// Initialize custom time input widgets
	gp.startTimeInput = NewTimeInputWidget()
	gp.startTimeInput.SetEnabled(false)
	gp.startSlider.Disable()
	gp.startTimeInput.SetOnChanged(func(timeStr string) {
		if gp.audioDuration > 0 {
			seconds := parseTime(timeStr)
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
			seconds := parseTime(timeStr)
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
	w.Resize(fyne.NewSize(680, 360))
	w.CenterOnScreen()
	w.Show()
}

func main() {
	myApp := app.NewWithID("com.gopronounce.app")
	myWindow := myApp.NewWindow("Trainer for sharpening pronunciation")
	myWindow.Resize(fyne.NewSize(600, 450))

	recorder := NewGoPronounce(myWindow)

	//make sure every external audio process dies when the user quits.
	myWindow.SetCloseIntercept(func() {
		recorder.StopAllAudio() // defined in audio.go
		myApp.Quit()            // closes the window afterwards
	})

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
