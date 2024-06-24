package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anggetz/golangwa/kernel"
	"github.com/anggetz/golangwa/pubsup"

	naNwa "github.com/anggetz/golangwa/pubsup/nats"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
	qrCode "github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Println("Received a message!", v.Message.GetConversation())
	}
}

func NewClientWA(implementor pubsup.Whatsapp, clientLog *waLog.Logger) {
	var err error

	client := whatsmeow.NewClient(implementor.GetStoreDevice(), *clientLog)
	client.AddEventHandler(eventHandler)

	implementor.SetClient(client)

	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
				qrPNG, _ := qrCode.Encode(evt.Code, qrCode.Medium, 256)

				Base64QrCode := base64.StdEncoding.EncodeToString(qrPNG)

				implementor.SetBase64QrCode(Base64QrCode)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)

	kernel.Kernel = kernel.NewKernel("simada_wa")
	// Make sure you add appropriate DB connector imports, e.g. github.com/mattn/go-sqlite3 for SQLite

	nc, _ := nats.Connect(fmt.Sprintf("%s:%s", os.Getenv("NATS_HOST"), os.Getenv("NATS_PORT")))
	nc.Subscribe("config.share", func(msg *nats.Msg) {
		err := json.Unmarshal(msg.Data, &kernel.Kernel.Config)
		if err != nil {
			panic(err)
		}

		log.Println("new config receive", kernel.Kernel.Config)
	})

	msg, err := nc.Request("config.get", []byte(""), time.Second*10)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(msg.Data, &kernel.Kernel.Config)
	if err != nil {
		panic(err)
	}

	log.Println("config receive", kernel.Kernel.Config)

	container, err := sqlstore.New("postgres", fmt.Sprintf("postgresql://%s:%s@%s:%s/%s",
		kernel.Kernel.Config.DB.User,
		kernel.Kernel.Config.DB.Password,
		kernel.Kernel.Config.DB.Host,
		kernel.Kernel.Config.DB.Port,
		kernel.Kernel.Config.DB.Database,
	), dbLog)

	defer container.Close()

	// register nats
	natsWaImpl := new(naNwa.NatsWa)

	natsWaImpl.ContainerSqlStore = container

	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	clientLogger := waLog.Stdout("Client", "DEBUG", true)

	natsWaImpl.Logger = &clientLogger
	natsWaImpl.CurrentDevice = deviceStore

	NewClientWA(natsWaImpl, &clientLogger)

	naNwa.RegisterHandler(nc, natsWaImpl)

	// go func() {
	// 	for {
	// 		if kernel.Kernel.Client == nil || !kernel.Kernel.Client.IsConnected() {
	// 			fmt.Println("restarting....")
	// 			NewClientWA(kernel.Kernel.Client, kernel.Kernel.CurrentDevice, kernel.Kernel.Logger)
	// 		}
	// 	}
	// }()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	natsWaImpl.Client.Disconnect()
}
