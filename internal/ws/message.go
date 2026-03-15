package ws

type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ServerMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
