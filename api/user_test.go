package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/techschool/simplebank/db/mock"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/util"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams // 要匹配的参数
	password string              // 用于验证的密码
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams) // 尝试将 x 转换为 db.CreateUserParams
	if !ok {
		return false // 如果转换失败，返回 false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword) // 验证密码
	if err != nil {
		return false // 如果密码验证失败，返回 false
	}
	e.arg.HashedPassword = arg.HashedPassword // 设定 hashed password

	return reflect.DeepEqual(e.arg, arg) // 比较 arg 和传入的参数是否相等
}
func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

// 自定义gomock.Eq(arg)匹配器
func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {
	user, password := RandomUser(t)

	testCase := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": password,
				"fullname": user.FullName,
				"email":    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					Username: user.Username,
					FullName: user.FullName,
					Email:    user.Email,
				}
				// 设置期望:
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code) // 检查状态码
				requireBodyMatchUser(t, recorder.Body, user)   // 检查body
			},
		},
	}

	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {

			// 初始化 GoMock 控制器
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server ,err:= NewTestServer(t,store)
			require.NoError(t, err)
			recorder := httptest.NewRecorder() // 用来记录HTTP响应结果

			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			// 拼接url
			url := "/users"
			// 生成post请求
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			// 使用 ServeHTTP 模拟对路由器的请求处理
			server.router.ServeHTTP(recorder, request)
			// 检查返回结果
			tc.checkResponse(t, recorder)
		})
	}
}

// 生成一个随机user
func RandomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashpassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashpassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

// 检查返回的body是否正确
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	date, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(date, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}
