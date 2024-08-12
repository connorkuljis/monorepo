// FILE     : main.go
// AUTHOR   : conkuljis@gmail.com
// DATE     : August 2024
// PURPOSE  : Main program for facilitating a time clock command line utility

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// A card is created once when logging in, and recreated (from the last login card) when logging out.
// For each
// One card for the start time
// One card the end time.
// / sharing the same card id.
type card struct {
	id          int32
	timestamp   int64
	description string
	mode        mode
}

// Each 'card' is enumerated 0 or 1 to represent the 'mode' of login.
type mode int

const (
	Login mode = iota
	Logout
)

const output = "/tmp/logs.txt"

type timesheet []card

func main() {
	in := flag.Bool("in", false, "punch in")
	out := flag.Bool("out", false, "punch out")
	flag.Parse()

	timesheet, err := getTimesheet()
	if err != nil {
		log.Fatal(err)
	}

	var currentCard card

	if *in {
		// initialise currentCard to avoid if-else
		currentCard = NewCard()

		lastCard, ok := timesheet.Last()
		if ok {
			currentCard = lastCard
			if currentCard.mode == Login {
				log.Fatal(errors.New("Please punch out before punching in."))
			}
		}

		if len(flag.Args()) < 1 {
			log.Fatal(errors.New("Please provide a description."))
		}

		currentCard.description = flag.Args()[0]
		currentCard.login()
		currentCard.save() // remember to save any time we login or logout.
		fmt.Println("you are now logged in.")
	}

	if *out {
		lastCard, ok := timesheet.Last()
		if !ok {
			log.Fatal("You have no cards on record.")
		}

		if lastCard.mode == Logout {
			log.Fatal(errors.New("Please punch in before punching out."))
		}

		currentCard = lastCard

		currentCard.logout()
		currentCard.save()
		fmt.Println("you are now logged out.")
	}
}

func NewCard() card {
	return card{
		id: rand.Int31(),
	}
}

func (c *card) login() {
	c.mode = Login
	c.stamp()
}

func (c *card) logout() {
	c.mode = Logout
	c.stamp()
}

func (c *card) stamp() {
	c.timestamp = time.Now().UnixMilli()
}

func (c *card) save() {
	file, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintln(file, c.string())
}

func (c *card) string() string {
	var mode string

	if c.mode == Login {
		mode = "LOGIN"
	}

	if c.mode == Logout {
		mode = "LOGOUT"
	}

	return fmt.Sprintf("%d\t%d\t%s\t%s", c.id, c.timestamp, c.description, mode)
}

func (t *timesheet) IsEmpty() bool {
	return len(*t) == 0
}

func (t *timesheet) Last() (card, bool) {
	if t.IsEmpty() {
		return card{}, false
	}

	timesheet := *t
	return timesheet[len(timesheet)-1], true
}

func getTimesheet() (timesheet, error) {
	file, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var timesheet timesheet
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, "\t", 4)
		if len(fields) != 4 {
			log.Printf("Invalid line: %s\n", line)
			continue
		}

		id, err := strconv.Atoi(fields[0])
		if err != nil {
			log.Printf("Invalid id: %s\n", fields[0])
			continue
		}

		time, err := strconv.Atoi(fields[1])
		if err != nil {
			log.Printf("Invalid time: %s\n", fields[1])
			continue
		}

		desc := fields[2]

		var mode mode
		switch fields[3] {
		case "LOGIN":
			mode = Login
		case "LOGOUT":
			mode = Logout
		default:
			log.Printf("Invalid mode string: %s\n", fields[3])
			continue
		}

		card := card{
			id:          int32(id),
			timestamp:   int64(time),
			description: desc,
			mode:        mode,
		}

		timesheet = append(timesheet, card)
	}

	if err := scanner.Err(); err != nil {
		return timesheet, err
	}

	return timesheet, nil
}
