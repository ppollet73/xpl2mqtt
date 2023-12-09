# xPL2MQTT

## About this project

This project aims to bridge the [xPL](http://xplproject.org.uk/) protocol to [MQTT](https://mqtt.org/).

This was primarily designed to allow the usage of the [RFXLAN](https://web.archive.org/web/20130617090417/http://www.rfxcom.com/store/all/11201) with [Home Assistant](https://www.home-assistant.io/), but it should be possible to easily extend it to support other devices.

## Getting Started

### Binary

Download the correct binary for your OS/arch from the releases page and run it in command line.

To build from source: install golang, clone this repository and run `go build`

Run it: `./xpl2mqtt -broadcast-address "192.168.1.45:3865" -log-level debug -mqtt-broker "ssl://mqtt.domain.tld:8883" -mqtt-username "xpl2mqtt" -mqtt-password "PASSWORD"`

### Docker

You need to open the port `3865` in udp to receive the xPL messages and specify the ip:port of your RFXLAN with `X2M_BROADCAST_ADDRESS` (or use `--network=host`).

Use the provided docker image: `ghcr.io/droso-hass/xpl2mqtt`.

To manually build the docker image, follow the "build from source" steps above, and run `docker build . -t xpl2mqtt`.

Run it: `docker run -it --rm -p 3865:3865/udp -e X2M_MQTT_BROKER="ssl://mqtt.domain.tld:8883" -e X2M_BROADCAST_ADDRESS="192.168.1.45:3865" xpl2mqtt`

## Configuration

|CLI flag|Required|Default|Description|
|--|--|--|--|
|broadcast-address|false|255.255.255.255:3865|xPL broadcast address (used to send xPL messages) in the ip:port format|
|log-level|false|info|log level: debug,info,warning,error|
|retries|false|2|number of times a xPL packet is sent (as neither 433MHz and UDP are reliable)|
|mqtt-broker|true|-|mqtt broker address in the format {tcp,ssl,ws}://host:port|
|mqtt-username|false|-|mqtt username|
|mqtt-password|false|-|mqtt password|
|mqtt-verify-ssl|false|true|verify ssl cert for broker connection|
|client-id|false|hostname of the server|identifier used for both the mqtt broker and xPL source|
|mqtt-topic|false|xpl2mqtt|mqtt base topic|
|hass-discovery|false|true|enable home-assistant mqtt discovery|
|xpl-target|false|*|xpl target|
|xpl-hops|false|1|xpl max hops|

All cli flags can also be provided as environment variables (ex: `-broadcast-address` can be provided with the env var `X2M_BROADCAST_ADDRESS`).

## MQTT Format

Each topic only holds the raw value/command for the device.

The topics are formatted like this: `xpl2mqtt/<message_type>/<device_type>/<device_id>/<device_param>/<action>`

|Component|Description|Example|
|--|--|--|
|Message Type|the xPL message type|`ac.basic`, `sensor.basic`, ...|
|Device Type|the device type, for RFXLAN this is usually the `type` field of the xPL packet|`th-1`, `X10`, ...|
|Device ID|the device identifier, for RFXLAN this is usually the `address` or `device` field of the xPL packet|`0x12345678`|
|Device Param|specific parameter of the device|`temp` for the temperature value of a temp/hum sensor|
|Action|`state` when sending a value, `set` when sending a command|`state`, `set`|

## RFXLAN Usage

This project implements most of the [specification](https://web.archive.org/web/20140626135449/http://rfxcom.com/Documents/RFXCOM%20implementation%20xPL.pdf) (v7.8) provided by rfxcom.

If using home assistant, most of the devices should automatically show up if you enabled the autodiscovery.

Make sure that you have installed the xPL firmware (go on the web ui and check the firmare info, it should be something like: `RFXxPL_2_11.hex`) on your RFXLAN and NOT the tcp/ip one. If you need to change firmarwe, check the RFXLAN download section on their [website](https://web.archive.org/web/20140625050654/http://rfxcom.com/Downloads).

Also make sure that the Broadcast xPL address for the RFXLAN (in the web ui > Network Config) is set to `255.255.255.255` or to the address of the server running this software.

It has currently only been tested for DI.O plugs, but should work for the following schemas:
 - hbeat.basic (shown in debug logs)
 - log.basic (shown in logs)
 - x10.basic (r,w *)
 - x10.security (r,w)
 - ac.basic (r,w)
 - sensor.basic (r)
 - control.basic (r,w **)

\* only the X10 protocol is supported for reading, for writing, you may need to manually configure the mqtt topics in home assistant (please refer to the [specification](https://web.archive.org/web/20140626135449/http://rfxcom.com/Documents/RFXCOM%20implementation%20xPL.pdf) and the mqtt format).

\*\* only RFXLAN I/O lines are implemented, other types are documented but marked as unimplemented by the RFXLAN
