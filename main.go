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

type mode int

const (
	Login mode = iota
	Logout
)

const output = "/tmp/logs.txt"

type card struct {
	id          int32
	timestamp   int64
	description string
	mode        mode
}

type cardStack []card

func main() {
	in := flag.Bool("in", false, "punch in")
	out := flag.Bool("out", false, "punch out")
	flag.Parse()

	cards, err := getCardStack()
	if err != nil {
		log.Fatal(err)
	}

	var currentCard card

	if *in {
		currentCard = NewCard()
		if len(cards) > 0 {
			lastCard := cards[len(cards)-1]
			if lastCard.mode == Login {
				log.Fatal(errors.New("Please punch out before punching in."))
			}
			currentCard = lastCard
		}

		if len(flag.Args()) < 1 {
			log.Fatal(errors.New("Please provide a description."))
		}

		currentCard.description = flag.Args()[0]

		currentCard.login()
		currentCard.save()
	}

	if *out {
		if len(cards) == 0 {
			log.Fatal("You have no cards on record.")
		}

		lastCard := cards[len(cards)-1]
		if lastCard.mode == Logout {
			log.Fatal(errors.New("Please punch in before punching out."))
		}

		currentCard = lastCard

		currentCard.logout()
		currentCard.save()
	}

	log.Println(currentCard.string())
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
	fmt.Println("Successfully saved card!\n", c.string())
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

func getCardStack() (cardStack, error) {
	file, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var stack cardStack
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

		stack = append(stack, card)
	}

	if err := scanner.Err(); err != nil {
		return stack, err
	}

	return stack, nil
}
