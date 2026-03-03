package database

import (
	"testing"
	"time"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppSessionV1_TableName(t *testing.T) {
	session := AppSessionV1{}
	assert.Equal(t, "app_sessions_v1", session.TableName())
}

func TestAppParticipantV1_TableName(t *testing.T) {
	participant := AppParticipantV1{}
	assert.Equal(t, "app_session_participants_v1", participant.TableName())
}

func TestDBStore_CreateAppSession(t *testing.T) {
	t.Run("Success - Create app session with single participant", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session := app.AppSessionV1{
			SessionID:   "session123",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := store.CreateAppSession(session)
		require.NoError(t, err)

		// Verify session was created
		var dbSession AppSessionV1
		err = db.Where("id = ?", "session123").Preload("Participants").First(&dbSession).Error
		require.NoError(t, err)

		assert.Equal(t, "session123", dbSession.ID)
		assert.Equal(t, "poker", dbSession.ApplicationID)
		assert.Equal(t, uint64(1), dbSession.Nonce)
		assert.Equal(t, `{"state": "active"}`, dbSession.SessionData)
		assert.Equal(t, uint8(100), dbSession.Quorum)
		assert.Equal(t, uint64(1), dbSession.Version)
		assert.Equal(t, app.AppSessionStatusOpen, dbSession.Status)
		assert.Len(t, dbSession.Participants, 1)
		assert.Equal(t, "0xuser123", dbSession.Participants[0].WalletAddress)
		assert.Equal(t, uint8(100), dbSession.Participants[0].SignatureWeight)
	})

	t.Run("Success - Create app session with multiple participants", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session := app.AppSessionV1{
			SessionID:   "session456",
			Application: "chess",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 50,
				},
				{
					WalletAddress:   "0xuser456",
					SignatureWeight: 50,
				},
			},
			SessionData: `{"state": "waiting"}`,
			Quorum:      75,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := store.CreateAppSession(session)
		require.NoError(t, err)

		// Verify session was created
		var dbSession AppSessionV1
		err = db.Where("id = ?", "session456").Preload("Participants").First(&dbSession).Error
		require.NoError(t, err)

		assert.Equal(t, "session456", dbSession.ID)
		assert.Equal(t, uint8(75), dbSession.Quorum)
		assert.Len(t, dbSession.Participants, 2)

		// Verify participants
		wallets := []string{dbSession.Participants[0].WalletAddress, dbSession.Participants[1].WalletAddress}
		assert.Contains(t, wallets, "0xuser123")
		assert.Contains(t, wallets, "0xuser456")
	})

	t.Run("Error - Duplicate session ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session := app.AppSessionV1{
			SessionID:   "session789",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := store.CreateAppSession(session)
		require.NoError(t, err)

		// Try to create again with same ID
		err = store.CreateAppSession(session)
		assert.Error(t, err)
	})
}

func TestDBStore_GetAppSession(t *testing.T) {
	t.Run("Success - Get existing session", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session := app.AppSessionV1{
			SessionID:   "session123",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session))

		result, err := store.GetAppSession("session123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "session123", result.SessionID)
		assert.Equal(t, "poker", result.Application)
		assert.Equal(t, uint64(1), result.Nonce)
		assert.Equal(t, `{"state": "active"}`, result.SessionData)
		assert.Equal(t, uint8(100), result.Quorum)
		assert.Equal(t, uint64(1), result.Version)
		assert.Equal(t, app.AppSessionStatusOpen, result.Status)
		assert.Len(t, result.Participants, 1)
		assert.Equal(t, "0xuser123", result.Participants[0].WalletAddress)
	})

	t.Run("No session found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		result, err := store.GetAppSession("nonexistent")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestDBStore_GetAppSessions(t *testing.T) {
	t.Run("Success - Get all sessions", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create multiple sessions
		session1 := app.AppSessionV1{
			SessionID:   "session1",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now().Add(-2 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		}

		session2 := app.AppSessionV1{
			SessionID:   "session2",
			Application: "chess",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser456",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now().Add(-1 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		}

		require.NoError(t, store.CreateAppSession(session1))
		require.NoError(t, store.CreateAppSession(session2))

		pagination := &core.PaginationParams{}
		sessions, metadata, err := store.GetAppSessions(nil, nil, app.AppSessionStatusVoid, pagination)
		require.NoError(t, err)

		assert.Len(t, sessions, 2)
		assert.Equal(t, uint32(2), metadata.TotalCount)

		// Should be ordered by created_at DESC
		assert.Equal(t, "session2", sessions[0].SessionID)
		assert.Equal(t, "session1", sessions[1].SessionID)
	})

	t.Run("Success - Filter by session ID", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session1 := app.AppSessionV1{
			SessionID:   "session1",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		session2 := app.AppSessionV1{
			SessionID:   "session2",
			Application: "chess",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser456",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session1))
		require.NoError(t, store.CreateAppSession(session2))

		sessionID := "session1"
		pagination := &core.PaginationParams{}
		sessions, metadata, err := store.GetAppSessions(&sessionID, nil, app.AppSessionStatusVoid, pagination)
		require.NoError(t, err)

		assert.Len(t, sessions, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "session1", sessions[0].SessionID)
	})

	t.Run("Success - Filter by participant", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session1 := app.AppSessionV1{
			SessionID:   "session1",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		session2 := app.AppSessionV1{
			SessionID:   "session2",
			Application: "chess",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser456",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session1))
		require.NoError(t, store.CreateAppSession(session2))

		participant := "0xuser123"
		pagination := &core.PaginationParams{}
		sessions, metadata, err := store.GetAppSessions(nil, &participant, app.AppSessionStatusVoid, pagination)
		require.NoError(t, err)

		assert.Len(t, sessions, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "session1", sessions[0].SessionID)
	})

	t.Run("Success - Filter by status", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session1 := app.AppSessionV1{
			SessionID:   "session1",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		session2 := app.AppSessionV1{
			SessionID:   "session2",
			Application: "chess",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser456",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "closed"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusClosed,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session1))
		require.NoError(t, store.CreateAppSession(session2))

		pagination := &core.PaginationParams{}
		sessions, metadata, err := store.GetAppSessions(nil, nil, app.AppSessionStatusOpen, pagination)
		require.NoError(t, err)

		assert.Len(t, sessions, 1)
		assert.Equal(t, uint32(1), metadata.TotalCount)
		assert.Equal(t, "session1", sessions[0].SessionID)
		assert.Equal(t, app.AppSessionStatusOpen, sessions[0].Status)
	})

	t.Run("Success - Pagination", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		// Create 3 sessions
		for i := 1; i <= 3; i++ {
			sessionID := "session" + string(rune(i+'0'))
			session := app.AppSessionV1{
				SessionID:   sessionID,
				Application: "poker",
				Nonce:       uint64(i),
				Participants: []app.AppParticipantV1{
					{
						WalletAddress:   "0xuser123",
						SignatureWeight: 100,
					},
				},
				SessionData: `{"state": "active"}`,
				Quorum:      100,
				Version:     1,
				Status:      app.AppSessionStatusOpen,
				CreatedAt:   time.Now().Add(time.Duration(i) * time.Minute),
				UpdatedAt:   time.Now().Add(time.Duration(i) * time.Minute),
			}
			require.NoError(t, store.CreateAppSession(session))
		}

		// Get first page (2 items)
		limit := uint32(2)
		offset := uint32(0)
		pagination := &core.PaginationParams{
			Limit:  &limit,
			Offset: &offset,
		}
		sessions, metadata, err := store.GetAppSessions(nil, nil, app.AppSessionStatusVoid, pagination)
		require.NoError(t, err)

		assert.Len(t, sessions, 2)
		assert.Equal(t, uint32(3), metadata.TotalCount)
		assert.Equal(t, uint32(1), metadata.Page)
		assert.Equal(t, uint32(2), metadata.PerPage)

		// Get second page
		offset = 2
		pagination.Offset = &offset
		sessions, metadata, err = store.GetAppSessions(nil, nil, app.AppSessionStatusVoid, pagination)
		require.NoError(t, err)

		assert.Len(t, sessions, 1)
		assert.Equal(t, uint32(2), metadata.Page)
	})

	t.Run("No sessions found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		sessionID := "nonexistent"
		pagination := &core.PaginationParams{}
		sessions, metadata, err := store.GetAppSessions(&sessionID, nil, app.AppSessionStatusVoid, pagination)
		require.NoError(t, err)

		assert.Empty(t, sessions)
		assert.Equal(t, uint32(0), metadata.TotalCount)
	})
}

func TestDBStore_UpdateAppSession(t *testing.T) {
	t.Run("Success - Update session data and version", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session := app.AppSessionV1{
			SessionID:   "session123",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session))

		// Update session
		session.SessionData = `{"state": "updated"}`
		session.Version = 2
		session.Status = app.AppSessionStatusClosed

		err := store.UpdateAppSession(session)
		require.NoError(t, err)

		// Verify update
		result, err := store.GetAppSession("session123")
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, `{"state": "updated"}`, result.SessionData)
		assert.Equal(t, uint64(2), result.Version)
		assert.Equal(t, app.AppSessionStatusClosed, result.Status)
	})

	t.Run("Error - Update non-existent session", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session := app.AppSessionV1{
			SessionID:   "nonexistent",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     2,
			Status:      app.AppSessionStatusClosed,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err := store.UpdateAppSession(session)
		require.Error(t, err) // Optimistic locking detects no rows affected
		assert.Contains(t, err.Error(), "concurrent modification detected")
	})

	t.Run("Error - Version mismatch (concurrent modification)", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()

		store := NewDBStore(db)

		session := app.AppSessionV1{
			SessionID:   "session456",
			Application: "poker",
			Nonce:       1,
			Participants: []app.AppParticipantV1{
				{
					WalletAddress:   "0xuser123",
					SignatureWeight: 100,
				},
			},
			SessionData: `{"state": "active"}`,
			Quorum:      100,
			Version:     1,
			Status:      app.AppSessionStatusOpen,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		require.NoError(t, store.CreateAppSession(session))

		// Try to update with wrong version (expecting version 2, but DB has version 1)
		session.SessionData = `{"state": "updated"}`
		session.Version = 3 // This expects DB to have version 2, but it has version 1

		err := store.UpdateAppSession(session)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "concurrent modification detected")
	})
}
