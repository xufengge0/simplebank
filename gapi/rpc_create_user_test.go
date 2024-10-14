package gapi

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	mockdb "github.com/techschool/simplebank/db/mock"
	db "github.com/techschool/simplebank/db/sqlc"
	"github.com/techschool/simplebank/pb"
	"github.com/techschool/simplebank/util"
	mockwk "github.com/techschool/simplebank/worker/mock"
)

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTXParams // 要匹配的参数
	password string                // 用于验证的密码
}

func (expected eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actualArg, ok := x.(db.CreateUserTXParams) // 尝试将 x 转换为 db.CreateUserParams
	if !ok {
		return false // 如果转换失败，返回 false
	}

	err := util.CheckPassword(expected.password, actualArg.HashedPassword) // 验证密码
	if err != nil {
		return false // 如果密码验证失败，返回 false
	}
	expected.arg.HashedPassword = actualArg.HashedPassword // 设定 hashed password

	return reflect.DeepEqual(expected.arg.CreateUserParams, actualArg.CreateUserParams) // 比较 arg 和传入的参数是否相等
}
func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

// 自定义gomock.Eq(arg)匹配器
func EqCreateUserTxParams(arg db.CreateUserTXParams, password string) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg, password}
}

func TestCreateUserAPI(t *testing.T) {

	user, password := RandomUser(t)

	testCase := []struct {
		name          string
		req           *pb.CreateUserRequest
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserTXParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}
				// 设置期望:
				store.EXPECT().
					CreateUserTX(gomock.Any(), EqCreateUserTxParams(arg, password)).
					Times(1).
					Return(db.CreateUserTXResult{User: user}, nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, res)
				createUser := res.GetUser()
				require.Equal(t, user.Username, createUser.Username)
				require.Equal(t, user.FullName, createUser.FullName)
				require.Equal(t, user.Email, createUser.Email)
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

			taskDistributor := mockwk.NewMockTaskDistributor(ctrl)

			server, err := NewTestServer(t, store, taskDistributor)
			res, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
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
