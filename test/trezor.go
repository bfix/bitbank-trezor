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
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"trezor"
)

type ConsoleEntry struct{}

func (e *ConsoleEntry) Ask(prompt string) (in string) {
	fmt.Printf("%s? ", prompt)
	rdr := bufio.NewReader(os.Stdin)
	data, _, _ := rdr.ReadLine()
	in = strings.TrimSpace(string(data))
	return
}

func main() {
	ce := new(ConsoleEntry)
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

	if err := trezor.Unlock(); err != nil {
		fmt.Printf("Unlock: %s\n", err.Error())
	} else {
		fmt.Println("Device unlocked.")
	}

	addr, err := trezor.DeriveAddress("m/49'/0'/0'/0/0")
	if err != nil {
		fmt.Println("DeriveAddress: " + err.Error())
		return
	}
	fmt.Println("Address: " + addr.String())
}
