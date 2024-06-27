package pubsup

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	qrCode "github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type PubSupLogic struct {
	Base64QrCode      string
	ContainerSqlStore *sqlstore.Container
	Logger            *waLog.Logger
	Client            *whatsmeow.Client
	CurrentDevice     *store.Device
	IsInRequestLogin  bool
}

// get qr base64
func (nwa *PubSupLogic) GetQRCOde() ([]byte, error) {
	if !nwa.IsInRequestLogin {
		chanQrGenerated := make(chan string)
		go nwa.RequestQRCode(chanQrGenerated)

		nwa.Base64QrCode = <-chanQrGenerated
	}

	fmt.Println(nwa.Base64QrCode)

	return json.Marshal(struct {
		Base64QR string
	}{
		Base64QR: nwa.Base64QrCode,
	})
}

// send message
func (nwa *PubSupLogic) Send(jid string, message string) (whatsmeow.SendResponse, error) {
	targetJid, _ := types.ParseJID(jid)

	return nwa.Client.SendMessage(context.Background(), targetJid, &waE2E.Message{
		Conversation: &message,
	})
}

func (nwa *PubSupLogic) GetDevices() ([]*store.Device, error) {

	return nwa.ContainerSqlStore.GetAllDevices()
}

func (nwa *PubSupLogic) SetBase64QrCode(base64QrCode string) {

	nwa.Base64QrCode = base64QrCode
}

func (nwa *PubSupLogic) GetStoreDevice() *store.Device {

	return nwa.CurrentDevice
}

func (nwa *PubSupLogic) GetClient() *whatsmeow.Client {

	return nwa.Client
}

func (nwa *PubSupLogic) SetClient(c *whatsmeow.Client) {

	nwa.Client = c
}

func (nwa *PubSupLogic) IsLoggedIn() bool {

	return nwa.Client.IsLoggedIn()
}

func (mwa *PubSupLogic) GetPairCode(number string) string {
	s, err := mwa.Client.PairPhone("62"+number, true, whatsmeow.PairClientChrome, "Chrome (Windows)")
	if err != nil {
		panic(err)
	}

	return s
}

func (nwa *PubSupLogic) SetAktifSender(jid string) {
	targetJid, _ := types.ParseJID(jid)

	device, _ := nwa.ContainerSqlStore.GetDevice(targetJid)

	nwa.Client = whatsmeow.NewClient(device, *nwa.Logger)
}

func (nwa *PubSupLogic) RequestQRCode(flagQrGenerated chan string) {
	var err error

	nwa.IsInRequestLogin = true

	defer func() {
		nwa.IsInRequestLogin = false
		nwa.SetBase64QrCode("")
	}()

	qrChan, _ := nwa.Client.GetQRChannel(context.Background())
	err = nwa.Client.Connect()
	if err != nil {
		log.Println("ERROR", err.Error())
		return
	}
	for evt := range qrChan {
		if evt.Event == "code" {
			// Render the QR code here
			// e.g. qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			// or just manually `echo 2@... | qrencode -t ansiutf8` in a terminal
			qrPNG, _ := qrCode.Encode(evt.Code, qrCode.Highest, 256)

			Base64QrCode := base64.StdEncoding.EncodeToString(qrPNG)

			flagQrGenerated <- Base64QrCode
		} else {

			fmt.Println("Login event:", evt.Event)
		}
	}

}
