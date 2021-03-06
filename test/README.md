[![Go Report Card](https://goreportcard.com/badge/github.com/bfix/bitbank-trezor)](https://goreportcard.com/report/github.com/bfix/bitbank-trezor)
[![GoDoc](https://godoc.org/github.com/bfix/bitbank-trezor?status.svg)](https://godoc.org/github.com/bfix/bitbank-trezor)

# Bitbank - Trezor

(c) 2022 Bernd Fix <brf@hoi-polloi.org>   >Y<

bitbank-trezor is free software: you can redistribute it and/or modify it
under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License,
or (at your option) any later version.

bitbank-trezor is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.

SPDX-License-Identifier: AGPL3.0-or-later

# WARNING

This software is not yet ready for productive use; it is work-in-progress.

It is designed and implemented for the purpose of initializing `bitbank-relay`
configuration files (see [bitbank-relay](https://github.com/bfix/bitbank-relay)).

The implementation of all other Trezor supported functionality is straight
foreward; new implementations will be available as soon as they happen :)

# Tests

## Building the test program

To build the test, execute the following commands in this folder:

```bash
go build
```
## Preparing the test data

The test program generates

The test data used must be stored in a file `testdata.json`. You can copy
`testdata-template.json` to get started.

The JSON-encoded data is an array of elements: 

```json
[
    {
        "symb": "btg",
        "path": "m/49'/156'/0'",
        "mode": "P2SH",
        "pk": "",
        "addr": ""
    },
    :
]
```
The list includes all coins that are supported by Trezor One (Trezor Model T
can handle more coins but that is currently not supported).

The fields of each test data element are as follows:

* **symb**: The coin ticker symbol in lowe-case characters
* **path**: The derivation path to the root of the coin-related branch
* **mode**: Either P2PKH (pay to public key hash) or P2SH (pay to script hash)
* **pk**: Public master key for the branch (optional; only if you know it)
* **addr**: First address (index 0) or the receiving branch (0). This is also
optional and can be left blank.

if `pk` and `addr` are known, the test program will compare the computed value
with the defined one and signal a mismatch. Otherwise it will just generate
the information that can be checked manually.

## Running the test program

Run the executable file `test` from the command line:

```bash
./test
```

By default it tries to load the test data from `testdata.json`; you can use
the `-i` flag to hand in a different file (of same format).

Make sure a Trezor is connected via USB with the computer.
