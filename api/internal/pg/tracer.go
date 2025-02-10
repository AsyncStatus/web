package pg

import (
	"api/internal/log"
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Tracer struct{ logger *log.Logger }

func (tracer *Tracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	tracer.logger.Logger.Sugar().Debugw("PGSQL", stripQuery(data.SQL), data.Args)
	return ctx
}

func (tracer *Tracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {}

var whitespaces = regexp.MustCompile(`\s+`)

func stripQuery(query string) string {
	return strings.Trim(whitespaces.ReplaceAllString(query, " "), " ")
}
