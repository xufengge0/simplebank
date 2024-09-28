package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)
func NewTestServer(t *testing.T, store db.Store) (*Server, error) {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDurationMinutes: time.Minute,
	}

	server, err := NewServer(config, store)
	if err!= nil {
		return nil, err
	}

	return server, nil
}
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}