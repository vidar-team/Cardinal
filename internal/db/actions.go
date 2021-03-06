// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by Apache-2.0
// license that can be found in the LICENSE file.

package db

import (
	"context"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgconn"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ ActionsStore = (*actions)(nil)

// Actions is the default instance of the ActionsStore.
var Actions ActionsStore

// ActionsStore is the persistent interface for actions.
type ActionsStore interface {
	// Create creates a new action and persists to database.
	Create(ctx context.Context, opts CreateActionOptions) error
	// Get returns the actions according to the given options.
	Get(ctx context.Context, opts GetActionOptions) ([]*Action, error)
	// SetScore updates the action's score.
	SetScore(ctx context.Context, round, gameBoxID uint, score float64, replace ...bool) error
	// GetEmptyScore returns the empty score actions in the given round.
	GetEmptyScore(ctx context.Context, round uint, actionType ActionType) ([]*Action, error)
	// DeleteAll deletes all the actions.
	DeleteAll(ctx context.Context) error
}

// NewActionsStore returns a ActionsStore instance with the given database connection.
func NewActionsStore(db *gorm.DB) ActionsStore {
	return &actions{DB: db}
}

type ActionType uint

const (
	ActionTypeBeenAttack ActionType = iota
	ActionTypeCheckDown
	ActionTypeAttack
	ActionTypeServiceOnline
)

// Action represents the action such as check down or being attacked.
type Action struct {
	gorm.Model

	Type           ActionType `gorm:"uniqueIndex:action_unique_idx"`
	TeamID         uint       `gorm:"uniqueIndex:action_unique_idx"`
	ChallengeID    uint       `gorm:"uniqueIndex:action_unique_idx"`
	GameBoxID      uint       `gorm:"uniqueIndex:action_unique_idx"`
	AttackerTeamID uint       `gorm:"uniqueIndex:action_unique_idx"`
	Round          uint       `gorm:"uniqueIndex:action_unique_idx"`

	Score float64
}

type actions struct {
	*gorm.DB
}

type CreateActionOptions struct {
	Type           ActionType
	GameBoxID      uint
	AttackerTeamID uint
	Round          uint
}

var ErrDuplicateAction = errors.New("duplicate action")

func (db *actions) Create(ctx context.Context, opts CreateActionOptions) error {
	if opts.Type == ActionTypeCheckDown {
		opts.AttackerTeamID = 0
	}

	gameBoxStore := NewGameBoxesStore(db.DB)
	gameBox, err := gameBoxStore.GetByID(ctx, opts.GameBoxID)
	if err != nil {
		return err
	}

	tx := db.WithContext(ctx).Begin()
	var action Action
	err = tx.Model(&Action{}).Where(&Action{
		Type:           opts.Type,
		TeamID:         gameBox.TeamID,
		ChallengeID:    gameBox.ChallengeID,
		GameBoxID:      gameBox.ID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}).First(&action).Error
	if err == nil {
		tx.Rollback()
		return ErrDuplicateAction
	} else if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return errors.Wrap(err, "get action")
	}

	err = tx.Create(&Action{
		Type:           opts.Type,
		TeamID:         gameBox.TeamID,
		ChallengeID:    gameBox.ChallengeID,
		GameBoxID:      gameBox.ID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}).Error
	if err != nil {
		tx.Rollback()

		// NOTE: How to check if error type is DUPLICATE KEY in GORM.
		// https://github.com/go-gorm/gorm/issues/4037

		// Postgres
		if pgError, ok := err.(*pgconn.PgError); ok && errors.Is(err, pgError) && pgError.Code == "23505" {
			return ErrDuplicateAction
		}
		// MySQL
		var mysqlErr mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrDuplicateAction
		}
		return err
	}

	return tx.Commit().Error
}

type GetActionOptions struct {
	Type           ActionType
	TeamID         uint
	ChallengeID    uint
	GameBoxID      uint
	AttackerTeamID uint
	Round          uint
}

func (db *actions) Get(ctx context.Context, opts GetActionOptions) ([]*Action, error) {
	var actions []*Action
	return actions, db.WithContext(ctx).Model(&Action{}).Where(&Action{
		Type:           opts.Type,
		TeamID:         opts.TeamID,
		ChallengeID:    opts.ChallengeID,
		GameBoxID:      opts.GameBoxID,
		AttackerTeamID: opts.AttackerTeamID,
		Round:          opts.Round,
	}).Find(&actions).Error
}

var ErrActionNotExists = errors.New("action does not exist")

func (db *actions) SetScore(ctx context.Context, round, gameBoxID uint, score float64, replace ...bool) error {
	actions, err := db.Get(ctx, GetActionOptions{
		GameBoxID: gameBoxID,
		Round:     round,
	})
	if err != nil {
		return errors.Wrap(err, "get action")
	}
	if len(actions) == 0 {
		return ErrActionNotExists
	}

	action := actions[0]
	if action.Score == 0 || (len(replace) == 1 && replace[0]) {
		return db.WithContext(ctx).Model(&Action{}).Where("id = ?", action.ID).Update("score", score).Error
	}
	return nil
}

func (db *actions) GetEmptyScore(ctx context.Context, round uint, actionType ActionType) ([]*Action, error) {
	var actions []*Action
	return actions, db.WithContext(ctx).Model(&Action{}).Where("round = ? AND type = ? AND score = 0", round, actionType).Find(&actions).Error
}

func (db *actions) DeleteAll(ctx context.Context) error {
	return db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&Action{}).Error
}
