package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"

	"github.com/droso-hass/xpl2mqtt/cmd"
	"github.com/droso-hass/xpl2mqtt/utils"
	"github.com/droso-hass/xpl2mqtt/xpl"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	cmd.Parse()

	opts := mqtt.NewClientOptions()
	opts.SetOrderMatters(false)
	opts.SetAutoReconnect(true)
	opts.SetClientID(cmd.ConfigData.ClientID)
	opts.AddBroker(cmd.ConfigData.MqttBroker)
	opts.SetUsername(cmd.ConfigData.MqttUsername)
	opts.SetPassword(cmd.ConfigData.MqttPassword)
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: cmd.ConfigData.MqttVerifySSL})

	fmt.Printf("%v\n", cmd.ConfigData)

	client := mqtt.NewClient(opts)
	srv := xpl.NewServer(xpl.XPLPort, &client)

	err := utils.MqttError(client.Connect())
	if err != nil {
		log.Fatal(err)
	}
	x := client.Subscribe(cmd.ConfigData.MqttBaseTopic+"/#", 0, func(c mqtt.Client, m mqtt.Message) { xpl.ProcessMqtt(c, m, srv) })
	err = utils.MqttError(x)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("xpl2mqtt started")
	srv.Run()
}
