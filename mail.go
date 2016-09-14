package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"os/signal"
)

var addr = flag.String("addr", "localhost:8081", "http service address")

func main() {
	to := flag.String("to", "", "destination Internet mail address")
	from := flag.String("from", "", "sendors Internet mail address")
	pwd := flag.String("password", "", "sendors Internet mail password")
	flag.Usage = func() {
		fmt.Printf("Syntax:\n\tgsend [flags]\nwhere flags are:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/alarms"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	for {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			body := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n\r\n%s", *to, "Alarm", message, *from)
			auth := smtp.PlainAuth("", *from, *pwd, "smtp.gmail.com")
			err = smtp.SendMail("smtp.gmail.com:587", auth, *from,
				[]string{*to}, []byte(body))
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("recv: %s", message)
		}
	}
}
