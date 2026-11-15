package protocol

const PingPath = "/ping"

type PingResponse struct {
	// Message is pong when the server is reachable and serving the ping endpoint.
	Message string `json:"message"`
}
