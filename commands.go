package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type Args struct {
	Name     string
	Duration float64
}

type Remote struct {
	wg     *sync.WaitGroup
	Pause  chan bool
	Cancel chan bool
	Finish chan bool
}

var rootCmd = &cobra.Command{
	Use:   "block",
	Short: "Block removes distractions when you work on tasks.",
	Long: `
Block saves you time by blocking websites at IP level.
Progress bar is displayed directly in the terminal. 
Automatically unblock sites when the task is complete.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

var startCmd = &cobra.Command{
	Use: "start",
	Run: func(cmd *cobra.Command, args []string) {
		myArgs, err := parseArgs(args)
		if err != nil {
			cmd.Usage()
			log.Fatal(err)
		}

		currentTask = InsertTask(NewTask(myArgs.Name, myArgs.Duration))
		createdAt := time.Now()

		var b Blocker
		useBlocker := !flags.DisableBlocker
		if useBlocker {
			b = NewBlocker()
			if err := b.Block(); err != nil {
				log.Fatal(err)
			}
			if err = ResetDNS(); err != nil {
				log.Fatal(err)
			}
		}

		color.Red("ESC or 'q' to exit. Press any key to pause.")

		r := Remote{
			wg:     &sync.WaitGroup{},
			Pause:  make(chan bool, 1),
			Cancel: make(chan bool, 1),
			Finish: make(chan bool, 1),
		}

		r.wg.Add(2)
		go RenderProgressBar(r)
		go PollInput(r)

		if flags.ScreenRecorder {
			r.wg.Add(1)
			go FfmpegCaptureScreen(r)
		}

		r.wg.Wait()

		if useBlocker {
			if err = b.Unblock(); err != nil {
				log.Fatal(err)
			}
		}

		finishedAt := time.Now()
		actualDuration := finishedAt.Sub(createdAt)

		if err = UpdateFinishTimeAndDuration(currentTask, finishedAt, actualDuration); err != nil {
			log.Fatal(err)
		}

		if flags.Verbose {
			fmt.Printf("Start time:\t%s\n", createdAt.Format("3:04:05pm"))
			fmt.Printf("End time:\t%s\n", finishedAt.Format("3:04:05pm"))
			fmt.Printf("Duration:\t%d hours, %d minutes and %d seconds.\n", int(actualDuration.Hours()), int(actualDuration.Minutes())%60, int(actualDuration.Seconds())%60)
		}

		fmt.Println("Goodbye.")
	},
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show task history.",
	Run: func(cmd *cobra.Command, args []string) {
		RenderHistory()
	},
}

var deleteTaskCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a task by given ID.",
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		err := DeleteTaskByID(id)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var resetDNSCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset DNS cache.",
	Run: func(cmd *cobra.Command, args []string) {
		flags.Verbose = true
		err := ResetDNS()
		if err != nil {
			log.Println(err)
		}
	},
}

func parseArgs(args []string) (Args, error) {
	var myArgs = Args{}

	if len(args) != 2 {
		return myArgs, fmt.Errorf("Invalid number of arguments, expected 2, recieved: %d", len(args))
	}

	inDuration := args[0]
	inName := args[1]

	duration, err := strconv.ParseFloat(inDuration, 64)
	if err != nil {
		return myArgs, fmt.Errorf("Error converting %s to float. Please provide a valid float.", inDuration)
	}

	myArgs.Duration = duration
	myArgs.Name = inName

	return myArgs, nil
}
