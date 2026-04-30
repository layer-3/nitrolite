package database

import (
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"gorm.io/gorm"
)

// QueryDurationObserver records the time a single gorm DB operation took.
// Implemented by the runtime metric exporter to avoid an import cycle
// (the metrics package depends on pkg/{app,core,rpc}, store/database does not).
type QueryDurationObserver interface {
	ObserveDBQueryDuration(queryKind string, duration time.Duration)
}

// metricsStartKey is the gorm.DB context key the before-callback uses to stash
// the operation start timestamp for the after-callback to read.
const metricsStartKey = "nitronode:metrics:start"

// callbackName is registered on every gorm callback chain we instrument; if
// you change it remember to keep the before/after pair in sync.
const callbackName = "nitronode:metrics"

// queryKinds is the set of gorm callback chains we hook. Order matches the
// gorm callback registry; "raw" covers Raw / Exec catch-all on the gorm v2
// callback chain "raw".
var queryKinds = []string{"create", "query", "update", "delete", "row", "raw"}

// RegisterMetricsCallbacks installs gorm callbacks that observe wall-clock
// duration of each Create / Query / Update / Delete / Row / Raw operation
// onto obs. Pass nil to skip registration (test / sqlite-in-memory cases).
//
// The callbacks add no per-call allocation beyond a single time.Time stashed
// in the gorm context dict.
func RegisterMetricsCallbacks(db *gorm.DB, obs QueryDurationObserver) error {
	if db == nil {
		return errors.New("database: nil gorm.DB")
	}
	if obs == nil {
		return nil
	}

	for _, kind := range queryKinds {
		kind := kind // pin for closures
		// Explicit case per kind + default that errors out, so adding a new
		// entry to queryKinds without an arm here trips at runtime instead
		// of silently hooking the wrong gorm processor. (Type of `chain` is
		// gorm's unexported *callbacks.processor; inferred from the dummy
		// init so the compile-time variable type lines up.)
		chain := db.Callback().Query()
		switch kind {
		case "create":
			chain = db.Callback().Create()
		case "query":
			chain = db.Callback().Query()
		case "update":
			chain = db.Callback().Update()
		case "delete":
			chain = db.Callback().Delete()
		case "row":
			chain = db.Callback().Row()
		case "raw":
			chain = db.Callback().Raw()
		default:
			return fmt.Errorf("database: unknown query kind %q", kind)
		}

		// Use Replace, not Register: Replace is idempotent on the callback
		// name, so calling RegisterMetricsCallbacks twice on the same gorm.DB
		// (test helpers, DI restart) doesn't double-fire the after callback
		// and double-count duration into the histogram.
		if err := chain.Before("*").Replace(callbackName+":before:"+kind, func(tx *gorm.DB) {
			tx.Set(metricsStartKey, time.Now())
		}); err != nil {
			return err
		}
		if err := chain.After("*").Replace(callbackName+":after:"+kind, func(tx *gorm.DB) {
			v, ok := tx.Get(metricsStartKey)
			if !ok {
				return
			}
			start, ok := v.(time.Time)
			if !ok {
				return
			}
			obs.ObserveDBQueryDuration(kind, time.Since(start))
		}); err != nil {
			return err
		}
	}
	return nil
}

// RegisterDBStatsCollector registers the standard go_sql_* collector on reg
// for the underlying *sql.DB, so pool-state gauges and wait counters become
// scrapable. Returns an error if the gorm DB doesn't expose a *sql.DB
// (sqlite-in-memory variants normally do; misconfigured pools don't).
//
// Emits, under "nitronode" db name:
//
//	go_sql_max_open_connections
//	go_sql_open_connections
//	go_sql_in_use_connections
//	go_sql_idle_connections
//	go_sql_wait_count_total
//	go_sql_wait_duration_seconds_total
//	go_sql_max_idle_closed_total
//	go_sql_max_idle_time_closed_total
//	go_sql_max_lifetime_closed_total
func RegisterDBStatsCollector(db *gorm.DB, reg prometheus.Registerer) error {
	if db == nil {
		return errors.New("database: nil gorm.DB")
	}
	if reg == nil {
		return errors.New("database: nil prometheus registerer")
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return reg.Register(collectors.NewDBStatsCollector(sqlDB, "nitronode"))
}
