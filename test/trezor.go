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
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	trezor "github.com/bfix/bitbank-trezor"
)

type testData struct {
	Symb string `json:"symb"`
	Path string `json:"path"`
	Mode string `json:"mode"`
	Pk   string `json:"pk"`
	Addr string `json:"addr"`
}

func main() {
	var fname string
	flag.StringVar(&fname, "i", "testdata.json", "Name of JSON-encode test data file")
	flag.Parse()

	testData := make([]*testData, 0)
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(data, &testData); err != nil {
		log.Fatal(err)
	}

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
		fmt.Println("-----------------------------------")
		fmt.Printf("Coin: %s:\n", td.Symb)
		path := td.Path
		fmt.Printf("   Base path: %s\n", path)
		// get public master
		pk, err := trezor.GetXpub(path, td.Symb, td.Mode)
		if err != nil {
			fmt.Println("PublicMaster: " + err.Error())
			continue
		}
		fmt.Println("   PublicMaster  (ist): " + pk)
		if len(td.Pk) > 0 {
			fmt.Println("   PublicMaster (soll): " + td.Pk)
		}

		// get first address
		addr, err := trezor.GetAddress(path, td.Symb, td.Mode)
		if err != nil {
			fmt.Println("DeriveAddress: " + err.Error())
			continue
		}
		fmt.Println("   Address  (ist): " + addr)
		if len(td.Addr) > 0 {
			fmt.Println("   Address (soll): " + td.Addr)
		}
	}
}
