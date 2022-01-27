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

package main

import (
	"fmt"
	"log"
	"strings"

	"trezor"
)

var (
	testData = []map[string]string{
		{
			"symb": "btc",
			"path": "m/49'/0'/0'",
			"mode": "P2SH",
			"pk":   "ypub6WmbbPMK8TMbtgxBusGbTBZGmqSgkjZuUzkJaqf8yLEXwPn54L8XacsqemKSgBiYUtkmLBFEv6LpzP73TWLKH14vRw2yvusS4xQUXiTxmNf",
			"addr": "36U4RLGuGZZhteZFtyWcQtR7kPuVywtQmj",
		},
		{
			"symb": "bch",
			"path": "m/44'/145'/0'",
			"mode": "P2PKH",
			"pk":   "xpub6Bragftuto4duWegVmyjnm4rYWHcL2cF7NV3pqJZ2RsHjkx5L3T7sEJhmgdTaz9MApN8Zrun9ypjxU4K1zEeFEc4YovVhs9SjP4uN87N4QX",
			"addr": "qr8e3qng3357fgw4eh6s0nkx3qcua2vv7ckedk9g86",
		},
		{
			"symb": "btg",
			"path": "m/49'/156'/0'",
			"mode": "P2SH",
			"pk":   "ypub6Xwi8Duv647y2EmRLz7RFA2Rjujr4YcqFYcwn6a8ZMTBS3uTkUr4KrhN6NnsSuY4Xnd3JEgD4VBEepQj29T8H1dmsMb3EicN5abpSF6vfRX",
			"addr": "AKexaufMJkFbrxvBb2e358cLVrkcZpQktF",
		},
		{
			"symb": "dash",
			"path": "m/44'/5'/0'",
			"mode": "P2PKH",
			"pk":   "drkpS1J1jLCsBzx6JoxnkUhSMT62bYxQSmLi2z7xNDPg16KUTnQLPBCjDdXrrtopkf3iXi3MSRQVL3kHvZKVKdPwVumX2e7Dm8n3aWowZiAhTTX",
			"addr": "XhTS8HSTPwrafUxQ1hJxW2Pw4ovn4dhFt7",
		},
		{
			"symb": "dgb",
			"path": "m/49'/20'/0'",
			"mode": "P2SH",
			"pk":   "ypub6XkbuwvsVZxFbDRy25pmzAnaW8KmspCtjhR3AM1SPZ3DdTsZYggPGyocq22FmiVpLeEFonpFKcjY7ML8ic6HKsozNBwKMLH42MfidSxMNna",
			"addr": "STwhxvh2dNC6yLoT1YNPn2WN8fWbGEEyFD",
		},
		{
			"symb": "doge",
			"path": "m/44'/3'/0'",
			"mode": "P2PKH",
			"pk":   "dgub8sN6Ldw8HMU2B2UkuYQGxooNs1bmji9DEY1o5DdvewZXe6i3HLtUiMtVX7315YKShw7nwtuQCYcu6MvQoXjJSUnLsDCC3uvWUVjveMcpdXS",
			"addr": "DKwSnhaCPRuyiPrEncfNPq2khD2DktjPYu",
		},
		{
			"symb": "ltc",
			"path": "m/49'/2'/0'",
			"mode": "P2SH",
			"pk":   "Mtub2sJWKRi18SqdXas5omXSWK8e3LZwmg4P8DD3p8pr5eAVinPz9L21bJLquTHXWdAAeeDUZ3XXRSfQj6fzh8JptDypA9m8t989VPYZxfT96HX",
			"addr": "MVwZr3fk1vLLKPo87RcoPmuMvsZAbsXSfi",
		},
		{
			"symb": "nmc",
			"path": "m/44'/7'/0'",
			"mode": "P2PKH",
			"pk":   "xpub6DDyvA2GMPPhidsu2BRL2hgxquG5re56MszZHGKzuUSzyC6FiQBjj6884jbqDXZgCRoXz9UqB9o2XZspHGNfHL2mCfaLmp4x2G8g8fcFkFQ",
			"addr": "NBwfrEYZiSnyqU4iQE79fD1A7Q5Rey5DB1",
		},
		{
			"symb": "vtc",
			"path": "m/49'/28'/0'",
			"mode": "P2SH",
			"pk":   "ypub6XzCXMBpswgFJsfZazMTZG4yoM2irUHrcyhJAswcHjnWrEu1PJTVKRAPPuC2iP5Q7veGE2exkiGhit224DoiX6fSirUSpin2T4wqokACCva",
			"addr": "323zw2Kr1DmKKs7dc4tPZ5tA15qZKPgre7",
		},
		{
			"symb": "zec",
			"path": "m/44'/133'/0'",
			"mode": "P2PKH",
			"pk":   "xpub6CbmrsDTmRxGJiDwi5LxSkTp5hHTcWMLY8sbDASHL7T4RxXqL69ZdB9w7Kp43xREcf4JFHojtcdcr3uBovTY3oaAA14PTupodyLqqVtxjSd",
			"addr": "t1XmgYfBtiALSpHKjdLW2j1novPVCHh6DXL",
		},
		{
			"symb": "eth",
			"path": "m/44'/60'/0'/0",
			"mode": "",
			"pk":   "xpub6DpTzFwx9SiUyN3QxGVkLREikjrat2ZbzPzEhjV8yr4HcjqywbRhmNgbmnz4wXk6G77veiEzi3Cvi5zTW5BLo29Yj96ZoDSZhovjCGhazY8",
			"addr": "0xf8c1bc608a08e95605ce145165d62d6e1dc752fe",
		},
		{
			"symb": "etc",
			"path": "m/44'/61'/0'/0",
			"mode": "",
			"pk":   "xpub6ESsVDThMxADPpFGt1ATbexxVCHSMcHqAB7W9C656RhK5w52sXhwnechzJZ8jSffJ6nu5jVHdjpStmdeUS2NzAMm8BQvAtV9pzGACQTp2BF",
			"addr": "0x8ff2bc448c3de3c2ab7d2be0262960d04135a2f1",
		},
	}
)

func main() {
	ce := new(trezor.ConsoleEntry)
	trezor, err := trezor.OpenTrezor(ce)
	if err != nil {
		log.Fatal(err)
	}
	if trezor == nil {
		log.Fatal("no Trezor found")
	}
	defer trezor.Close()

	fmt.Println("Trezor connected:")
	fw := trezor.Firmware()
	fmt.Printf("    Firmare: %d.%d.%d\n", fw[0], fw[1], fw[2])
	fmt.Printf("      Label: '%s'\n", trezor.Label())

	for _, td := range testData {
		path := td["path"]
		for strings.Count(path, "/") < 5 {
			path += "/0"
		}
		addr, err := trezor.DeriveAddress(path, td["symb"], td["mode"])
		if err != nil {
			fmt.Println("DeriveAddress: " + err.Error())
			continue
		}
		fmt.Println("Address  (ist): " + addr)
		fmt.Println("Address (soll): " + td["addr"])
	}
}
