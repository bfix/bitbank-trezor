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

package trezor

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	entryPin = iota
	entryPasswd
)

// PinEntry interface for PIN/password dialogs
type PinEntry interface {
	Ask(mode int) string
}

// ConsoleEntry handle PIN/password dialogs on stdin/stdout
type ConsoleEntry struct{}

// Ask for PIN or passphrase
func (e *ConsoleEntry) Ask(mode int) (in string) {
	if mode == entryPin {
		fmt.Println()
		fmt.Println("+---+---+---+")
		fmt.Println("| 7 | 8 | 9 |")
		fmt.Println("+---+---+---+")
		fmt.Println("| 4 | 5 | 6 |")
		fmt.Println("+---+---+---+")
		fmt.Println("| 1 | 2 | 3 |")
		fmt.Println("+---+---+---+")
		fmt.Println()
		fmt.Printf("PIN? ")
	} else {
		fmt.Printf("Password? ")
	}
	rdr := bufio.NewReader(os.Stdin)
	data, _, _ := rdr.ReadLine()
	in = strings.TrimSpace(string(data))
	fmt.Println()
	return
}
