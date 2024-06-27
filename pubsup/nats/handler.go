package nats

import (
	"encoding/json"
	"fmt"

	"github.com/anggetz/golangwa/kernel"
	"github.com/anggetz/golangwa/pubsup"

	"github.com/nats-io/nats.go"
)

func RegisterHandler(n *nats.Conn, implementor pubsup.Whatsapp) {
	n.Subscribe(kernel.Kernel.AppName+".login", func(msg *nats.Msg) {

		base64string, _ := implementor.GetQRCOde()

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

	n.Subscribe(kernel.Kernel.AppName+".get-pair-code", func(msg *nats.Msg) {
		payload := pubsup.PairCode{}

		err := json.Unmarshal(msg.Data, &payload)
		if err != nil {
			fmt.Println("error unmarshall payload: " + err.Error())
			return
		}

		pairCode := implementor.GetPairCode(payload.Number)

		resp := pubsup.PairCodeResponse{
			PairCode: pairCode,
		}

		byResp, err := json.Marshal(resp)
		if err != nil {
			fmt.Println("error marshal: " + err.Error())
			return
		}

		msg.Respond(byResp)
	})

}
