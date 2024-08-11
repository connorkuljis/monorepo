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

type mode string

const (
	Login  mode = "LOGIN"
	Logout mode = "LOGOUT"
	output      = "/tmp/logs.txt"
)

type card struct {
	id          int32
	timestamp   int64
	description string
	mode        mode
}

func (c *card) String() string {
	return fmt.Sprintf("%d\t%d\t%s\t%s", c.id, c.timestamp, c.description, c.mode)
}

func NewCard(description string) card {
	return card{
		id:          rand.Int31(),
		description: description,
	}
}

func (c *card) login() error {
	c.mode = Login
	c.stamp()
	return nil
}

func (c *card) logout() error {
	c.mode = Logout
	c.stamp()
	return nil
}

func (c *card) stamp() {
	c.timestamp = time.Now().UnixMilli()
}

func (c *card) save(file *os.File) {
	fmt.Fprintln(file, c.String())
	fmt.Println("Successfully saved card!:", c.String())
}

func main() {
	in := flag.Bool("in", false, "punch in")
	out := flag.Bool("out", false, "punch out")
	flag.Parse()

	file, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var previousCard card
	previousCard, err = getLastCard(file)
	if err != nil {
		log.Fatal(err)
	}

	var currentCard card
	if *in {
		if previousCard.mode == Login {
			log.Fatal(errors.New("Please punch out before punching in."))
		}

		if len(flag.Args()) < 1 {
			log.Fatal(errors.New("Please provide a description."))
		}

		currentCard = NewCard(flag.Args()[0])

		err := currentCard.login()
		if err != nil {
			log.Fatal(err)
		}
		currentCard.save(file)
	}

	if *out {
		if previousCard.mode != Login {
			log.Fatal(errors.New("Please punch in before punching out."))
		}

		currentCard = previousCard

		err := currentCard.logout()
		if err != nil {
			log.Fatal(err)
		}
		currentCard.save(file)
	}

	if currentCard.mode != "" {
		log.Println("Current mode:", currentCard.mode)
	}
}

func getLastCard(file *os.File) (card, error) {
	cards, err := loadAllCards(file)
	if err != nil {
		return card{}, err
	}

	if len(cards) == 0 {
		return card{}, nil
	}

	return cards[len(cards)-1], nil
}

func loadAllCards(file *os.File) ([]card, error) {
	scanner := bufio.NewScanner(file)
	var cards []card

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
		case string(Login):
			mode = Login
		case string(Logout):
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

		cards = append(cards, card)
	}

	if err := scanner.Err(); err != nil {
		return cards, err
	}

	return cards, nil
}
