package database

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type LifespanMetric struct {
	ID            string          `gorm:"column:id;primaryKey;size:66"`
	Name          string          `gorm:"column:name;not null"`
	Labels        datatypes.JSON  `gorm:"column:labels;type:text"`
	Value         decimal.Decimal `gorm:"column:value;type:varchar(78);not null"`
	LastTimestamp time.Time       `gorm:"column:last_timestamp;not null"`
	UpdatedAt     time.Time
}

func (LifespanMetric) TableName() string {
	return "lifespan_metrics"
}

// ChannelCount holds the result of a COUNT() GROUP BY query on channels.
type ChannelCount struct {
	Asset       string             `gorm:"column:asset"`
	Status      core.ChannelStatus `gorm:"column:status"`
	Count       uint64             `gorm:"column:count"`
	LastUpdated time.Time          `gorm:"column:last_updated"`
}

// GetChannelsCountByLabels computes channel count deltas since last processed timestamp,
// upserts them as lifespan metrics, and returns the updated totals.
func (s *DBStore) GetChannelsCountByLabels() ([]ChannelCount, error) {
	metricName := "channels_total"

	lastProcessedTimestamp, err := s.GetLifetimeMetricLastTimestamp(metricName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last processed timestamp: %w", err)
	}

	var deltas []ChannelCount
	err = s.db.Raw(`
		SELECT asset,
		       status AS status,
		       COUNT(channel_id)::bigint AS count,
		       MAX(updated_at) AS last_updated
		FROM channels
		WHERE updated_at > ?
		GROUP BY asset, status
	`, lastProcessedTimestamp).Scan(&deltas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to compute channel deltas: %w", err)
	}

	if len(deltas) > 0 {
		now := time.Now()
		valuesSQL := make([]string, 0, len(deltas))
		args := make([]any, 0, len(deltas)*6)

		for i, d := range deltas {
			labelsMap := map[string]string{
				"asset":  d.Asset,
				"status": d.Status.String(),
			}
			labelsJSON, err := json.Marshal(labelsMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal labels for asset=%s status=%s: %w", d.Asset, d.Status, err)
			}

			id, err := getMetricID(metricName, "asset", d.Asset, "status", d.Status.String())
			if err != nil {
				return nil, fmt.Errorf("failed to compute metric ID for asset=%s status=%s: %w", d.Asset, d.Status, err)
			}

			deltaValue := decimal.NewFromUint64(d.Count)
			base := i * 6
			valuesSQL = append(valuesSQL,
				fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d)",
					base+1, base+2, base+3, base+4, base+5, base+6,
				),
			)

			args = append(args,
				id,                         // $1
				metricName,                 // $2
				datatypes.JSON(labelsJSON), // $3
				deltaValue,                 // $4
				d.LastUpdated,              // $5
				now,                        // $6
			)
		}

		upsertSQL := fmt.Sprintf(`
			INSERT INTO lifespan_metrics (id, name, labels, value, last_timestamp, updated_at)
			VALUES %s
			ON CONFLICT (id) DO UPDATE
			SET
				value = lifespan_metrics.value + EXCLUDED.value,
				last_timestamp = GREATEST(lifespan_metrics.last_timestamp, EXCLUDED.last_timestamp),
				updated_at = now()
		`, strings.Join(valuesSQL, ","))

		if err := s.db.Exec(upsertSQL, args...).Error; err != nil {
			return nil, fmt.Errorf("failed to upsert lifespan metrics: %w", err)
		}
	}

	var results []ChannelCount
	err = s.db.Raw(`
		SELECT labels->>'asset' AS asset,
		       labels->>'status' AS status,
		       value::bigint AS count,
		       last_timestamp AS last_updated
		FROM lifespan_metrics
		WHERE name = ?
	`, metricName).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to read lifespan metrics: %w", err)
	}

	return results, nil
}

// AppSessionCount holds the result of a COUNT() GROUP BY query on app sessions.
type AppSessionCount struct {
	Application string               `gorm:"column:application_id"`
	Status      app.AppSessionStatus `gorm:"column:status"`
	Count       uint64               `gorm:"column:count"`
	LastUpdated time.Time            `gorm:"column:last_updated"`
}

// GetAppSessionsCountByLabels computes app session count deltas since last processed timestamp,
// upserts them as lifespan metrics, and returns the updated totals.
func (s *DBStore) GetAppSessionsCountByLabels() ([]AppSessionCount, error) {
	metricName := "app_sessions_total"

	lastProcessedTimestamp, err := s.GetLifetimeMetricLastTimestamp(metricName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last processed timestamp: %w", err)
	}

	// 2) Compute deltas since lastProcessedTimestamp.
	var deltas []AppSessionCount
	err = s.db.Raw(`
		SELECT application_id,
		       status AS status,
		       COUNT(id)::bigint AS count,
		       MAX(updated_at) AS last_updated
		FROM app_sessions_v1
		WHERE updated_at > ?
		GROUP BY application_id, status
	`, lastProcessedTimestamp).Scan(&deltas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to compute app sessions deltas: %w", err)
	}

	if len(deltas) > 0 {
		now := time.Now()
		valuesSQL := make([]string, 0, len(deltas))
		args := make([]any, 0, len(deltas)*6)

		for i, d := range deltas {
			labelsMap := map[string]string{
				"application_id": d.Application,
				"status":         d.Status.String(),
			}
			labelsJSON, err := json.Marshal(labelsMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal labels for application_id=%s status=%s: %w", d.Application, d.Status, err)
			}

			id, err := getMetricID(metricName, "application_id", d.Application, "status", d.Status.String())
			if err != nil {
				return nil, fmt.Errorf("failed to compute metric ID for application_id=%s status=%s: %w", d.Application, d.Status, err)
			}

			deltaValue := decimal.NewFromUint64(d.Count)
			base := i * 6
			valuesSQL = append(valuesSQL,
				fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d)",
					base+1, base+2, base+3, base+4, base+5, base+6,
				),
			)

			args = append(args,
				id,                         // $1
				metricName,                 // $2
				datatypes.JSON(labelsJSON), // $3
				deltaValue,                 // $4
				d.LastUpdated,              // $5
				now,                        // $6
			)
		}

		upsertSQL := fmt.Sprintf(`
			INSERT INTO lifespan_metrics (id, name, labels, value, last_timestamp, updated_at)
			VALUES %s
			ON CONFLICT (id) DO UPDATE
			SET
				value = lifespan_metrics.value + EXCLUDED.value,
				last_timestamp = GREATEST(lifespan_metrics.last_timestamp, EXCLUDED.last_timestamp),
				updated_at = now()
		`, strings.Join(valuesSQL, ","))

		if err := s.db.Exec(upsertSQL, args...).Error; err != nil {
			return nil, fmt.Errorf("failed to upsert lifespan metrics: %w", err)
		}
	}

	var results []AppSessionCount
	err = s.db.Raw(`
		SELECT labels->>'application_id' AS application_id,
		       labels->>'status' AS status,
		       value::bigint AS count,
		       last_timestamp AS last_updated
		FROM lifespan_metrics
		WHERE name = ?
	`, metricName).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to read lifespan metrics: %w", err)
	}

	return results, nil
}

// TotalValueLocked holds the total value locked for a given asset, along with the last update timestamp.
type TotalValueLocked struct {
	Asset       string          `gorm:"column:asset"`
	Domain      string          `gorm:"column:domain"`
	Value       decimal.Decimal `gorm:"column:value"`
	LastUpdated time.Time       `gorm:"column:last_updated"`
}

func (s *DBStore) GetTotalValueLocked() ([]TotalValueLocked, error) {
	metricName := "total_value_locked"

	lastProcessedTimestamp, err := s.GetLifetimeMetricLastTimestamp(metricName)
	if err != nil {
		return nil, fmt.Errorf("failed to get last processed timestamp: %w", err)
	}

	// Compute net TVL deltas since lastProcessedTimestamp:
	// - channels: deposits (tx_type=10) minus withdrawals (tx_type=11)
	// - app_sessions: commits (tx_type=40) minus releases (tx_type=41)
	var deltas []TotalValueLocked
	err = s.db.Raw(`
		SELECT domain, asset_symbol AS asset, SUM(net) AS value, MAX(created_at) AS last_updated
		FROM (
			SELECT 'channels' AS domain, asset_symbol,
			       CASE WHEN tx_type = ? THEN amount ELSE -amount END AS net,
			       created_at
			FROM transactions
			WHERE tx_type IN (?, ?) AND created_at > ?
			UNION ALL
			SELECT 'app_sessions' AS domain, asset_symbol,
			       CASE WHEN tx_type = ? THEN amount ELSE -amount END AS net,
			       created_at
			FROM transactions
			WHERE tx_type IN (?, ?) AND created_at > ?
		) t
		GROUP BY domain, asset_symbol
	`,
		core.TransactionTypeHomeDeposit, core.TransactionTypeHomeDeposit, core.TransactionTypeHomeWithdrawal, lastProcessedTimestamp,
		core.TransactionTypeCommit, core.TransactionTypeCommit, core.TransactionTypeRelease, lastProcessedTimestamp,
	).Scan(&deltas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to compute TVL deltas: %w", err)
	}

	if len(deltas) > 0 {
		now := time.Now()
		valuesSQL := make([]string, 0, len(deltas))
		args := make([]any, 0, len(deltas)*6)

		for i, d := range deltas {
			labelsMap := map[string]string{
				"domain": d.Domain,
				"asset":  d.Asset,
			}
			labelsJSON, err := json.Marshal(labelsMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal labels for domain=%s asset=%s: %w", d.Domain, d.Asset, err)
			}

			id, err := getMetricID(metricName, "domain", d.Domain, "asset", d.Asset)
			if err != nil {
				return nil, fmt.Errorf("failed to compute metric ID for domain=%s asset=%s: %w", d.Domain, d.Asset, err)
			}

			base := i * 6
			valuesSQL = append(valuesSQL,
				fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d)",
					base+1, base+2, base+3, base+4, base+5, base+6,
				),
			)

			args = append(args,
				id,                         // $1
				metricName,                 // $2
				datatypes.JSON(labelsJSON), // $3
				d.Value,                    // $4
				d.LastUpdated,              // $5
				now,                        // $6
			)
		}

		upsertSQL := fmt.Sprintf(`
			INSERT INTO lifespan_metrics (id, name, labels, value, last_timestamp, updated_at)
			VALUES %s
			ON CONFLICT (id) DO UPDATE
			SET
				value = lifespan_metrics.value + EXCLUDED.value,
				last_timestamp = GREATEST(lifespan_metrics.last_timestamp, EXCLUDED.last_timestamp),
				updated_at = now()
		`, strings.Join(valuesSQL, ","))

		if err := s.db.Exec(upsertSQL, args...).Error; err != nil {
			return nil, fmt.Errorf("failed to upsert lifespan metrics: %w", err)
		}
	}

	var results []TotalValueLocked
	err = s.db.Raw(`
		SELECT labels->>'domain' AS domain,
		       labels->>'asset' AS asset,
		       value,
		       last_timestamp AS last_updated
		FROM lifespan_metrics
		WHERE name = ?
	`, metricName).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to read lifespan metrics: %w", err)
	}

	return results, nil
}

// CountActiveUsers returns the number of distinct users who had channel state updates
// within the given duration, grouped by asset. If asset is empty, counts across all assets.
// ActiveCountByLabel holds a count grouped by a label (asset or application_id).
type ActiveCountByLabel struct {
	Label string `gorm:"column:label"`
	Count uint64 `gorm:"column:count"`
}

// CountActiveUsers returns distinct user counts per asset and an "all" aggregate
// for users with channel state updates within the given window.
func (s *DBStore) CountActiveUsers(window time.Duration) ([]ActiveCountByLabel, error) {
	since := time.Now().Add(-window)

	var results []ActiveCountByLabel
	err := s.db.Raw(`
		SELECT asset AS label, COUNT(DISTINCT user_wallet) AS count
		FROM user_balances
		WHERE updated_at > ?
		GROUP BY asset
	`, since).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	// "ALL" aggregate: distinct users across all assets.
	var total uint64
	err = s.db.Model(&UserBalance{}).
		Select("COUNT(DISTINCT user_wallet)").
		Where("updated_at > ?", since).
		Scan(&total).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total active users: %w", err)
	}

	results = append(results, ActiveCountByLabel{Label: "ALL", Count: total})
	return results, nil
}

// CountActiveAppSessions returns app session counts per application within the given window.
func (s *DBStore) CountActiveAppSessions(window time.Duration) ([]ActiveCountByLabel, error) {
	since := time.Now().Add(-window)

	var results []ActiveCountByLabel
	err := s.db.Raw(`
		SELECT application_id AS label, COUNT(id) AS count
		FROM app_sessions_v1
		WHERE updated_at > ?
		GROUP BY application_id
	`, since).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active app sessions: %w", err)
	}

	return results, nil
}

// GetLifetimeMetricLastTimestamp returns the most recent last_timestamp among all metrics with the given name.
func (s *DBStore) GetLifetimeMetricLastTimestamp(name string) (time.Time, error) {
	var metric LifespanMetric
	err := s.db.Where("name = ?", name).
		Order("last_timestamp DESC").
		First(&metric).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to get metric last timestamp: %w", err)
	}

	return metric.LastTimestamp, nil
}

func getMetricID(name string, labels ...string) (string, error) {
	var labelsArray = []string{}
	labelsArray = append(labelsArray, labels...)

	stringTy, _ := abi.NewType("string", "", nil)
	stringSliceTy, _ := abi.NewType("string[]", "", nil)
	args := abi.Arguments{
		{Type: stringTy},      // name
		{Type: stringSliceTy}, // labels array
	}

	packed, err := args.Pack(
		name,
		labelsArray,
	)
	if err != nil {
		return "", fmt.Errorf("failed to pack app session request: %w", err)
	}

	return hexutil.Encode(crypto.Keccak256(packed)), nil
}
