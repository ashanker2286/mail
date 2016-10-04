package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/smtp"
	"net/url"
	"os"
	"os/signal"
)

type AlarmState struct {
	OwnerId          int32  `SNAPROUTE: "KEY", ACCESS:"r", MULTIPLICITY:"*", DESCRIPTION: "Alarm owner daemon Id picked up from events.json"`
	EventId          int32  `SNAPROUTE: "KEY", ACCESS:"r", MULTIPLICITY:"*", DESCRIPTION: "Alarm event id picked up from events.json"`
	OwnerName        string `SNAPROUTE: "KEY", ACCESS:"r", MULTIPLICITY:"*", DESCRIPTION: "Alarm owner daemon name picked up from events.json"`
	EventName        string `SNAPROUTE: "KEY", ACCESS:"r", MULTIPLICITY:"*", DESCRIPTION: "Alarm event name picked up from events.json"`
	SrcObjName       string `SNAPROUTE: "KEY", ACCESS:"r", MULTIPLICITY:"*", DESCRIPTION: "Alarm event name picked up from events.json"`
	Severity         string `DESCRIPTION: "Alarm Severity"`
	Description      string `DESCRIPTION: "Description explaining the fault"`
	OccuranceTime    string `DESCRIPTION: "Timestamp at which fault occured"`
	SrcObjKey        string `DESCRIPTION: "Fault Object Key"`
	SrcObjUUID       string `DESCRIPTION: "Fault Object UUID"`
	ResolutionTime   string `DESCRIPTION: "Resolution Time stamp"`
	ResolutionReason string `DESCRIPTION: "Cleared/Disabled"`
}

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
			var alarmState AlarmState
			_ = json.Unmarshal(message, &alarmState)
			alarmMsg := "Event Name: " + alarmState.EventName + "\r\n" + "Object: " + alarmState.SrcObjName + "\r\n " + alarmState.SrcObjKey + "\r\n" + "Occurance Time: " + alarmState.OccuranceTime + "\r\nResolution Time: " + alarmState.ResolutionTime + "\r\n" + "Resolution Reason: " + alarmState.ResolutionReason
			subject := "Alarm from " + *addr
			body := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n\r\n%s", *to, subject, alarmMsg, "support@snaproute.com")
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
