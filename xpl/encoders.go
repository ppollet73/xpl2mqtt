package xpl

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/droso-hass/xpl2mqtt/cmd"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var encoders = map[string](func(Topic, string, *Server)){
	"x10.basic":     encodeX10,
	"ac.basic":      encodeAC,
	"x10.security":  encodeX10sec,
	"control.basic": encodeControl,
}

var x10secStateToCmd = map[string]string{
	"ARM_HOME": "arm-home",
	"ARM_AWAY": "arm-away",
	"DISARM":   "disarm",
	"TRIGGER":  "panic",
}

func sendXplPacket(srv *Server, msgType string, data map[string]string) {
	p := XPLPacket{
		Type:        TypeCmnd,
		Hop:         cmd.ConfigData.XPLHops,
		Source:      fmt.Sprintf("xpl2mqtt-%s", cmd.ConfigData.ClientID),
		Target:      cmd.ConfigData.XPLTarget,
		MessageType: msgType,
		Data:        data,
	}
	slog.Debug("sending xpl packet", "packet", p)
	err := srv.Write(&p, cmd.ConfigData.BroadcastAddress, cmd.ConfigData.Retries)
	if err != nil {
		slog.Error("error sending xpl packet", "error", err.Error())
	}
}

func ProcessMqtt(client mqtt.Client, msg mqtt.Message, srv *Server) {
	p := string(msg.Payload())
	slog.Debug("received mqtt message", "topic", msg.Topic(), "message", p)
	t := Topic{}
	err := t.Parse(msg.Topic())
	if err != nil {
		slog.Error("error parsing topic", "error", err)
		return
	}
	if t.Action == "set" {
		enc, _ := encoders[t.MessageType]
		enc(t, p, srv)
	}
}

func encodeX10(topic Topic, payload string, srv *Server) {
	// value for brightness topic should be between 0 and 10
	// possible protocols are: X10,arc,flamingo,koppla,waveman,harrison,he105,rts10
	data := map[string]string{
		"device":   topic.DeviceID,
		"protocol": topic.DeviceType,
	}
	if topic.DeviceParam == "switch" && payload == "ON" {
		data["command"] = "on"
	} else if topic.DeviceParam == "switch" && payload == "OFF" {
		data["command"] = "off"
	} else if topic.DeviceParam == "all" && payload == "ON" {
		data["command"] = "all_lights_on"
	} else if topic.DeviceParam == "all" && payload == "OFF" {
		data["command"] = "all_lights_off"
	} else if topic.DeviceParam == "bright" && payload == "ON" {
		data["command"] = "bright"
	} else if topic.DeviceParam == "bright" && payload == "OFF" {
		data["command"] = "dim"
	} else if topic.DeviceParam == "brightness" {
		data["command"] = "on"
		val, err := strconv.Atoi(payload)
		if err != nil {
			return
		}
		data["level"] = strconv.Itoa(val * 10)
	}

	sendXplPacket(srv, topic.MessageType, data)
}

func encodeAC(topic Topic, payload string, srv *Server) {
	data := map[string]string{
		"address": topic.DeviceID,
		"unit":    topic.DeviceType,
	}
	if topic.DeviceParam == "switch" && payload == "ON" {
		data["command"] = "on"
	} else if topic.DeviceParam == "switch" && payload == "OFF" {
		data["command"] = "off"
	} else if topic.DeviceParam == "brightness" {
		data["command"] = "preset"
		data["level"] = payload
	}
	sendXplPacket(srv, topic.MessageType, data)
}

func encodeX10sec(topic Topic, payload string, srv *Server) {
	data := map[string]string{
		"device": topic.DeviceID,
	}

	switch topic.DeviceParam {
	case "alarm":
		data["command"] = x10secStateToCmd[payload]

	case "panic":
		if payload == "ON" {
			data["command"] = "panic"
		} else {
			data["command"] = "normal"
		}

	case "motion":
		if payload == "ON" {
			data["command"] = "motion"
		} else {
			data["command"] = "normal"
		}

	case "brightness":
		if payload == "ON" {
			data["command"] = "light"
		} else {
			data["command"] = "dark"
		}

	case "switch":
		if payload == "ON" {
			data["command"] = "lights-on"
		} else {
			data["command"] = "lights-off"
		}
	}
	sendXplPacket(srv, topic.MessageType, data)
}

func encodeControl(topic Topic, payload string, srv *Server) {
	data := map[string]string{
		"device": topic.DeviceID,
		"type":   topic.DeviceType,
	}
	if topic.DeviceType == "output" {
		if payload == "ON" {
			data["current"] = "high" // is it really current (and not command) ?
		} else {
			data["current"] = "low"
		}
		sendXplPacket(srv, topic.MessageType, data)
	}
}
