package xpl

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

var ErrInvalidPacket = errors.New("invalid xpl packet")

type XPLType string

const (
	TypeCmnd XPLType = "xpl-cmnd"
	TypeStat XPLType = "xpl-stat"
	TypeTrig XPLType = "xpl-trig"
)

var XPLTypes = []XPLType{TypeCmnd, TypeStat, TypeTrig}

type XPLPacket struct {
	Type        XPLType
	Hop         int
	Source      string
	Target      string
	MessageType string
	Data        map[string]string
}

func EncodePacket(pkt XPLPacket) string {
	data := fmt.Sprintf(
		"%s\n{\nhop=%d\nsource=%s\ntarget=%s\n}\n%s\n{\n",
		pkt.Type,
		pkt.Hop,
		pkt.Source,
		pkt.Target,
		pkt.MessageType,
	)
	content := ""
	for k, v := range pkt.Data {
		content += fmt.Sprintf("%s=%s\n", k, v)
	}
	data += content + "}"
	return data
}

func DecodePacket(data string) (XPLPacket, error) {
	split := strings.Split(data, "\n")
	if len(split) < 9 {
		slog.Warn("invalid xpl packet: too short")
		return XPLPacket{}, ErrInvalidPacket
	}

	pkt := XPLPacket{
		Type: XPLType(split[0]),
	}

	var xplHeader map[string]string
	xplData := map[string]string{}
	isOpen := false
	cpt := 0
	for _, x := range split {
		if x == "" {
		} else if x == "{" {
			xplData = map[string]string{}
			isOpen = true
		} else if x == "}" {
			isOpen = false
			if cpt == 0 {
				xplHeader = xplData
			} else if cpt == 1 {
				pkt.Data = xplData
			} else {
				slog.Warn("invalid xpl packet: too much data")
				return XPLPacket{}, ErrInvalidPacket
			}
			cpt++
		} else if !isOpen {
			if cpt == 0 {
				pkt.Type = XPLType(x)
			} else if cpt == 1 {
				pkt.MessageType = x
			} else {
				slog.Warn("invalid xpl packet: too much message types")
				return XPLPacket{}, ErrInvalidPacket
			}
		} else {
			kv := strings.Split(x, "=")
			if len(kv) != 2 {
				slog.Warn("invalid xpl packet: invalid key/value pair")
				return XPLPacket{}, ErrInvalidPacket
			}
			xplData[kv[0]] = kv[1]
		}
	}

	if !slices.Contains(XPLTypes, pkt.Type) {
		return XPLPacket{}, ErrInvalidPacket
	}

	h, ok := xplHeader["hop"]
	if !ok {
		slog.Warn("invalid xpl packet: invalid packet type")
		return XPLPacket{}, ErrInvalidPacket
	}
	hop, err := strconv.Atoi(h)
	if err != nil {
		slog.Warn("invalid xpl packet: cannot parse hop")
		return XPLPacket{}, ErrInvalidPacket
	}
	pkt.Hop = hop

	src, ok := xplHeader["source"]
	if !ok {
		slog.Warn("invalid xpl packet: no source")
		return XPLPacket{}, ErrInvalidPacket
	}
	pkt.Source = src

	target, ok := xplHeader["target"]
	if !ok {
		slog.Warn("invalid xpl packet: no target")
		return XPLPacket{}, ErrInvalidPacket
	}
	pkt.Target = target

	return pkt, nil
}
