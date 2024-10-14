package gapi

import (
	"testing"
	"time"

	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
	"github.com/techschool/simplebank/worker"
)

func NewTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) (*Server, error) {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}

	server, err := NewServer(config, store,taskDistributor)
	if err != nil {
		return nil, err
	}

	return server, nil
}
