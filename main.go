package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anggetz/golangwa/kernel"
	"github.com/anggetz/golangwa/pubsup"

	naNwa "github.com/anggetz/golangwa/pubsup/nats"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
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
		"simada_whatsapp",
	), dbLog)

	defer container.Close()

	// register nats
	logicImpl := new(pubsup.PubSupLogic)

	logicImpl.ContainerSqlStore = container

	if err != nil {
		panic(err)
	}
	// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	clientLogger := waLog.Stdout("Client", "DEBUG", true)

	logicImpl.Logger = &clientLogger
	logicImpl.CurrentDevice = deviceStore

	// NewClientWA(logicImpl, &clientLogger)
	client := whatsmeow.NewClient(logicImpl.CurrentDevice, *&clientLogger)

	logicImpl.SetClient(client)

	naNwa.RegisterHandler(nc, logicImpl)

	// go func() {
	// 	for {
	// 		if kernel.Kernel.Client == nil || !kernel.Kernel.Client.IsConnected() {
	// 			fmt.Println("restarting....")
	// 			NewClientWA(kernel.Kernel.Client, kernel.Kernel.CurrentDevice, kernel.Kernel.Logger)
	// 		}
	// 	}
	// }()
	// if logicImpl.Client.Store.ID != nil {
	// 	// No ID stored, new login
	// 	err = logicImpl.Client.Connect()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	log.Println("logged in")
	// }
	logicImpl.Client.Connect()

	defer logicImpl.Client.Disconnect()

	for {
		if logicImpl.Client.IsConnected() {
			logicImpl.Client.Connect()
		}

		fmt.Println("keep alive manual")
		time.Sleep(2 * time.Second)
	}

}
