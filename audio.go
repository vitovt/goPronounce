package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
)

const audioFileName = "goPronounceRecording.wav"

var audioFile = filepath.Join(os.TempDir(), audioFileName)
var recordDevice string

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
		gp.durationLabel.SetText(fmt.Sprintf("Duration: %s", formatTime(duration)))
		gp.endTimeInput.SetTime(formatTime(duration))
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

func formatTime(seconds float64) string {
	mins := int(seconds) / 60
	secs := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d", mins, secs)
}

func parseTime(timeStr string) float64 {
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

	startTime := parseTime(gp.startTimeInput.GetTime())
	endTime := parseTime(gp.endTimeInput.GetTime())

	if endTime <= startTime {
		gp.statusLabel.SetText("End time must be greater than start time")
		return
	}

	if endTime > gp.audioDuration {
		endTime = gp.audioDuration
		gp.endTimeInput.SetTime(formatTime(endTime))
	}

	duration := endTime - startTime

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ffplay", "-ss", fmt.Sprintf("%.2f", startTime),
			"-t", fmt.Sprintf("%.2f", duration), "-nodisp", "-autoexit", gp.referenceFile)
	case "darwin":
		// Create a temporary file with the segment for macOS
		tempFile := filepath.Join(os.TempDir(), "temp_segment.wav")
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
	var commandArgs []string

	switch runtime.GOOS {
	case "windows":
		source := "audio=default"
		if recordDevice != "" {
			source = fmt.Sprintf("%s", recordDevice)
		}
		commandArgs = []string{"ffmpeg", "-f", "dshow", "-i", source, "-y", audioFile}

	case "darwin":
		source := "default"
		if recordDevice != "" {
			source = recordDevice
		}
		commandArgs = []string{"sox", "-t", "coreaudio", source, audioFile}

	case "linux":
		source := "default"
		if recordDevice != "" {
			source = fmt.Sprintf("%s", recordDevice)
		}
		// record through Pulse/PipeWire so virtual sources (rnnoise, monitors) work
		commandArgs = []string{"ffmpeg", "-f", "pulse",
			"-i", fmt.Sprintf("%s", source),
			"-ac", "1", "-y", audioFile}
	default:
		gp.statusLabel.SetText("Unsupported operating system")
		return
	}

	log.Printf("Executing Command: %v", commandArgs)
	cmd = exec.Command(commandArgs[0], commandArgs[1:]...)

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

func listInputDevices() ([]string, error) {
	switch runtime.GOOS {
	case "linux":
		out, err := exec.Command("pactl", "list", "short", "sources").Output()
		if err != nil {
			return nil, err
		}
		var devs []string
		for _, ln := range strings.Split(string(out), "\n") {
			f := strings.Fields(ln)
			if len(f) >= 2 {
				devs = append(devs, f[1]) // column “name”
			}
		}
		sort.Strings(devs)
		return devs, nil

	case "windows":
		out, _ := exec.Command("ffmpeg", "-list_devices", "true",
			"-f", "dshow", "-i", "dummy").CombinedOutput()
		re := regexp.MustCompile(`"(.+?)"$`)
		m := re.FindAllStringSubmatch(string(out), -1)
		var devs []string
		for _, v := range m {
			devs = append(devs, v[1])
		}
		return devs, nil

	case "darwin":
		out, _ := exec.Command("ffmpeg", "-f", "avfoundation",
			"-list_devices", "true", "-i", "").CombinedOutput()
		re := regexp.MustCompile(`\[(\d+)\] (.+)`)
		m := re.FindAllStringSubmatch(string(out), -1)
		var devs []string
		for _, v := range m {
			devs = append(devs, v[2])
		}
		return devs, nil
	}
	return nil, fmt.Errorf("unsupported OS")
}

// StopAllAudio makes sure *any* recording or playback process is killed.
func (gp *GoPronounce) StopAllAudio() {
	if gp.isRecording {
		gp.stopRecording()
	}
	if gp.isPlayingRef {
		gp.stopReferencePlayback()
	}
	if gp.isPlayingRec {
		gp.stopRecordingPlayback()
	}
}
