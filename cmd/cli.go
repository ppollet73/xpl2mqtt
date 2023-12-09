package cmd

import (
	"flag"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/jamiealquiza/envy"
	"github.com/lmittmann/tint"
)

type Config struct {
	BroadcastAddress *net.UDPAddr
	Retries          int
	MqttBroker       string
	MqttUsername     string
	MqttPassword     string
	MqttVerifySSL    bool
	MqttBaseTopic    string
	ClientID         string
	HassDiscovery    bool
	XPLHops          int
	XPLTarget        string
}

var ConfigData Config

func Parse() {
	hn, _ := os.Hostname()

	saddr := flag.String("broadcast-address", "255.255.255.255:3865", "address to send the xpl packets")
	level := flag.String("log-level", "info", "log level")
	retries := flag.Int("retries", 2, "number of times a packet is sent")
	mqttBroker := flag.String("mqtt-broker", "", "broker address, in the format: [tcp,ssl,ws]://host:port")
	mqttUser := flag.String("mqtt-username", "", "mqtt username")
	mqttPass := flag.String("mqtt-password", "", "mqtt password")
	mqttSsl := flag.Bool("mqtt-verify-ssl", true, "verify ssl certs for mqtt")
	id := flag.String("client-id", hn, "identifier for this device")
	mqttBaseTopic := flag.String("mqtt-topic", "xpl2mqtt", "mqtt base topic")
	hass := flag.Bool("hass-discovery", true, "enable home-assistant mqtt discovery")
	xplTarget := flag.String("xpl-target", "*", "xpl target")
	xplHops := flag.Int("xpl-hops", 1, "xpl hops")

	envy.Parse("X2M")
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp4", *saddr)
	if err != nil {
		log.Fatalf("unable to resolve udp address: %s", err.Error())
	}

	logLevel := new(slog.LevelVar)
	switch *level {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warning":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	}
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      logLevel,
		TimeFormat: time.Kitchen,
	})
	slog.SetDefault(slog.New(handler))

	ConfigData = Config{
		BroadcastAddress: addr,
		Retries:          *retries,
		MqttBroker:       *mqttBroker,
		MqttUsername:     *mqttUser,
		MqttPassword:     *mqttPass,
		MqttVerifySSL:    *mqttSsl,
		ClientID:         *id,
		MqttBaseTopic:    *mqttBaseTopic,
		HassDiscovery:    *hass,
		XPLHops:          *xplHops,
		XPLTarget:        *xplTarget,
	}
}
