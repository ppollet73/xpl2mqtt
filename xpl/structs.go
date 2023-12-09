package xpl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/droso-hass/xpl2mqtt/cmd"
)

type HAConfig struct {
	Name                   string   `json:"name,omitempty"`
	UniqueID               string   `json:"unique_id,omitempty"`
	DeviceClass            string   `json:"device_class,omitempty"`
	StateTopic             string   `json:"state_topic,omitempty"`
	Unit                   string   `json:"unit_of_measurement,omitempty"`
	CommandTopic           string   `json:"command_topic,omitempty"`
	BrightnessScale        int      `json:"brightness_scale,omitempty"`
	BrightnessStateTopic   string   `json:"brightness_state_topic,omitempty"`
	BrightnessCommandTopic string   `json:"brightness_command_topic,omitempty"`
	Icon                   string   `json:"icon,omitempty"`
	Device                 HADevice `json:"device,omitempty"`
	SupportedFeatures      []string `json:"supported_features,omitempty"`
	CodeArmRequired        bool     `json:"code_arm_required,omitempty"`
	CodeDisarmRequired     bool     `json:"code_disarm_required,omitempty"`
	CodeTriggerRequired    bool     `json:"code_trigger_required,omitempty"`
}

type HADevice struct {
	Identifiers  []string `json:"identifiers,omitempty"`
	Manifacturer string   `json:"manufacturer"`
	Name         string   `json:"name"`
	Model        string   `json:"model"`
}

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
