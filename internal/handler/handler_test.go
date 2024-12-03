package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/EgorcaA/create_db/internal/generator"
	"github.com/EgorcaA/create_db/internal/handler"
	"github.com/EgorcaA/create_db/internal/logger/handlers/slogdiscard"
	mocksredis "github.com/EgorcaA/create_db/internal/mocks/CacheClient"
	mocksdb "github.com/EgorcaA/create_db/internal/mocks/Database"
)

func TestHandle_message(t *testing.T) {

	msg := generator.GenerateFakeOrder()

	// Discard logger
	logger := slogdiscard.NewDiscardLogger()

	ctx := context.Background()

	// Define test cases
	tests := []struct {
		name           string
		mockDBSetup    func(mockDB *mocksdb.Database)
		mockCacheSetup func(mockCache *mocksredis.CacheClient)
	}{
		{
			name: "Successful order save and cache",
			mockDBSetup: func(mockDB *mocksdb.Database) {
				mockDB.On("InsertOrder", ctx, msg).Return(nil)
			},
			mockCacheSetup: func(mockCache *mocksredis.CacheClient) {
				mockCache.On("SaveOrder", ctx, msg).Return(nil)
			},
		},
		{
			name: "DB save fails",
			mockDBSetup: func(mockDB *mocksdb.Database) {
				mockDB.On("InsertOrder", ctx, msg).Return(errors.New("DB error"))
			},
			mockCacheSetup: func(mockCache *mocksredis.CacheClient) {
				// No cache interaction since DB save fails
			},
		},
		{
			name: "Cache save fails",
			mockDBSetup: func(mockDB *mocksdb.Database) {
				mockDB.On("InsertOrder", ctx, msg).Return(nil)
			},
			mockCacheSetup: func(mockCache *mocksredis.CacheClient) {
				mockCache.On("SaveOrder", ctx, msg).Return(errors.New("Cache error"))
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockCache := mocksredis.NewCacheClient(t)
			mockDB := mocksdb.NewDatabase(t)

			// Set up expectations
			tt.mockDBSetup(mockDB)
			tt.mockCacheSetup(mockCache)

			// Call the function
			handler.Handle_message(logger, ctx, mockCache, msg, mockDB)

			// Assert expectations
			mockDB.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}
