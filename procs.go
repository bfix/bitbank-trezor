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
	"strings"

	"github.com/bfix/bitbank-trezor/protob"
)

//======================================================================
// Bitcoin and derivates
//======================================================================

type BitcoinProc struct{}

// DeriveAddress returns an address referenced by the derivation path
func (p *BitcoinProc) GetAddress(dev *Trezor, path []uint32, coin, mode string) (addr string, err error) {
	// request generic address
	scriptType := scriptType(mode)
	coinName := coinName(coin)
	req := &protob.GetAddress{
		AddressN:   path,
		CoinName:   &coinName,
		ScriptType: &scriptType,
	}
	addrMsg := &protob.Address{}
	if err = dev.handleExchange(req, addrMsg); err == nil {
		addr = addrMsg.GetAddress()
	}
	// special post-processing
	addr = strings.Replace(addr, "bitcoincash:", "", 1)
	return
}

// PublicMaster returns the master public key for given derivation path
func (p *BitcoinProc) GetXpub(dev *Trezor, path []uint32, coin, mode string) (pk string, err error) {
	scriptType := scriptType(mode)
	coinName := coinName(coin)
	req := &protob.GetPublicKey{
		AddressN:   path,
		CoinName:   &coinName,
		ScriptType: &scriptType,
	}
	pkMsg := &protob.PublicKey{}
	if err = dev.handleExchange(req, pkMsg); err == nil {
		pk = pkMsg.GetXpub()
	}
	return
}

//======================================================================
// Ethereum and derivates
//======================================================================

type EthereumProc struct{}

// DeriveAddress returns an address referenced by the derivation path
func (p *EthereumProc) GetAddress(dev *Trezor, path []uint32, coin, mode string) (addr string, err error) {
	// request generic address
	req := &protob.EthereumGetAddress{
		AddressN: path,
	}
	addrMsg := &protob.EthereumAddress{}
	if err = dev.handleExchange(req, addrMsg); err == nil {
		addr = addrMsg.GetAddress()
	}
	// special post-processing
	addr = strings.Replace(addr, "bitcoincash:", "", 1)
	return
}

// PublicMaster returns the master public key for given derivation path
func (p *EthereumProc) GetXpub(dev *Trezor, path []uint32, coin, mode string) (pk string, err error) {
	req := &protob.EthereumGetPublicKey{
		AddressN: path,
	}
	pkMsg := &protob.EthereumPublicKey{}
	if err = dev.handleExchange(req, pkMsg); err == nil {
		pk = pkMsg.GetXpub()
	}
	return
}
