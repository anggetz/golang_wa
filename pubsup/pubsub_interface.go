package pubsup

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
)

type Whatsapp interface {
	Login() ([]byte, error)
	SetBase64QrCode(string)
	GetDevices() ([]*store.Device, error)
	SetAktifSender(jid string)
	Send(jid string, message string) (whatsmeow.SendResponse, error)
	GetStoreDevice() *store.Device
	GetClient() *whatsmeow.Client
	SetClient(*whatsmeow.Client)
	IsLoggedIn() bool
}

type WaSend struct {
	Jid     string
	Message string
}
