// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package db

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/vidar-team/Cardinal/internal/conf"
	"github.com/vidar-team/Cardinal/internal/dbutil"
)

var AllTables = []interface{}{
	&Action{},
	&Bulletin{},
	&Challenge{},
	&Flag{},
	&GameBox{},
	&Log{},
	&Manager{},
	&Team{},
}

// Init initializes the database.
func Init() error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		conf.Database.User,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.Name,
		conf.Database.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return dbutil.Now()
		},
	})
	if err != nil {
		return errors.Wrap(err, "open connection")
	}

	// Migrate databases.
	if db.AutoMigrate(AllTables...) != nil {
		return errors.Wrap(err, "auto migrate")
	}

	SetDatabaseStore(db)

	return nil
}

// SetDatabaseStore sets the database table store.
func SetDatabaseStore(db *gorm.DB) {
	Actions = NewActionsStore(db)
	Bulletins = NewBulletinsStore(db)
	Challenges = NewChallengesStore(db)
	Flags = NewFlagsStore(db)
	GameBoxes = NewGameBoxesStore(db)
	Logs = NewLogsStore(db)
	Managers = NewManagersStore(db)
	Teams = NewTeamsStore(db)
}
