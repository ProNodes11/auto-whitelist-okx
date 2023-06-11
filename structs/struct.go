package structs

type RequestData struct {
	Url         string
  AuthHeader  string
  DevIdHeader string
  Payload     []byte
}

type AddressInfo struct {
	Address      string `json:"address"`
	ValidateName string `json:"validateName"`
}
