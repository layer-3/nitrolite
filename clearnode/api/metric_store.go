package api

import (
	"github.com/layer-3/nitrolite/clearnode/metrics"
	"github.com/layer-3/nitrolite/clearnode/store/database"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
)

// metricStore wraps a DatabaseStore to buffer metric callbacks during a DB transaction.
// Callbacks are executed only after the transaction commits successfully via flush().
type metricStore struct {
	database.DatabaseStore
	m         metrics.RuntimeMetricExporter
	callbacks []func()
}

func (s *metricStore) RecordTransaction(tx core.Transaction, applicationID string) error {
	if err := s.DatabaseStore.RecordTransaction(tx, applicationID); err != nil {
		return err
	}
	s.callbacks = append(s.callbacks, func() {
		s.m.RecordTransaction(tx.Asset, tx.TxType, tx.Amount, applicationID)
	})
	return nil
}

func (s *metricStore) StoreUserState(state core.State, applicationID string) error {
	if err := s.DatabaseStore.StoreUserState(state, applicationID); err != nil {
		return err
	}
	s.callbacks = append(s.callbacks, func() {
		s.m.IncUserState(state.Asset, state.HomeLedger.BlockchainID, state.Transition.Type, applicationID)
	})
	return nil
}

func (s *metricStore) UpdateAppSession(session app.AppSessionV1) error {
	if err := s.DatabaseStore.UpdateAppSession(session); err != nil {
		return err
	}
	s.callbacks = append(s.callbacks, func() {
		s.m.IncAppStateUpdate(session.ApplicationID)
	})
	return nil
}

func (s *metricStore) StoreChannelSessionKeyState(state core.ChannelSessionKeyStateV1) error {
	if err := s.DatabaseStore.StoreChannelSessionKeyState(state); err != nil {
		return err
	}
	s.callbacks = append(s.callbacks, s.m.IncChannelSessionKeys)
	return nil
}

func (s *metricStore) StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error {
	if err := s.DatabaseStore.StoreAppSessionKeyState(state); err != nil {
		return err
	}
	s.callbacks = append(s.callbacks, s.m.IncAppSessionKeys)
	return nil
}

func (s *metricStore) flush() {
	for _, cb := range s.callbacks {
		cb()
	}
}
