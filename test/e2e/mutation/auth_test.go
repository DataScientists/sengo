package mutation_test

import (
	"context"
	"fmt"
	"net/http"
	"sheng-go-backend/ent"
	"sheng-go-backend/pkg/infrastructure/router"
	"sheng-go-backend/testutil"
	"sheng-go-backend/testutil/e2e"
	"testing"

	"github.com/gavv/httpexpect/v2"
)

func TestAuth_Login(t *testing.T) {
	expect, client, teardown := e2e.Setup(t, e2e.SetupOption{
		TearDown: func(t *testing.T, client *ent.Client) {
			testutil.DropUser(t, client)
		},
	})

	defer teardown()

	tests := []struct {
		name    string
		arrange func(t *testing.T) string
		act     func(t *testing.T, email string) *httpexpect.Response
		assert  func(t *testing.T, got *httpexpect.Response)
		args    struct {
			ctx context.Context
		}
		teardown func(t *testing.T)
	}{
		{
			name: "It should login test user",
			arrange: func(t *testing.T) string {
				// Create a user and return its id
				resp := expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": `

          mutation {
        createUser (input:{
        name :"test",
          email:"test@mail.com",
        age : 34,
          password:"sometestpassword"
        }) {
        id
        name
        age
          email
        createdAt
        updatedAt
        }
        }
        `,
				}).Expect().Status(http.StatusOK)
				// Extract the ID of the created user
				data := e2e.GetData(resp).Object()
				createdUser := e2e.GetObject(data, "createUser")
				email := createdUser.Value("email").String().Raw()
				return string(email)
			},
			act: func(t *testing.T, email string) *httpexpect.Response {
				query := fmt.Sprintf(`
	mutation {
		login(input: {email: "%s", password: "sometestpassword"}) {
			accessToken
			refreshToken
			user {
				id
				name
			}
		}
	}`, email)
				return expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": query,
				}).Expect()
			},
			assert: func(t *testing.T, got *httpexpect.Response) {
				// Assert: Check if the user's details were updated correctly

				got.Status(http.StatusOK)
				data := e2e.GetData(got).Object()
				me := e2e.GetObject(data, "login")
				user := e2e.GetObject(me, "user")
				user.Value("name").String().IsEqual("test")
			},
			teardown: func(t *testing.T) {
				// Teardown : Cleanup user after test
				testutil.DropUser(t, client)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userId := tt.arrange(t)
			got := tt.act(t, userId)
			tt.assert(t, got)
			tt.teardown(t)
		})
	}
}

func TestAuth_RefreshToken(t *testing.T) {
	expect, client, teardown := e2e.Setup(t, e2e.SetupOption{
		TearDown: func(t *testing.T, client *ent.Client) {
			testutil.DropUser(t, client)
		},
	})

	defer teardown()

	tests := []struct {
		name    string
		arrange func(t *testing.T) string
		act     func(t *testing.T, refreshToken string) *httpexpect.Response
		assert  func(t *testing.T, got *httpexpect.Response)
		args    struct {
			ctx context.Context
		}
		teardown func(t *testing.T)
	}{
		{
			name: "It should generate new tokens",
			arrange: func(t *testing.T) string {
				// Create a user and return its id
				expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": `

          mutation {
        createUser (input:{
        name :"test",
          email:"test@mail.com",
        age : 34,
          password:"sometestpassword"
        }) {
        id
        name
        age
          email
        createdAt
        updatedAt
        }
        }
        `,
				}).Expect().Status(http.StatusOK)

				loginResponse := expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": `

          mutation {
        login (input:{
          email:"test@mail.com",
          password:"sometestpassword"
        
        }) {
        accessToken
          refreshToken
        }
        }
        `,
				}).Expect().Status(http.StatusOK)

				loginData := e2e.GetData(loginResponse).Object()
				loggedInUser := e2e.GetObject(loginData, "login")
				refreshToken := loggedInUser.Value("refreshToken").String().Raw()
				return string(refreshToken)
			},
			act: func(t *testing.T, refreshToken string) *httpexpect.Response {
				query := `
	                mutation {
		                refreshToken {
			              accessToken
			              refreshToken
			
		                }
	             }`
				return expect.POST(router.QueryPath).
					WithHeader("RefreshToken", fmt.Sprint(refreshToken)).
					WithJSON(map[string]string{
						"query": query,
					}).Expect()
			},
			assert: func(t *testing.T, got *httpexpect.Response) {
				// Assert: Check if the user's details were updated correctly

				got.Status(http.StatusOK)
				data := e2e.GetData(got).Object()
				payload := e2e.GetObject(data, "refreshToken")
				payload.Value("refreshToken").String()
				payload.Value("accessToken").String()
			},
			teardown: func(t *testing.T) {
				// Teardown : Cleanup user after test
				testutil.DropUser(t, client)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := tt.arrange(t)
			got := tt.act(t, refreshToken)
			tt.assert(t, got)
			tt.teardown(t)
		})
	}
}
