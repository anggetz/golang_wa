package nats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anggetz/golangwa/kernel"
	"github.com/anggetz/golangwa/pubsup"

	"github.com/nats-io/nats.go"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func RegisterHandler(n *nats.Conn, implementor pubsup.Whatsapp) {
	n.Subscribe(kernel.Kernel.AppName+".login", func(msg *nats.Msg) {

		base64string, _ := implementor.Login()

		msg.Respond(base64string)
	})

	n.Subscribe(kernel.Kernel.AppName+".send", func(msg *nats.Msg) {
		payload := pubsup.WaSend{}

		err := json.Unmarshal(msg.Data, &payload)
		if err != nil {
			fmt.Println("error unmarshall payload: " + err.Error())
			return
		}

		_, err = implementor.Send(payload.Jid, payload.Message)
		if err != nil {
			fmt.Println("error send: " + err.Error())
			return
		}

		byResp, _ := json.Marshal(struct {
			Ok bool
		}{
			Ok: true,
		})

		msg.Respond(byResp)
	})

	n.Subscribe(kernel.Kernel.AppName+".devices", func(msg *nats.Msg) {
		resp, err := implementor.GetDevices()

		if err != nil {
			fmt.Println("error send: " + err.Error())
			return
		}

		byResp, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("error marshal: " + err.Error())
			return
		}

		msg.Respond(byResp)
	})

	n.Subscribe(kernel.Kernel.AppName+".check-login", func(msg *nats.Msg) {
		ok := implementor.IsLoggedIn()

		resp := pubsup.IsLoggedInResponse{
			IsLoggedIn: ok,
		}

		byResp, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("error marshal: " + err.Error())
			return
		}

		msg.Respond(byResp)
	})
}

type NatsWa struct {
	Base64QrCode      string
	ContainerSqlStore *sqlstore.Container
	Logger            *waLog.Logger
	Client            *whatsmeow.Client
	CurrentDevice     *store.Device
}

// get qr base64
func (nwa *NatsWa) Login() ([]byte, error) {
	return json.Marshal(struct {
		Base64QR string
	}{
		Base64QR: nwa.Base64QrCode,
	})
}

// send message
func (nwa *NatsWa) Send(jid string, message string) (whatsmeow.SendResponse, error) {
	targetJid, _ := types.ParseJID(jid)

	return nwa.Client.SendMessage(context.Background(), targetJid, &waE2E.Message{
		Conversation: &message,
	})
}

func (nwa *NatsWa) GetDevices() ([]*store.Device, error) {

	return nwa.ContainerSqlStore.GetAllDevices()
}

func (nwa *NatsWa) SetBase64QrCode(base64QrCode string) {

	nwa.Base64QrCode = base64QrCode
}

func (nwa *NatsWa) GetStoreDevice() *store.Device {

	return nwa.CurrentDevice
}

func (nwa *NatsWa) GetClient() *whatsmeow.Client {

	return nwa.Client
}

func (nwa *NatsWa) SetClient(c *whatsmeow.Client) {

	nwa.Client = c
}

func (nwa *NatsWa) IsLoggedIn() bool {

	return nwa.Client.IsConnected()
}

func (nwa *NatsWa) SetAktifSender(jid string) {
	targetJid, _ := types.ParseJID(jid)

	device, _ := nwa.ContainerSqlStore.GetDevice(targetJid)

	nwa.Client = whatsmeow.NewClient(device, *nwa.Logger)
}
