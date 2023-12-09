package utils

import (
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func MqttError(t mqtt.Token) error {
	_ = t.Wait()
	if t.Error() != nil {
		slog.Error("mqtt error: " + t.Error().Error())
	}
	return t.Error()
}
