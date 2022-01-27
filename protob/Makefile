#----------------------------------------------------------------------
# This file is part of bitbank-trezor.
# Copyright (C) 2022 Bernd Fix >Y<
#
# bitbank-trezor is free software: you can redistribute it and/or modify it
# under the terms of the GNU Affero General Public License as published
# by the Free Software Foundation, either version 3 of the License,
# or (at your option) any later version.
#
# bitbank-trezor is distributed in the hope that it will be useful, but
# WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
# Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <http:#www.gnu.org/licenses/>.
#
# SPDX-License-Identifier: AGPL3.0-or-later
#----------------------------------------------------------------------

sync:
	rm *.proto
	git clone --depth 1 https://github.com/trezor/trezor-firmware/
	cp trezor-firmware/common/protob/*.proto .
	rm -rf trezor-firmware/

prepare: $(wildcard *.proto)
	$(foreach f, $^, sed \
		-e "/^option go_package = .*/d" \
		-e "s/option java_package/option go_package = \".;protob\";\noption java_package/" \
		-i $(f); \
	)

setup: sync prepare

targets = $(patsubst %.proto,%.pb.go,$(wildcard *.proto))

%.pb.go: %.proto
	protoc --go_out=. --go_opt=paths=source_relative $?

build: $(targets)

clean:
	rm -f *.pb.go
