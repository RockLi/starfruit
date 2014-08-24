/*
 * Copyright 2014 The flatpeach Authors. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package version

import (
	"fmt"
)

const (
	Major      = 0
	Minor      = 1
	PatchLevel = "pre-beta"
	MagicCode  = "1141119910738101108108105101"
)

func Version() string {
	return fmt.Sprintf("ircd-%d.%d-%s", Major, Minor, PatchLevel)
}
