// Copyright 2021 E99p1ant. All rights reserved.
// Use of this source code is governed by an AGPL-style
// license that can be found in the LICENSE file.

package route

import (
	"net/http"

	"github.com/flamego/binding"
	"github.com/flamego/flamego"
	"github.com/flamego/session"

	"github.com/vidar-team/Cardinal/internal/context"
	"github.com/vidar-team/Cardinal/internal/form"
	"github.com/vidar-team/Cardinal/internal/i18n"
)

// NewRouter returns the router.
func NewRouter() *flamego.Flame {
	f := flamego.Classic()

	f.Use(session.Sessioner(session.Options{
		ReadIDFunc:  func(r *http.Request) string { return r.Header.Get("Authorization") },
		WriteIDFunc: func(w http.ResponseWriter, r *http.Request, sid string, created bool) {},
	}))

	f.Use(context.Contexter())
	f.Use(i18n.I18n())

	general := NewGeneralHandler()
	f.NotFound(general.NotFound)

	bulletin := NewBulletinHandler()
	challenge := NewChallengeHandler()
	gameBox := NewGameBoxHandler()

	f.Group("/api", func() {
		f.Any("/", general.Hello)
		f.Get("/init", general.Init)
		f.Get("/time")
		f.Get("/uploads")
		f.Get("/asteroid")

		f.Group("/team", func() {
			f.Post("/login", binding.JSON(form.TeamLogin{}))

			f.Group("", func() {
				f.Post("/logout")
				f.Post("/submitFlag")
				f.Get("/info")
				f.Get("/gameBoxes")
				f.Get("/bulletins")
				f.Get("/liveLog")
			})
		})

		manager := NewManagerHandler()
		f.Group("/manager", func() {
			f.Post("/login", form.Bind(form.ManagerLogin{}), manager.Login)

			f.Group("", func() {
				f.Post("/logout", manager.Logout)

				// Panel
				f.Get("/panel")
				f.Get("/logs")
				f.Get("/rank")

				// Challenge
				f.Get("/challenges", challenge.List)
				f.Post("/challenge", form.Bind(form.NewChallenge{}), challenge.New)
				f.Put("/challenge", form.Bind(form.UpdateChallenge{}), challenge.Update)
				f.Delete("/challenge", challenge.Delete)
				f.Post("/challenge/visible", form.Bind(form.SetChallengeVisible{}), challenge.SetVisible)

				// Team
				f.Get("/teams")
				f.Post("/teams", binding.JSON(form.NewTeam{}))
				f.Put("/team", binding.JSON(form.UpdateTeam{}))
				f.Delete("/team")
				f.Post("/team/resetPassword")

				// Game Box
				f.Get("/gameBoxes", gameBox.List)
				f.Post("/gameBox", form.Bind(form.NewGameBox{}), gameBox.New)
				f.Put("/gameBox", form.Bind(form.UpdateGameBox{}), gameBox.Update)
				f.Post("/gameBox/sshTest", gameBox.SSHTest)
				f.Post("/gameBox/refreshFlag", gameBox.RefreshFlag)
				f.Post("/gameBoxes/reset", gameBox.ResetAll)

				// Flag
				f.Get("/flags")
				f.Post("/flags")
				f.Get("/flags/export")

				// Bulletins
				f.Get("/bulletins", bulletin.List)
				f.Post("/bulletin", form.Bind(form.NewBulletin{}), bulletin.New)
				f.Put("/bulletin", form.Bind(form.UpdateBulletin{}), bulletin.Update)
				f.Delete("/bulletin", bulletin.Delete)

				// Asteroid
				f.Group("/asteroid", func() {
					f.Get("/status")
					f.Post("/attack")
					f.Post("/rank")
					f.Post("/status")
					f.Post("/round")
					f.Post("/easterEgg")
					f.Post("/time")
					f.Post("/clear")
				})

				// Check
				f.Get("/checkDown")
			}, manager.Authenticator)
		})
	})

	return f
}
