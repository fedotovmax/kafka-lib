package eventspostgres

import (
	"log/slog"

	"github.com/fedotovmax/pgxtx"
)

type postgres struct {
	ex  pgxtx.Extractor
	log *slog.Logger
}

func New(ex pgxtx.Extractor, log *slog.Logger) *postgres {
	return &postgres{
		ex:  ex,
		log: log,
	}
}
