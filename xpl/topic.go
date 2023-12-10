package xpl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/droso-hass/xpl2mqtt/cmd"
)

// xpl2mqtt/messagetype/device_type/device_id/device_parameter/action
// ex: xpl2mqtt/sensor.basic/th-1/0x12345678/temp/state
type Topic struct {
	MessageType string
	DeviceType  string
	DeviceID    string
	DeviceParam string
	Action      string
}

func (t *Topic) String() string {
	return fmt.Sprintf(
		"%s/%s/%s/%s/%s/%s",
		cmd.ConfigData.MqttBaseTopic,
		t.MessageType,
		t.DeviceType,
		t.DeviceID,
		t.DeviceParam,
		t.Action,
	)
}

func (t *Topic) StringO(o Topic) string {
	return fmt.Sprintf(
		"%s/%s/%s/%s/%s/%s",
		cmd.ConfigData.MqttBaseTopic,
		getStr(o.MessageType, t.MessageType),
		getStr(o.DeviceType, t.DeviceType),
		getStr(o.DeviceID, t.DeviceID),
		getStr(o.DeviceParam, t.DeviceParam),
		getStr(o.Action, t.Action),
	)
}

func (t *Topic) Parse(topic string) error {
	data := strings.Split(topic[len(cmd.ConfigData.MqttBaseTopic)+1:], "/")
	if len(data) != 5 {
		return errors.New("invalid topic")
	}
	t.MessageType = data[0]
	t.DeviceType = data[1]
	t.DeviceID = data[2]
	t.DeviceParam = data[3]
	t.Action = data[4]
	return nil
}

func getStr(a string, b string) string {
	if a != "" {
		return a
	}
	return b
}
