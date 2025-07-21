package shared

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var Logger zerolog.Logger

func init() {
	// Enable stack traces for errors
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = time.DateTime
	Logger = zerolog.New(os.Stdout).With().Timestamp().Stack().Logger()
}
