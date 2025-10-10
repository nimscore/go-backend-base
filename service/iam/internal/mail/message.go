package schema

import "encoding/json"

const KIND_MAIL_CONFIRM = "mail_confirm"
const KIND_MAIL_RECOVER = "mail_recover"

type Message struct {
	From      string            `json:"from"`
	To        string            `json:"to"`
	Subject   string            `json:"subject"`
	Arguments map[string]string `json:"arguments"`
}

func NewMessageMailConfirm(from string, to string, subject string, user string, time string, url string) *Message {
	return &Message{
		From:    from,
		To:      to,
		Subject: subject,
		Arguments: map[string]string{
			"user": user,
			"time": time,
			"url":  url,
		},
	}
}

func NewMessageMailRecover(from string, to string, subject string, user string, time string, url string) *Message {
	return &Message{
		From:    from,
		To:      to,
		Subject: subject,
		Arguments: map[string]string{
			"user": user,
			"time": time,
			"url":  url,
		},
	}
}

func MessageFromJson(content string) (*Message, error) {
	var result Message
	err := json.Unmarshal([]byte(content), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func MessageToJson(message *Message) (string, error) {
	content, err := json.Marshal(message)
	return string(content), err
}
