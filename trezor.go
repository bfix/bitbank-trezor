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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"trezor/protob"

	"github.com/google/gousb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// PinEntry interface for PIN/password dialogs
type PinEntry interface {
	Ask(prompt string) string
}

// Trezor device
type Trezor struct {
	dev      *gousb.Device  // USB device
	ctx      *gousb.Context // USB context
	firmware [3]uint32      // firmware version
	label    string         // device label
	pe       PinEntry       // associated entry dialog
}

//----------------------------------------------------------------------
// Trezor device management
//----------------------------------------------------------------------

// USB identifiers (vendor:product) of Trezor devices
// (Trezor One, Trezor Model T)
const (
	trezorVendor  = 0x1209
	trezorProduct = 0x53c1

	sig_fail = iota - 1
	sig_none
	sig_PinNeeded
	sig_PasswordNeeded
)

// Error codes
var (
	ErrTrezorPINNeeded      = errors.New("pin needed")
	ErrTrezorPasswordNeeded = errors.New("password required")
	ErrTrezorAddrPath       = errors.New("invalid address path")
)

// data (message) used to check for protocol version
var versionCheck = [65]byte{
	0, 63, 255, 255, 255, // ... 60 bytes following
}

// OpenTrezor: open a Trezor connected via USB
// (only one Trezor must be connected)
func OpenTrezor(pe PinEntry) (*Trezor, error) {
	// Initialize a new Context.
	ctx := gousb.NewContext()

	// find Trezor device(s)
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		if desc.Vendor == trezorVendor && desc.Product == trezorProduct {
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
		pe:  pe,
	}
	// get firmware version and device label
	features := new(protob.Features)
	if _, _, err := t.exchange(&protob.Initialize{}, features); err != nil {
		return nil, err
	}
	t.firmware = [3]uint32{features.GetMajorVersion(), features.GetMinorVersion(), features.GetPatchVersion()}
	t.label = features.GetLabel()

	return t, err
}

// Close Trezor device
func (t *Trezor) Close() (err error) {
	if err = t.dev.Close(); err != nil {
		return
	}
	return t.ctx.Close()
}

// Firmware returns the firmware versionof the device
func (t *Trezor) Firmware() [3]uint32 {
	return t.firmware
}

// Label returns the hman-readable label of the device
func (t *Trezor) Label() string {
	return t.label
}

//----------------------------------------------------------------------
// High-level device methods (functionality)
//----------------------------------------------------------------------

// Ping a device to see if it is still online.
func (t *Trezor) Ping() (err error) {
	_, _, err = t.exchange(&protob.Ping{}, new(protob.Success))
	return
}

// Unlock a device: may require PIN and/or password entry
func (t *Trezor) Unlock() (err error) {
	// get device status
	var res int
	if res, _, err = t.exchange(
		&protob.Ping{},
		new(protob.PinMatrixRequest),
		new(protob.PassphraseRequest),
		new(protob.Success),
	); err == nil {
		if res == 0 {
			// PIN required
			pin := t.pe.Ask("PIN")
			if len(pin) == 0 {
				err = ErrTrezorPINNeeded
			} else {
				if res, _, err = t.exchange(
					&protob.PinMatrixAck{
						Pin: &pin},
					new(protob.Success),
					new(protob.PassphraseRequest),
				); err != nil {
					return
				}
			}
		}
		if res == 1 {
			// Password required? Ask for it:
			passwd := t.pe.Ask("Password")
			if len(passwd) == 0 {
				err = ErrTrezorPasswordNeeded
			} else {
				_, _, err = t.exchange(
					&protob.PassphraseAck{
						Passphrase: &passwd,
					},
					new(protob.Success),
				)
			}
		}
	}
	return
}

func (t *Trezor) DeriveAddress(path string) (addr *protob.Address, err error) {
	// decode path
	if !strings.HasPrefix(path, "m/") {
		return nil, ErrTrezorAddrPath
	}
	pathInts := make([]uint32, 0)
	for _, id := range strings.Split(path[2:], "/") {
		var (
			j int64
			i uint32
		)
		if strings.HasSuffix(id, "'") {
			j, err = strconv.ParseInt(id[:len(id)-1], 10, 32)
			i = uint32(j) + (1 << 31)
		} else {
			j, err = strconv.ParseInt(id, 10, 32)
			i = uint32(j)
		}
		if err != nil {
			return
		}
		pathInts = append(pathInts, i)
	}
	// request address
	addr = &protob.Address{}
	if err = t.handleExchange(
		&protob.GetAddress{
			AddressN: pathInts,
		},
		addr,
	); err != nil {
		addr = &protob.Address{}
	}
	return
}

//----------------------------------------------------------------------
// Low-level message exchange and read/write operations.
//----------------------------------------------------------------------

func (t *Trezor) handleExchange(req protoreflect.ProtoMessage, results ...protoreflect.ProtoMessage) error {
	for {
		// perform exchange
		_, sig, err := t.exchange(req, results...)
		if err != nil {
			return err
		}
		// handle signals
		if done, err := t.handleSignal(sig); err == nil && done {
			if err != nil {
				return err
			}
			continue
		}
		break
	}
	return nil
}

func (t *Trezor) handleSignal(sig int) (done bool, err error) {
	var res int
	done = false
	if sig == sig_PinNeeded {
		// PIN required? Ask for it:
		pin := t.pe.Ask("PIN")
		if len(pin) == 0 {
			err = ErrTrezorPINNeeded
			return
		}
		if res, sig, err = t.exchange(
			&protob.PinMatrixAck{
				Pin: &pin},
			new(protob.Success),
			new(protob.PassphraseRequest),
		); err != nil {
			return
		}
		if res == 1 {
			sig = sig_PasswordNeeded
		} else {
			done = true
		}
	}
	if sig == sig_PasswordNeeded {
		// Password required? Ask for it:
		passwd := t.pe.Ask("Password")
		if len(passwd) == 0 {
			err = ErrTrezorPasswordNeeded
			return
		}
		_, _, err = t.exchange(
			&protob.PassphraseAck{
				Passphrase: &passwd,
			},
			new(protob.Success),
		)
		done = true
	}
	return
}

// exchange performs a data exchange with the Trezor wallet, sending it a
// message and retrieving the response. If multiple responses are possible, the
// method will also return the index of the destination object used.
func (t *Trezor) exchange(req proto.Message, results ...proto.Message) (res, sig int, err error) {
	// Construct the original message payload to chunk up
	data, err := proto.Marshal(req)
	if err != nil {
		return
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
		if _, err = t.write(chunk); err != nil {
			return
		}
	}
	// Stream the reply back from the wallet in 64 byte chunks
	var (
		kind  uint16
		reply []byte
	)
	for {
		// Read the next chunk from the Trezor wallet
		if _, err = t.read(chunk); err != nil {
			return
		}
		// Make sure the transport header matches
		if chunk[0] != 0x3f || (len(reply) == 0 && (chunk[1] != 0x23 || chunk[2] != 0x23)) {
			err = fmt.Errorf("invalid header")
			return
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
		if err = proto.Unmarshal(reply, failure); err == nil {
			err = fmt.Errorf("trezor: %s", failure.GetMessage())
		}
		return
	}
	if kind == uint16(protob.MessageType_MessageType_ButtonRequest) {
		// Trezor is waiting for user confirmation, ack and wait for the next message
		return t.exchange(&protob.ButtonAck{}, results...)
	}

	if kind == uint16(protob.MessageType_MessageType_PinMatrixRequest) {
		// Trezor requires a PIN entry
		sig = sig_PinNeeded
		return
	}

	for i, result := range results {
		if protob.Type(result) == kind {
			res = i
			err = proto.Unmarshal(reply, result)
			return
		}
	}
	expected := make([]string, len(results))
	for i, res := range results {
		expected[i] = protob.Name(protob.Type(res))
	}
	err = fmt.Errorf("trezor: expected reply types %s, got %s", expected, protob.Name(kind))
	return
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
