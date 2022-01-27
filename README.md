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
So currently on extracting a public master key and generating a derived
coin address is supported.

The implementation of all other Trezor supported functionality is straight
foreward; new implementations will be available as soon as they happen :)

# Building the Trezor module

Building the module is only possible on Linux or BSD (MacOS, FreeBSD);
Windows is not supported because of a missing `libusb-1.0-dev` library.

## Prerequisites

* Go 1.16+
* libusb-1.0-dev

To make sure all external dependencies are available, issue the command

```bash
go mod tidy
```

## Generating sources from protobuf definitions

Change to the `protob` folder that contains the Protobuf definitions for
messages exchanged with the Trezor. The definition originate from the
`trezor-firmware` [repository on Github](https://github.com/trezor/trezor-firmware)
in the folder `common/protob`. These definitions should be up-to-date in
this module, but if you want to make sure you can copy and prepare the
definition files with:

```bash
make setup
```

In any case you need to generate the Go sources from the definitions with

```bash
make build
```

Check if the files have been created successfully.

## Building the module

Just issue a `go build` to make sure the module builds fine.

If you use the module in your own project, the Go module system should make
sure automatically that the build process is successful.

## Test the module

Testing is described in a
[separate page](https://github.com/bfix/bitbank-trezor/tree/main/test/README.md).
