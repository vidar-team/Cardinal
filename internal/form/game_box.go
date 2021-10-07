// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package form

type NewGameBox []struct {
	ChallengeID uint   `binding:"required"`
	TeamID      uint   `binding:"required"`
	IPAddress   string `binding:"required"`
	Port        uint
	Description string
	SSHPort     uint
	SSHUser     string
	SSHPassword string
}

type UpdateGameBox struct {
	ID          uint   `binding:"required"`
	IPAddress   string `binding:"required"`
	Port        uint
	Description string
	SSHPort     uint
	SSHUser     string
	SSHPassword string
}
