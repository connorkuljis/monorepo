// NOTE: this code has been tested on macOS. It may or may not work on other platforms.

package main

import (
	"fmt"
	"time"

	"github.com/eiannone/keyboard"
)

type KeyPress struct {
	Character rune
	Delay     time.Duration
}

func main() {
	kkyy()
}

func kkyy() {
	sequence := make([]KeyPress, 1000)
	fmt.Println("Recording character sequence. Please input keys (press 'ENTER' " +
		"to begin playback): ")

	err := keyboard.Open()
	if err != nil {
		fmt.Println("Error opening keyboard:", err)
		return
	}

	var prevTime time.Time
	i := 0
	for i < 1000 {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("Error reading key:", err)
			return
		}

		if key == keyboard.KeyEnter {
			fmt.Print("\nPlaying recording: ")
			playbackRecording(sequence, i)
			break
		}

		if key == keyboard.KeyBackspace2 || key == keyboard.KeyBackspace {
			if i == 0 {
				continue
			}

			i = i - 1 // moves cursor back one position

			clearTerminalLine()
			for j := 0; j < i; j++ {
				fmt.Printf("%c", sequence[j].Character)
			}

			continue
		}

		if key == keyboard.KeySpace {
			char = ' '
		}

		fmt.Printf("%c", char)

		currentTime := time.Now()
		kp := NewKeyPress(char, currentTime, prevTime)
		sequence[i] = kp
		prevTime = currentTime
		i++
	}
	keyboard.Close()

	debugRecording(sequence, i)
	calculateWPM(sequence, i)
}

func NewKeyPress(char rune, currentTime, prevTime time.Time) KeyPress {
	var delay time.Duration
	if !prevTime.IsZero() {
		delay = currentTime.Sub(prevTime)
	}
	return KeyPress{Character: char, Delay: delay}
}

func playbackRecording(recording []KeyPress, n int) {
	for i := 0; i < n; i++ {
		time.Sleep(recording[i].Delay)
		fmt.Printf("%c", recording[i].Character)
	}
	fmt.Printf("\n")
}

func clearTerminalLine() {
	// escape code to erase current line and move cursor to beginning.
	fmt.Printf("\u001b[2K\r")
}

func debugRecording(recording []KeyPress, n int) {
	time.Sleep(500 * time.Millisecond)
	fmt.Println("\nDebug information:")
	for i := 0; i < n; i++ {
		fmt.Printf("%d - %c - %s\n", i, recording[i].Character, recording[i].Delay.String())
	}
}

func calculateWPM(recording []KeyPress, n int) {
	var totalTime time.Duration
	for i := 0; i < n; i++ {
		totalTime = totalTime + recording[i].Delay
	}

	wordCount := float32(n) / 5

	wpm := wordCount / float32(totalTime.Minutes())

	fmt.Println("word count: ", wordCount)
	fmt.Println("total time: " + totalTime.String())
	fmt.Printf("WPM: %.f\n", wpm)
}
