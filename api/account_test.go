package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/techschool/simplebank/db/mock"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/token"
	"github.com/techschool/simplebank/util"
)

func TestGetAccount(t *testing.T) {
	user ,_:= RandomUser(t)
	account := RandomAccount(user.Username)
	
	testCase := []struct {
		name          string
		accountID     int64
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// 设置期望: 当 GetAccount 被调用时返回 account 和 nil
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)     // 检查状态码
				requireBodeMatchAccount(t, recorder.Body, account) // 检查body
			},
		},
		{
			name:      "Not Found",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// 设置期望: 当 GetAccount 被调用时返回 空account
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code) // 检查状态码

			},
		},
		{
			name:      "InternalErr",
			accountID: account.ID,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// 设置期望: 当 GetAccount 被调用时返回 空account
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code) // 检查状态码

			},
		},
		{
			name:      "BadRequest",
			accountID: 0,
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				// 设置期望: GetAccount 被调用0次
				store.EXPECT().
					GetAccount(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code) // 检查状态码

			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {

			// 初始化 GoMock 控制器
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server,err := NewTestServer(t,store)
			require.NoError(t, err)
			recorder := httptest.NewRecorder() // 用来记录HTTP响应结果

			// 拼接url
			url := fmt.Sprintf("/accounts/%d", tc.accountID)        
			// 生成get请求
			request, err := http.NewRequest(http.MethodGet, url, nil) 
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			// 使用 ServeHTTP 模拟对路由器的请求处理
			server.router.ServeHTTP(recorder, request)
			// 检查返回结果
			tc.checkResponse(t, recorder)
		})

	}

}

// 生成一个随机账户
func RandomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomBlance(),
		Currency: util.RandomCurrency(),
	}
}

// 检查返回的body是否正确
func requireBodeMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	date, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(date, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
