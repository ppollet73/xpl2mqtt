package xpl

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/droso-hass/xpl2mqtt/cmd"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var HADiscovery = make(map[string][]string)

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
	Manufacturer string   `json:"manufacturer"`
	Name         string   `json:"name"`
	Model        string   `json:"model"`
}

func sendHassPacket(client *mqtt.Client, deviceType string, deviceId string, data HAConfig) {
	if !cmd.ConfigData.HassDiscovery {
		return
	}
	data.Device.Manufacturer = "xpl2mqtt"
	sdata, err := json.Marshal(data)
	if err != nil {
		return
	}
	t := fmt.Sprintf("homeassistant/%s/%s/config", deviceType, deviceId)
	p := string(sdata)

	if _, ok := HADiscovery[t]; !ok || (ok && !slices.Contains(HADiscovery[t], p)) {
		HADiscovery[t] = append(HADiscovery[t], p)
		(*client).Publish(t, 1, true, p)
	}
}

func ProcessMqttDiscovery(c mqtt.Client, m mqtt.Message) {
	t := m.Topic()
	p := string(m.Payload())
	if _, ok := HADiscovery[t]; !ok || (ok && !slices.Contains(HADiscovery[t], p)) {
		HADiscovery[t] = append(HADiscovery[t], p)
	}
}
