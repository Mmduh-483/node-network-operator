// Copyright 2018 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networkd

import (
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"
)

func init() {
	register.Register(register.NegativeTest, NetworkdDropinInvalidExtension())
}

func NetworkdDropinInvalidExtension() types.Test {
	name := "Create a networkd drop-in"
	in := types.GetBaseDisk()
	out := in
	config := `{
		"ignition": { "version": "2.2.0" },
		"networkd": {
			"units": [{
				"name": "static.network",
				"dropins": [{
					"name": "dropin.network",
					"contents": "[Match]\nName=enp2s0\n\n[Network]\nAddress=192.168.0.15/24\nGateway=192.168.0.1\n"
				}]
			}]
		}
	}`

	return types.Test{
		Name:              name,
		In:                in,
		Out:               out,
		Config:            config,
		ConfigShouldBeBad: true,
	}
}
