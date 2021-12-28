package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	answers  = make(chan map[string]float64)
	answer   float64
	task     string
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()
	go check()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()

	ch <- "Task: " + task
	messages <- who + " has arrived"
	entering <- ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		messages <- who + ": " + input.Text()
		ans, err := strconv.ParseFloat(input.Text(), 32)
		if err == nil {
			answers <- map[string]float64{who: math.Floor(ans*100) / 100}
		}

	}
	leaving <- ch
	messages <- who + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func check() {
	random()
	for {
		select {
		case ans := <-answers:
			for k, v := range ans {
				fmt.Println(v, answer, v == answer)
				if v == answer {
					messages <- k + " has won!"
					random()
					break
				}
			}
		}
	}
}

func randomNumber(o int, n int) float64 {
	rand.Seed(time.Now().Unix())
	rangeLower := o
	rangeUpper := n
	randomNum := rangeLower + rand.Intn(rangeUpper-rangeLower+1)
	return float64(randomNum)
}

func random() {

	a := randomNumber(0, 5)
	b := randomNumber(1, 5)
	op := randomNumber(0, 3)

	switch op {
	case 0:
		answer = a + b
		t := fmt.Sprintf("%.0f + %.0f", a, b)
		messages <- t
		task = t
	case 1:
		answer = a - b
		t := fmt.Sprintf("%.0f - %.0f", a, b)
		messages <- t
		task = t
	case 2:
		answer = a * b
		t := fmt.Sprintf("%.0f * %.0f", a, b)
		messages <- t
		task = t
	case 3:
		answer = math.Floor(a/b*100) / 100
		t := fmt.Sprintf("%.0f / %.0f answer is rounded to 2 decimal places", a, b)
		messages <- t
		task = t
	default:
		log.Fatal("Wrong operator")
	}
	log.Println(answer)
}
