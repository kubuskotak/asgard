package tracer

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// EntHook defines the "mutation middleware". A function that gets a Mutator
// and returns a Mutator.
func EntHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			// Extract dynamic labels from mutation.
			var (
				labels      = fmt.Sprintf("%s.%s", m.Type(), m.Op().String())
				cx, span, l = StartSpanLogTrace(ctx, labels)
				start       = time.Now()
			)
			defer span.End()
			// Before mutation, start measuring time.
			v, err := next.Mutate(cx, m)
			if err != nil {
				l.Error().Err(err).Msg("error hook tracer")
			}
			// Stop time measure.
			var duration = time.Since(start)
			l.Info().Dur("duration", duration).
				Str("table", m.Type()).
				Str("operation", m.Op().String()).
				Msg("The latency of calls in milliseconds")
			return v, err
		})
	}
}

// EntInterceptor defines an execution middleware for various types of Ent queries.
func EntInterceptor() ent.Interceptor {
	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
			var (
				labels      = strings.ToLower(reflect.TypeOf(query).String())
				cx, span, l = StartSpanLogTrace(ctx, fmt.Sprintf("query.%s", labels))
				start       = time.Now()
			)
			if strings.HasSuffix(labels, "query") {
				defer span.End()
				labels = strings.ReplaceAll(labels, "query", "")
				labels = strings.ReplaceAll(labels, "*ent.", "")
			}
			// Before mutation, start measuring time.
			v, err := next.Query(cx, query)
			if err != nil {
				l.Error().Err(err).Msg("error interceptor tracer")
			}
			// Stop time measure.
			var duration = time.Since(start)
			l.Info().Dur("duration", duration).
				Str("table", labels).
				Str("operation", "query").
				Msg("The latency of calls in milliseconds")
			return v, err
		})
	})
}

// EntDriver is a driver that logs all driver operations.
type EntDriver struct {
	dialect.Driver                                                     // underlying driver.
	log            func(context.Context, zerolog.Level, string, error) // log function. defaults to log.Println.
}

// EntDriverWithContext gets a driver and a logging function, and returns
// a new tracer-driver that prints all outgoing operations with context.
func EntDriverWithContext(d dialect.Driver, logger func(context.Context, zerolog.Level, string, error)) dialect.Driver {
	drv := &EntDriver{d, logger}
	return drv
}

// Exec logs its params and calls the underlying driver Exec method.
func (d *EntDriver) Exec(ctx context.Context, query string, args, v any) error {
	if err := d.Driver.Exec(ctx, query, args, v); err != nil {
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("driver.Exec: query=%v args=%v", query, args), err)
		return err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("driver.Exec: query=%v args=%v", query, args), nil)
	return nil
}

// ExecContext logs its params and calls the underlying driver ExecContext method if it is supported.
func (d *EntDriver) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	drv, ok := d.Driver.(interface {
		ExecContext(context.Context, string, ...any) (sql.Result, error)
	})
	if !ok {
		var err = fmt.Errorf("Driver.ExecContext is not supported")
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("driver.ExecContext: query=%v args=%v", query, args), err)
		return nil, err
	}
	result, err := drv.ExecContext(ctx, query, args...)
	if err != nil {
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("driver.ExecContext: query=%v args=%v", query, args), err)
		return nil, err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("driver.ExecContext: query=%v args=%v", query, args), nil)
	return result, nil
}

// Query logs its params and calls the underlying driver Query method.
func (d *EntDriver) Query(ctx context.Context, query string, args, v any) error {
	if err := d.Driver.Query(ctx, query, args, v); err != nil {
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("driver.Query: query=%v args=%v", query, args), err)
		return err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("driver.Query: query=%v args=%v", query, args), nil)
	return nil
}

// QueryContext logs its params and calls the underlying driver QueryContext method if it is supported.
func (d *EntDriver) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	drv, ok := d.Driver.(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	})
	if !ok {
		var err = fmt.Errorf("Driver.QueryContext is not supported")
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("driver.QueryContext: query=%v args=%v", query, args), err)
		return nil, err
	}
	rows, err := drv.QueryContext(ctx, query, args...)
	if err != nil {
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("driver.QueryContext: query=%v args=%v", query, args), err)
		return nil, err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("driver.QueryContext: query=%v args=%v", query, args), nil)
	return rows, nil
}

// Tx adds a log-id for the transaction and calls the underlying driver Tx command.
func (d *EntDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("driver.Tx(%s): started", id), nil)
	return &Tx{tx, id, d.log, ctx}, nil
}

// BeginTx adds a log-id for the transaction and calls the underlying driver BeginTx command if it is supported.
func (d *EntDriver) BeginTx(ctx context.Context, opts *sql.TxOptions) (dialect.Tx, error) {
	drv, ok := d.Driver.(interface {
		BeginTx(context.Context, *sql.TxOptions) (dialect.Tx, error)
	})
	if !ok {
		var err = fmt.Errorf("Driver.BeginTx is not supported")
		d.log(ctx, zerolog.ErrorLevel, "driver.BeginTx: cannot started", err)
		return nil, err
	}
	tx, err := drv.BeginTx(ctx, opts)
	if err != nil {
		d.log(ctx, zerolog.InfoLevel, "driver.BeginTx: cannot started", err)
		return nil, err
	}
	id := uuid.New().String()
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("driver.BeginTx(%s): started", id), nil)
	return &Tx{tx, id, d.log, ctx}, nil
}

// Tx is a transaction implementation that logs all transaction operations.
type Tx struct {
	dialect.Tx                                                     // underlying transaction.
	id         string                                              // transaction logging id.
	log        func(context.Context, zerolog.Level, string, error) // log function. defaults to fmt.Println.
	ctx        context.Context                                     // underlying transaction context.
}

// Exec logs its params and calls the underlying transaction Exec method.
func (d *Tx) Exec(ctx context.Context, query string, args, v any) error {
	if err := d.Tx.Exec(ctx, query, args, v); err != nil {
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("Tx(%s).Exec: query=%v args=%v", d.id, query, args), err)
		return err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("Tx(%s).Exec: query=%v args=%v", d.id, query, args), nil)
	return nil
}

// ExecContext logs its params and calls the underlying transaction ExecContext method if it is supported.
func (d *Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	drv, ok := d.Tx.(interface {
		ExecContext(context.Context, string, ...any) (sql.Result, error)
	})
	if !ok {
		var err = fmt.Errorf("Tx.ExecContext is not supported")
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("Tx(%s).Exec: query=%v args=%v", d.id, query, args), err)
		return nil, err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("Tx(%s).ExecContext: query=%v args=%v", d.id, query, args), nil)
	return drv.ExecContext(ctx, query, args...)
}

// Query logs its params and calls the underlying transaction Query method.
func (d *Tx) Query(ctx context.Context, query string, args, v any) error {
	if err := d.Tx.Query(ctx, query, args, v); err != nil {
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("Tx(%s).Exec: query=%v args=%v", d.id, query, args), err)
		return err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("Tx(%s).Query: query=%v args=%v", d.id, query, args), nil)
	return nil
}

// QueryContext logs its params and calls the underlying transaction QueryContext method if it is supported.
func (d *Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	drv, ok := d.Tx.(interface {
		QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	})
	if !ok {
		var err = fmt.Errorf("Tx.QueryContext is not supported")
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("Tx(%s).Exec: query=%v args=%v", d.id, query, args), err)
		return nil, err
	}
	row, err := drv.QueryContext(ctx, query, args...)
	if err != nil {
		d.log(ctx, zerolog.ErrorLevel, fmt.Sprintf("Tx(%s).Exec: query=%v args=%v", d.id, query, args), err)
		return nil, err
	}
	d.log(ctx, zerolog.InfoLevel, fmt.Sprintf("Tx(%s).QueryContext: query=%v args=%v", d.id, query, args), nil)
	return row, nil
}

// Commit logs this step and calls the underlying transaction Commit method.
func (d *Tx) Commit() error {
	if err := d.Tx.Commit(); err != nil {
		d.log(d.ctx, zerolog.ErrorLevel, fmt.Sprintf("Tx(%s): committed", d.id), err)
		return err
	}
	d.log(d.ctx, zerolog.InfoLevel, fmt.Sprintf("Tx(%s): committed", d.id), nil)
	return nil
}

// Rollback logs this step and calls the underlying transaction Rollback method.
func (d *Tx) Rollback() error {
	if err := d.Tx.Rollback(); err != nil {
		d.log(d.ctx, zerolog.ErrorLevel, fmt.Sprintf("Tx(%s): rollbacked", d.id), err)
		return err
	}
	d.log(d.ctx, zerolog.InfoLevel, fmt.Sprintf("Tx(%s): rollbacked", d.id), nil)
	return nil
}
