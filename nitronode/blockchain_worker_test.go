package main

// type MockCustody struct {
// 	checkpointFn func() (common.Hash, error)
// 	mu           sync.Mutex
// 	callCount    int
// }

// var _ custody.CustodyInterface = (*MockCustody)(nil)

// func (m *MockCustody) Checkpoint(channelID common.Hash, state db.UnsignedState, userSig, serverSig sign.Signature, proofs []nitrolite.State) (common.Hash, error) {
// 	m.mu.Lock()
// 	m.callCount++
// 	m.mu.Unlock()

// 	if m.checkpointFn != nil {
// 		return m.checkpointFn()
// 	}
// 	return common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"), nil
// }

// func (m *MockCustody) CallCount() int {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	return m.callCount
// }

// func setupWorker(t *testing.T, custodyClients map[uint32]custody.CustodyInterface) (*BlockchainWorker, *gorm.DB, func()) {
// 	t.Helper()
// 	database, cleanup := api.SetupTestDB(t)
// 	logger := log.NewNoopLogger()
// 	worker := NewBlockchainWorker(database, custodyClients, logger)
// 	return worker, database, cleanup
// }

// func validCheckpointData(t *testing.T) []byte {
// 	t.Helper()
// 	data := db.CheckpointData{
// 		State:     db.UnsignedState{Version: 1},
// 		UserSig:   sign.Signature{1},
// 		ServerSig: sign.Signature{2},
// 	}
// 	bytes, err := json.Marshal(data)
// 	require.NoError(t, err)
// 	return bytes
// }

// func TestGetPendingActionsForChain(t *testing.T) {
// 	worker, database, cleanup := setupWorker(t, map[uint32]custody.CustodyInterface{1: &MockCustody{}})
// 	defer cleanup()

// 	channelIdA := common.HexToHash("ch1-a")
// 	channelIdB := common.HexToHash("ch1-b")

// 	require.NoError(t, database.Create(&db.BlockchainAction{ChannelID: channelIdB, ChainID: 1, Status: db.StatusPending, Data: []byte{1}, CreatedAt: time.Now()}).Error)
// 	require.NoError(t, database.Create(&db.BlockchainAction{ChannelID: channelIdA, ChainID: 1, Status: db.StatusPending, Data: []byte{1}, CreatedAt: time.Now().Add(-time.Second)}).Error)

// 	result, err := db.GetActionsForChain(worker.db, 1, 5)
// 	require.NoError(t, err)
// 	assert.Len(t, result, 2)
// 	assert.Equal(t, channelIdA, result[0].ChannelID)
// }

// func TestProcessAction(t *testing.T) {
// 	t.Run("Success case completes the action", func(t *testing.T) {
// 		mockCustody := &MockCustody{}
// 		worker, database, cleanup := setupWorker(t, map[uint32]custody.CustodyInterface{1: mockCustody})
// 		defer cleanup()

// 		action := &db.BlockchainAction{Type: db.ActionTypeCheckpoint, ChainID: 1, Data: validCheckpointData(t), Status: db.StatusPending}
// 		require.NoError(t, database.Create(action).Error)

// 		worker.processAction(context.Background(), *action)

// 		assert.Equal(t, 1, mockCustody.CallCount())
// 		var updatedAction db.BlockchainAction
// 		require.NoError(t, database.First(&updatedAction, action.ID).Error)
// 		assert.Equal(t, db.StatusCompleted, updatedAction.Status)
// 		assert.Equal(t, 0, updatedAction.Retries)
// 		expected := common.HexToHash("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef")
// 		assert.Equal(t, expected, updatedAction.TxHash)
// 	})

// 	t.Run("Permanent failure for missing custody client", func(t *testing.T) {
// 		worker, database, cleanup := setupWorker(t, map[uint32]custody.CustodyInterface{})
// 		defer cleanup()

// 		action := &db.BlockchainAction{Type: db.ActionTypeCheckpoint, ChainID: 999, Data: validCheckpointData(t), Status: db.StatusPending}
// 		require.NoError(t, database.Create(action).Error)

// 		worker.processAction(context.Background(), *action)

// 		var updatedAction db.BlockchainAction
// 		require.NoError(t, database.First(&updatedAction, action.ID).Error)
// 		assert.Equal(t, db.StatusFailed, updatedAction.Status)
// 		assert.Contains(t, updatedAction.Error, "no custody client for chain 999")
// 	})

// 	t.Run("Transient error on first attempt increments retries and leaves action pending", func(t *testing.T) {
// 		mockCustody := &MockCustody{
// 			checkpointFn: func() (common.Hash, error) {
// 				return common.Hash{}, errors.New("RPC node is down")
// 			},
// 		}
// 		worker, database, cleanup := setupWorker(t, map[uint32]custody.CustodyInterface{1: mockCustody})
// 		defer cleanup()

// 		action := &db.BlockchainAction{Type: db.ActionTypeCheckpoint, ChainID: 1, Data: validCheckpointData(t), Status: db.StatusPending, Retries: 0}
// 		require.NoError(t, database.Create(action).Error)

// 		worker.processAction(context.Background(), *action)

// 		assert.Equal(t, 1, mockCustody.CallCount())
// 		var updatedAction db.BlockchainAction
// 		require.NoError(t, database.First(&updatedAction, action.ID).Error)
// 		assert.Equal(t, db.StatusPending, updatedAction.Status)
// 		assert.Equal(t, 1, updatedAction.Retries)
// 		assert.Equal(t, "RPC node is down", updatedAction.Error)
// 	})

// 	t.Run("Permanent failure for invalid action data fails the action", func(t *testing.T) {
// 		mockCustody := &MockCustody{}
// 		worker, database, cleanup := setupWorker(t, map[uint32]custody.CustodyInterface{1: mockCustody})
// 		defer cleanup()

// 		action := &db.BlockchainAction{Type: db.ActionTypeCheckpoint, ChainID: 1, Data: []byte{1, 2, 3}, Status: db.StatusPending}
// 		require.NoError(t, database.Create(action).Error)

// 		worker.processAction(context.Background(), *action)

// 		assert.Equal(t, 0, mockCustody.CallCount())
// 		var updatedAction db.BlockchainAction
// 		require.NoError(t, database.First(&updatedAction, action.ID).Error)
// 		assert.Equal(t, db.StatusFailed, updatedAction.Status)
// 		assert.Contains(t, updatedAction.Error, "unmarshal checkpoint data")
// 	})

// 	t.Run("Action fails after 5 attempts", func(t *testing.T) {
// 		mockCustody := &MockCustody{
// 			checkpointFn: func() (common.Hash, error) {
// 				return common.Hash{}, errors.New("RPC still down")
// 			},
// 		}
// 		worker, database, cleanup := setupWorker(t, map[uint32]custody.CustodyInterface{1: mockCustody})
// 		defer cleanup()

// 		action := &db.BlockchainAction{
// 			Type:    db.ActionTypeCheckpoint,
// 			ChainID: 1,
// 			Data:    validCheckpointData(t),
// 			Status:  db.StatusPending,
// 			Retries: maxActionRetries,
// 		}
// 		require.NoError(t, database.Create(action).Error)

// 		worker.processAction(context.Background(), *action)

// 		assert.Equal(t, 1, mockCustody.CallCount())
// 		var updatedAction db.BlockchainAction
// 		require.NoError(t, database.First(&updatedAction, action.ID).Error)

// 		assert.Equal(t, db.StatusFailed, updatedAction.Status)
// 		assert.Equal(t, maxActionRetries, updatedAction.Retries)
// 		assert.Contains(t, updatedAction.Error, fmt.Sprintf("failed after %d retries: RPC still down", maxActionRetries))
// 	})
// }
