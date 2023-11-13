package streaming

import (
	"L0/pkg/mynats"
)

func PublishNats(chanelName string, jsonText string) error {
	js := *mynats.Js
	_, err := js.Publish(chanelName, []byte(jsonText))

	return err
}
