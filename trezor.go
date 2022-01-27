//----------------------------------------------------------------------
// This file is part of bitbank-trezor.
// Copyright (C) 2022 Bernd Fix >Y<
//
// bitbank-trezor is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// bitbank-trezor is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later
//----------------------------------------------------------------------
// Parts of this implementation are reused (in modified form) from the
// Go-Ethereum repository at Github (https://github.com/ethereum/go-ethereum/);
// especially the Trezor-related code at "/accounts/usbwallet/trezor.go"):
//
// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// This file contains the implementation for interacting with the Trezor hardware
// wallets. The wire protocol spec can be found on the SatoshiLabs website:
// https://doc.satoshilabs.com/trezor-tech/api-protobuf.html
//
//----------------------------------------------------------------------

package trezor

import (
	"encoding/binary"
	"fmt"
	"trezor/protob"

	"github.com/google/gousb"
	"google.golang.org/protobuf/proto"
)

// USB identifiers (vendor:product) of Trezor devices
// (Trezor One, Trezor Model T)
const (
	TrezorVendor  = 0x1209
	TrezorProduct = 0x53c1
)

// data (message) used to check for protocol version
var versionCheck = [65]byte{
	0, 63, 255, 255, 255, // ... 60 bytes following
}

type Message struct{}

func (msg *Message) Bytes() []byte {
	return nil
}

func NewMessage(data []byte) (*Message, error) {
	return nil, nil
}

type Trezor struct {
	dev      *gousb.Device  // USB device
	ctx      *gousb.Context // USB context
	version  int            // protocol version (1, 2)
	Firmware [3]uint32      // firmware version
	Label    string         // device label
}

func OpenTrezor() (*Trezor, error) {
	// Initialize a new Context.
	ctx := gousb.NewContext()

	// find Trezor device(s)
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if desc.Vendor == TrezorVendor && desc.Product == TrezorProduct {
			return true
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	// check that exactly one Trezor is found
	switch len(devs) {
	case 0:
		return nil, fmt.Errorf("can't open device")
	case 1:
		break
	default:
		return nil, fmt.Errorf("too many devices")
	}
	// instantiate single Trezor worker
	t := &Trezor{
		dev: devs[0],
		ctx: ctx,
	}
	// check protocol version
	versionCheck := func() (int, error) {
		n, err := t.write(versionCheck[:])
		if err != nil {
			return 0, err
		}
		if n == 65 {
			return 2, nil
		}
		n, err = t.write(versionCheck[1:])
		if err != nil {
			return 0, err
		}
		if n == 64 {
			return 1, nil
		}
		return 0, fmt.Errorf("unknown HID version")

	}
	t.version, err = versionCheck()

	// get firmware version and device label
	features := new(protob.Features)
	if _, err := t.exchange(&protob.Initialize{}, features); err != nil {
		return nil, err
	}
	t.Firmware = [3]uint32{features.GetMajorVersion(), features.GetMinorVersion(), features.GetPatchVersion()}
	t.Label = features.GetLabel()

	return t, err
}

func (t *Trezor) Close() (err error) {
	if err = t.dev.Close(); err != nil {
		return
	}
	return t.ctx.Close()
}

//----------------------------------------------------------------------
// Low-level message exchange and read/write operations.
//----------------------------------------------------------------------

// exchange performs a data exchange with the Trezor wallet, sending it a
// message and retrieving the response. If multiple responses are possible, the
// method will also return the index of the destination object used.
func (t *Trezor) exchange(req proto.Message, results ...proto.Message) (int, error) {
	// Construct the original message payload to chunk up
	data, err := proto.Marshal(req)
	if err != nil {
		return 0, err
	}
	payload := make([]byte, 8+len(data))
	copy(payload, []byte{0x23, 0x23})
	binary.BigEndian.PutUint16(payload[2:], protob.Type(req))
	binary.BigEndian.PutUint32(payload[4:], uint32(len(data)))
	copy(payload[8:], data)

	// Stream all the chunks to the device
	chunk := make([]byte, 64)
	chunk[0] = 0x3f // Report ID magic number

	for len(payload) > 0 {
		// Construct the new message to stream, padding with zeroes if needed
		if len(payload) > 63 {
			copy(chunk[1:], payload[:63])
			payload = payload[63:]
		} else {
			copy(chunk[1:], payload)
			copy(chunk[1+len(payload):], make([]byte, 63-len(payload)))
			payload = nil
		}
		// Send over to the device
		if _, err := t.write(chunk); err != nil {
			return 0, err
		}
	}
	// Stream the reply back from the wallet in 64 byte chunks
	var (
		kind  uint16
		reply []byte
	)
	for {
		// Read the next chunk from the Trezor wallet
		if _, err := t.read(chunk); err != nil {
			return 0, err
		}
		// Make sure the transport header matches
		if chunk[0] != 0x3f || (len(reply) == 0 && (chunk[1] != 0x23 || chunk[2] != 0x23)) {
			return 0, fmt.Errorf("invalid header")
		}
		// If it's the first chunk, retrieve the reply message type and total message length
		var payload []byte

		if len(reply) == 0 {
			kind = binary.BigEndian.Uint16(chunk[3:5])
			reply = make([]byte, 0, int(binary.BigEndian.Uint32(chunk[5:9])))
			payload = chunk[9:]
		} else {
			payload = chunk[1:]
		}
		// Append to the reply and stop when filled up
		if left := cap(reply) - len(reply); left > len(payload) {
			reply = append(reply, payload...)
		} else {
			reply = append(reply, payload[:left]...)
			break
		}
	}
	// Try to parse the reply into the requested reply message
	if kind == uint16(protob.MessageType_MessageType_Failure) {
		// Trezor returned a failure, extract and return the message
		failure := new(protob.Failure)
		if err := proto.Unmarshal(reply, failure); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("trezor: %s", failure.GetMessage())
	}
	if kind == uint16(protob.MessageType_MessageType_ButtonRequest) {
		// Trezor is waiting for user confirmation, ack and wait for the next message
		return t.exchange(&protob.ButtonAck{}, results...)
	}
	for i, res := range results {
		if protob.Type(res) == kind {
			return i, proto.Unmarshal(reply, res)
		}
	}
	expected := make([]string, len(results))
	for i, res := range results {
		expected[i] = protob.Name(protob.Type(res))
	}
	return 0, fmt.Errorf("trezor: expected reply types %s, got %s", expected, protob.Name(kind))
}

func (t *Trezor) read(data []byte) (int, error) {
	intf, done, err := t.dev.DefaultInterface()
	if err != nil {
		return 0, err
	}
	defer done()
	ep, err := intf.InEndpoint(1)
	if err != nil {
		return 0, err
	}
	return ep.Read(data)
}

func (t *Trezor) write(data []byte) (int, error) {
	intf, done, err := t.dev.DefaultInterface()
	if err != nil {
		return 0, err
	}
	defer done()
	ep, err := intf.OutEndpoint(1)
	if err != nil {
		return 0, err
	}
	return ep.Write(data)
}
