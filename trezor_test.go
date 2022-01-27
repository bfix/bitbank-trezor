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
	"fmt"
	"testing"
)

func TestTrezor(t *testing.T) {
	trezor, err := OpenTrezor()
	if err != nil {
		t.Fatal(err)
	}
	if t == nil {
		t.Fatal("no Trezor found")
	}
	defer trezor.Close()
	fmt.Println("Trezor connected:")
	fmt.Printf("   Protocol: %d\n", trezor.version)
	fmt.Printf("    Firmare: %d.%d.%d\n", trezor.Firmware[0], trezor.Firmware[1], trezor.Firmware[2])
	fmt.Printf("      Label: '%s'\n", trezor.Label)
}
