package bot

import (
	"encoding/gob"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)


type waHandler struct {
	c *whatsapp.Conn
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		fmt.Println("Connection failed, underlying error: %v", e.Err)
		fmt.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		fmt.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			fmt.Println("Restore failed: %v", err)
		}
	} else {
		fmt.Println("error occoured: %v\n", err)
	}
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (waHandler *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.ContextInfo.QuotedMessageID, message.Text)

	if message.Text == "professor info"{
		msg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: message.Info.RemoteJid,
			},
			Text:        "TODO send professor info",
		}

	    msgId, err := waHandler.c.Send(msg)
	    if err != nil {
		  fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		  os.Exit(1)
	    } else {
		  fmt.Println("Message Sent -> ID : " + msgId)
	    }
	}

	if message.Text == "classrep info"{
		msg := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: message.Info.RemoteJid,
			},
			Text:        "TODO send class rep info",
		}

	    msgId, err := waHandler.c.Send(msg)
	    if err != nil {
		  fmt.Fprintf(os.Stderr, "error sending message: %v", err)
		  os.Exit(1)
	    } else {
		  fmt.Println("Message Sent -> ID : " + msgId)
	    }
	}
}


func Bot() {
	//create new WhatsApp connection
	wac, err := whatsapp.NewConn(5 * time.Second)
	if err != nil {
		fmt.Println("error creating connection: %v\n", err)
	}

	wac.SetClientVersion(2, 2021, 4)

	//Add handler
	wac.AddHandler(&waHandler{wac})

	//login or restore
	if err := login(wac); err != nil {
		fmt.Println("error logging in: %v\n", err)
	}

	//verifies phone connectivity
	pong, err := wac.AdminTest()

	if !pong || err != nil {
		fmt.Println("error pinging in: %v\n", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	//Disconnect safe
	fmt.Println("Shutting down now.")
	session, err := wac.Disconnect()
	if err != nil {
		fmt.Println("error disconnecting: %v\n", err)
	}
	if err := writeSession(session); err != nil {
		fmt.Println("error saving session: %v", err)
	}
}

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v\n", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v\n", err)
		}
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v\n", err)
	}
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}
