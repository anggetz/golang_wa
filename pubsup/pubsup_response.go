package pubsup

type IsLoggedInResponse struct {
	IsLoggedIn bool `json:"is_logged_in"`
}

type RequestQRCodeResponse struct {
	Message string `json:"message"`
}

type PairCodeResponse struct {
	PairCode string `json:"pair_code"`
}
