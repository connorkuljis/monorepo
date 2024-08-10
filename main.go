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

type loginMode string

const (
	Login  loginMode = "LOGIN"
	Logout loginMode = "LOGOUT"
)

type card struct {
	id          int32
	timestamp   int64
	description string
	loginMode   loginMode
}

func (c *card) String() string {
	return fmt.Sprintf("%d\t%d\t%s\t%s", c.id, c.timestamp, c.description, c.loginMode)
}

func main() {
	loginFlag := flag.Bool("in", false, "Login")
	logoutFlag := flag.Bool("out", false, "Logout")
	flag.Parse()

	log.Printf("login mode=%t", *loginFlag)
	log.Printf("logout mode=%t", *logoutFlag)

	if *loginFlag {
		msg := "hello, world"
		c, err := login(msg)
		if err != nil {
			log.Fatal(err)
		}
		save(c)
		fmt.Println("Logged in:", c.String())
	}

	if *logoutFlag {
		c, err := logout()
		if err != nil {
			log.Fatal(err)
		}
		save(c)
		fmt.Println("Logged out:", c.String())
	}

	fmt.Println("Exiting.")
}

func getLastCard() (card, error) {
	cards, err := loadCards()
	if err != nil {
		return card{}, err
	}

	if len(cards) == 0 {
		// return an empty card.
		// use the default value of loginMode to check if we should login/logout.
		return card{}, nil
	}

	return cards[len(cards)-1], nil
}

func login(desc string) (card, error) {
	lastCard, err := getLastCard()
	if err != nil {
		return card{}, err
	}

	// in the case of login, if the last card mode is LOGIN, an error is raised.
	if lastCard.loginMode == Login {
		return card{}, errors.New("Please logout before loggin in.")
	}

	time := time.Now().UnixMilli()
	id := rand.Int31()

	return card{
		id:          id,
		timestamp:   time,
		description: desc,
		loginMode:   Login,
	}, nil
}

func logout() (card, error) {
	lastCard, err := getLastCard()
	if err != nil {
		return card{}, err
	}

	// in the case of logout, loginMode must be LOGIN, or an error is raised.
	if lastCard.loginMode != Login {
		return card{}, errors.New("Please login before logging out.")
	}

	lastCard.loginMode = Logout
	lastCard.timestamp = time.Now().UnixMilli()

	return lastCard, nil
}

func save(card card) {
	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintln(file, card.String())
}

func loadCards() ([]card, error) {
	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

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

		var mode loginMode
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
			loginMode:   mode,
		}

		cards = append(cards, card)
	}

	if err := scanner.Err(); err != nil {
		return cards, err
	}

	return cards, nil
}
