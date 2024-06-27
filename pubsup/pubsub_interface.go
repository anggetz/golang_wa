package pubsup

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
)

type Whatsapp interface {
	GetQRCOde() ([]byte, error)
	SetBase64QrCode(string)
	GetDevices() ([]*store.Device, error)
	SetAktifSender(jid string)
	Send(jid string, message string) (whatsmeow.SendResponse, error)
	GetStoreDevice() *store.Device
	GetClient() *whatsmeow.Client
	SetClient(*whatsmeow.Client)
	IsLoggedIn() bool
	RequestQRCode(chan string)
	GetPairCode(number string) string
}

type WaSend struct {
	Jid     string
	Message string
}

type PairCode struct {
	Number string
}
