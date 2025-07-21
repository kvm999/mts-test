package main

import (
	"mts/internal/bootstrap"
	"shared"
)

func main() {
	app := bootstrap.NewApp()
	if err := app.Initialize(); err != nil {
		shared.Logger.Fatal().Err(err).Msg("app initialize")
	}
	if err := app.Start(); err != nil {
		shared.Logger.Fatal().Err(err).Msg("app start")
	}
}
