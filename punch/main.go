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
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
)

// When logging in a new card is created.
// When logging out, a copy of the login card is used.
type card struct {
	id          uuid.UUID
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

// TODO: convert flags to args eg:
//
//	`punch -in <desc>` becomes `punch in <desc>`
func main() {
	in := flag.Bool("in", false, "punch in")
	out := flag.Bool("out", false, "punch out")
	flag.Parse()

	timesheet, err := getTimesheet()
	if err != nil {
		log.Fatal(err)
	}

	if *in {
		// guard to protect double login
		lastCard, ok := timesheet.last()
		if ok && lastCard.mode == Login {
			log.Fatal(errors.New("Please punch out before punching in."))
		}

		currentCard := NewCard()

		if len(flag.Args()) < 1 {
			log.Fatal(errors.New("Please provide a description."))
		}

		currentCard.description = flag.Args()[0]
		currentCard.login()
		fmt.Println(currentCard.String())

		return
	}

	if *out {
		if timesheet.isEmpty() {
			log.Fatal("You have no cards on record.")
		}

		// guard to protect double log out
		lastCard, ok := timesheet.last()
		if ok && lastCard.mode == Logout {
			log.Fatal(errors.New("Please punch in before punching out."))
		}

		currentCard := lastCard
		currentCard.logout()
		fmt.Println(currentCard.String())

		return
	}

	// TODO: exec this as `-report` flag
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.TabIndent)
	pairs := timesheet.GetPairs()
	for _, pair := range pairs {
		if len(pair) == 2 {
			duration := diff(pair[0], pair[1])
			fmt.Fprintf(w, "%s\t%s\t%s\n", pair[0].id.String(), duration, pair[0].description)
		}
	}
	w.Flush()
}

func NewCard() card {
	return card{
		id: uuid.New(),
	}
}

func (c *card) login() {
	c.mode = Login
	c.stamp()
	c.save()
}

func (c *card) logout() {
	c.mode = Logout
	c.stamp()
	c.save()
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

	fmt.Fprintln(file, c.String())
}

func (c *card) String() string {
	var mode string

	if c.mode == Login {
		mode = "LOGIN"
	}

	if c.mode == Logout {
		mode = "LOGOUT"
	}

	return fmt.Sprintf("%s\t%d\t%s\t%s", c.id.String(), c.timestamp, c.description, mode)
}

func (t *timesheet) isEmpty() bool {
	return len(*t) == 0
}

func (t *timesheet) last() (card, bool) {
	if t.isEmpty() {
		return card{}, false
	}

	timesheet := *t
	return timesheet[len(timesheet)-1], true
}

func (t *timesheet) Print() {
	for i, c := range *t {
		fmt.Println(i, c.String())
	}
}

func (t *timesheet) GetPairs() [][]card {
	timesheet := *t
	var pairs [][]card
	for i := 0; i < len(timesheet); i += 2 {
		pair := []card{timesheet[i]}
		hasOneMore := i+1 < len(timesheet)
		if hasOneMore {
			pair = append(pair, timesheet[i+1])
			diff(timesheet[i], timesheet[i+1])
		}
		pairs = append(pairs, pair)
	}
	return pairs
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

		id, err := uuid.Parse(fields[0])
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
			id:          id,
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

func diff(c1, c2 card) time.Duration {
	t1 := time.UnixMilli(c1.timestamp)
	t2 := time.UnixMilli(c2.timestamp)
	return t2.Sub(t1)
}
