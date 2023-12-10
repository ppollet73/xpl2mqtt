package xpl

import (
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/droso-hass/xpl2mqtt/utils"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var decoders = map[string](func(pkt *XPLPacket, mqtt *mqtt.Client)){
	"log.basic":    decodeLogs,
	"hbeat.basic":  decodeHbeat,
	"x10.basic":    decodeX10,
	"ac.basic":     decodeAC,
	"x10.security": decodeX10Sec,
	"sensor.basic": decodeSensor,
}

var x10secCmdToState = map[string]string{
	"arm-home": "armed_home",
	"arm-away": "armed_away",
	"disarm":   "disarmed",
	"panic":    "triggered",
	"alert":    "triggered",
	"normal":   "armed_home", // not sure about this one
}

func sendMqttPacket(client *mqtt.Client, topic string, data string) {
	x := (*client).Publish(topic, 1, false, data)
	slog.Debug("sending mqtt packet", "topic", topic, "message", data)
	go utils.MqttError(x)
}

func ProcessXPL(pkt *XPLPacket, mqtt *mqtt.Client) {
	slog.Debug("received xpl packet", "packet", *pkt)
	dec, _ := decoders[pkt.MessageType]
	dec(pkt, mqtt)
}

func decodeLogs(pkt *XPLPacket, mqtt *mqtt.Client) {
	tp, ok := pkt.Data["type"]
	if ok {
		if tp == "inf" {
			slog.Info("xpl log received", "message", pkt.Data["text"], "code", pkt.Data["code"])
		} else if tp == "wrn" {
			slog.Warn("xpl log received", "message", pkt.Data["text"], "code", pkt.Data["code"])
		} else if tp == "err" {
			slog.Error("xpl log received", "message", pkt.Data["text"], "code", pkt.Data["code"])
		}
	}
}

func decodeHbeat(pkt *XPLPacket, mqtt *mqtt.Client) {
	slog.Debug("xpl heartbeat", "interval", pkt.Data["interval"], "version", pkt.Data["version"], "ip", pkt.Data["ip"])
}

func decodeX10(pkt *XPLPacket, c *mqtt.Client) {
	dev, ok := pkt.Data["device"]
	if !ok {
		return
	}
	command, ok := pkt.Data["command"]
	if !ok {
		return
	}

	topic := Topic{
		MessageType: pkt.MessageType,
		DeviceID:    dev,
		DeviceType:  "X10",
		Action:      "state",
	}

	state := "ON"
	switch command {
	case "on":
		topic.DeviceParam = "switch"
	case "off":
		topic.DeviceParam = "switch"
		state = "OFF"
	case "all_lights_on":
		topic.DeviceParam = "all"
	case "all_lights_off":
		topic.DeviceParam = "all"
		state = "OFF"
	case "bright":
		topic.DeviceParam = "bright"
	case "dim":
		topic.DeviceParam = "bright"
		state = "OFF"
	}

	cfg := HAConfig{
		CommandTopic: topic.StringO(Topic{Action: "set"}),
		StateTopic:   topic.String(),
		Device: HADevice{
			Identifiers: []string{dev},
			Name:        dev,
			Model:       pkt.MessageType,
		},
		UniqueID: "x2m" + pkt.MessageType + dev + topic.DeviceParam,
	}
	sendHassPacket(c, "switch", dev, cfg)
	sendMqttPacket(c, topic.String(), state)
}

func decodeAC(pkt *XPLPacket, c *mqtt.Client) {
	addr, ok := pkt.Data["address"]
	if !ok {
		return
	}
	unit, ok := pkt.Data["unit"]
	if !ok {
		return
	}
	command, ok := pkt.Data["command"]
	if !ok {
		return
	}

	topic := Topic{
		MessageType: pkt.MessageType,
		DeviceID:    addr,
		DeviceType:  unit,
		DeviceParam: "switch",
		Action:      "state",
	}
	cfg := HAConfig{
		CommandTopic: topic.StringO(Topic{Action: "set"}),
		StateTopic:   topic.String(),
		Device: HADevice{
			Identifiers: []string{addr, unit},
			Name:        addr,
			Model:       pkt.MessageType,
		},
		UniqueID: "x2m" + pkt.MessageType + addr + unit + topic.DeviceParam,
	}

	if command == "preset" {
		ct := topic.StringO(Topic{DeviceParam: "brightness", Action: "set"})
		cfg.BrightnessScale = 15
		cfg.BrightnessStateTopic = topic.StringO(Topic{DeviceParam: "brightness"})
		cfg.BrightnessCommandTopic = ct
		cfg.UniqueID += "brightness"
		sendHassPacket(c, "light", addr+unit, cfg)
		sendMqttPacket(c, topic.String(), "ON")
		level, ok := pkt.Data["level"]
		if ok {
			sendMqttPacket(c, ct, level)
		}
	} else if command == "on" {
		sendHassPacket(c, "switch", addr+unit, cfg)
		sendMqttPacket(c, topic.String(), "ON")
	} else {
		sendHassPacket(c, "switch", addr+unit, cfg)
		sendMqttPacket(c, topic.String(), "OFF")
	}
}

func decodeX10Sec(pkt *XPLPacket, c *mqtt.Client) {
	dev, ok := pkt.Data["device"]
	if !ok {
		return
	}
	tp, ok := pkt.Data["type"]
	if !ok {
		tp = "unknown"
	}
	command, ok := pkt.Data["command"]
	if !ok {
		return
	}

	topic := Topic{
		MessageType: pkt.MessageType,
		DeviceID:    dev,
		DeviceType:  tp,
		Action:      "state",
	}
	device := HADevice{
		Identifiers: []string{dev, tp},
		Name:        dev + " " + tp,
		Model:       pkt.MessageType,
	}
	uid := "x2m" + pkt.MessageType + dev + tp

	// low battery
	topic.DeviceParam = "low-battery"
	cfg := HAConfig{
		StateTopic: topic.String(),
		Device:     device,
		UniqueID:   uid + "battery",
	}
	sendHassPacket(c, "binary_sensor", dev, cfg)
	low, found := pkt.Data["low-battery"]
	if found && low == "true" {
		sendMqttPacket(c, topic.String(), "ON")
	} else {
		sendMqttPacket(c, topic.String(), "OFF")
	}

	// tamper
	topic.DeviceParam = "tamper"
	cfg = HAConfig{
		StateTopic: topic.String(),
		Device:     device,
		UniqueID:   uid + "tamper",
	}
	sendHassPacket(c, "binary_sensor", dev, cfg)
	tamper, found := pkt.Data["tamper"]
	if found && tamper == "true" {
		sendMqttPacket(c, topic.String(), "ON")
	} else {
		sendMqttPacket(c, topic.String(), "OFF")
	}

	// command
	switch command {
	case "arm-home", "arm-away", "disarm", "panic", "alert", "normal", "motion":
		topic.DeviceParam = "alarm"
		cfg = HAConfig{
			StateTopic:          topic.String(),
			CommandTopic:        topic.StringO(Topic{Action: "set"}),
			Device:              device,
			SupportedFeatures:   []string{"arm_home", "arm_away", "trigger"},
			CodeDisarmRequired:  false,
			CodeArmRequired:     false,
			CodeTriggerRequired: false,
			UniqueID:            uid + "alarm",
		}
		sendHassPacket(c, "alarm_control_panel", dev, cfg)
		sendMqttPacket(c, topic.String(), x10secCmdToState[command])

		topic.DeviceParam = "triggered"
		cfg = HAConfig{
			StateTopic: topic.String(),
			Device:     device,
			UniqueID:   uid + "triggered",
		}
		sendHassPacket(c, "binary_sensor", dev, cfg)
		if command == "alert" || command == "panic" || command == "motion" {
			sendMqttPacket(c, topic.String(), "ON")
		} else {
			sendMqttPacket(c, topic.String(), "OFF")
		}

	case "light", "dark":
		topic.DeviceParam = "brightness"
		cfg = HAConfig{
			StateTopic: topic.String(),
			Device:     device,
			UniqueID:   uid + "brightness",
		}
		sendHassPacket(c, "binary_sensor", dev, cfg)
		if command == "light" {
			sendMqttPacket(c, topic.String(), "ON")
		} else {
			sendMqttPacket(c, topic.String(), "OFF")
		}

	case "lights-on", "lights-off":
		topic.DeviceParam = "switch"
		cfg = HAConfig{
			StateTopic:   topic.String(),
			CommandTopic: topic.StringO(Topic{Action: "set"}),
			Device:       device,
			UniqueID:     uid + "switch",
		}
		sendHassPacket(c, "switch", dev, cfg)
		if command == "lights-on" {
			sendMqttPacket(c, topic.String(), "ON")
		} else {
			sendMqttPacket(c, topic.String(), "OFF")
		}
	}
}

func decodeSensor(pkt *XPLPacket, c *mqtt.Client) {
	dev, ok := pkt.Data["device"]
	if !ok {
		return
	}
	tp := "unknown"
	s := strings.Split(dev, " ")
	if len(s) == 2 {
		dev = s[0]
		tp = s[1]
	}

	param, ok := pkt.Data["type"]
	if !ok {
		return
	}

	topic := Topic{
		MessageType: pkt.MessageType,
		DeviceID:    dev,
		DeviceType:  tp,
		DeviceParam: param,
		Action:      "state",
	}
	cfg := HAConfig{
		Device: HADevice{
			Identifiers: []string{tp, dev},
			Name:        dev + " " + tp,
			Model:       pkt.MessageType,
		},
		UniqueID: "x2m" + pkt.MessageType + dev + tp + param,
	}

	value, ok := pkt.Data["current"]
	if !ok {
		return
	}

	switch param {
	case "temp", "setpoint":
		cfg.Unit = "°C"
		cfg.DeviceClass = "temperature"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "voltage":
		cfg.Unit = "V"
		cfg.DeviceClass = "voltage"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "input":
		sendHassPacket(c, "binary_sensor", dev, cfg)
		if value == "low" {
			sendMqttPacket(c, topic.String(), "OFF")
		} else {
			sendMqttPacket(c, topic.String(), "ON")
		}
	case "humidity":
		cfg.Unit = "%"
		cfg.DeviceClass = "humidity"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "status":
		cfg.DeviceClass = "enum"
		cfg.CommandTopic = cfg.StateTopic
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "pressure":
		cfg.Unit = "hPa"
		cfg.DeviceClass = "pressure"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "rainrate":
		cfg.Unit = "mm/h"
		cfg.DeviceClass = "precipitation_intensity"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "raintotal":
		cfg.Unit = "mm"
		cfg.DeviceClass = "precipitation"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "gust", "average_speed":
		cfg.Unit = "m/s"
		cfg.DeviceClass = "wind_speed"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "direction", "count", "uv":
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "battery":
		cfg.Unit = "%"
		cfg.DeviceClass = "battery"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "weight":
		cfg.Unit = "kg"
		cfg.DeviceClass = "weight"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "datetime":
		t, err := time.Parse("20060201150405", pkt.Data["datetime"])
		if err == nil {
			cfg.DeviceClass = "timestamp"
			sendHassPacket(c, "sensor", dev, cfg)
			sendMqttPacket(c, topic.String(), strconv.FormatInt(t.Unix(), 10))
		}
	case "current":
		cfg.Unit = "A"
		cfg.DeviceClass = "current"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "power":
		cfg.Unit = "kW"
		cfg.DeviceClass = "power"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	case "energy":
		cfg.Unit = "kWh"
		cfg.DeviceClass = "energy"
		sendHassPacket(c, "sensor", dev, cfg)
		sendMqttPacket(c, topic.String(), value)
	}
}
