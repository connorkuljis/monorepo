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
	recordings := make([]KeyPress, 1000)
	fmt.Println("Press keys (press 'ENTER' to quit): ")

	err := keyboard.Open()
	if err != nil {
		fmt.Println("Error opening keyboard:", err)
		return
	}

	var prevTime time.Time
	for i := 0; i < 1000; {
		char, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Println("Error reading key:", err)
			return
		}

		if key == keyboard.KeyEnter {
			fmt.Println("Playing recording...")
			playbackRecording(recordings)
			break
		}

		if key == keyboard.KeyBackspace2 {
			if i == 0 {
				continue
			}

			i--
			fmt.Printf("\u001b[2K\r") // escape sequence clears term row
			recordings[i] = KeyPress{}
			for j := 0; j < i; j++ {
				fmt.Printf("%c", recordings[j].Character)
			}
			continue
		}

		if key == keyboard.KeySpace {
			char = ' '
		}

		// Calculate the time difference
		currentTime := time.Now()
		var kp KeyPress
		if !prevTime.IsZero() {
			fmt.Printf("%c", char)
			kp = NewKeyPress(char, currentTime.Sub(prevTime))
		} else {
			fmt.Printf("%c", char)
			kp = NewKeyPress(char, 0)
		}
		// Update the previous time
		prevTime = currentTime
		recordings[i] = kp
		i++
	}
	keyboard.Close()
}

func NewKeyPress(char rune, delay time.Duration) KeyPress {
	return KeyPress{Character: char, Delay: delay}
}

func playbackRecording(recording []KeyPress) {
	for _, e := range recording {
		time.Sleep(e.Delay)
		fmt.Printf("%c", e.Character)
	}
}
