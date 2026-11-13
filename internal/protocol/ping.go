package protocol

const PingPath = "/ping"

type PingResponse struct {
	Message string `json:"message"`
}
