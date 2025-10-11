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

func TestUser_CreateUser(t *testing.T) {
	expect, client, teardown := e2e.Setup(t, e2e.SetupOption{
		TearDown: func(t *testing.T, client *ent.Client) {
			testutil.DropUser(t, client)
		},
	})
	defer teardown()

	tests := []struct {
		name    string
		arrange func(t *testing.T)
		act     func(t *testing.T) *httpexpect.Response
		assert  func(t *testing.T, got *httpexpect.Response)
		args    struct {
			ctx context.Context
		}
		teardown func(t *testing.T)
	}{
		{
			name:    "It Should create test user",
			arrange: func(t *testing.T) {},
			act: func(t *testing.T) *httpexpect.Response {
				return expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": `
                mutation {
        createUser(input :{name:"Bhuwan", 
          email:"test@mail.com",
          password:"sometestpassword",
            age:34
          }
          )
        {
        id
        name
        age
        createdAt
        updatedAt

        }
        }
        `,
				}).Expect()
			},
			assert: func(t *testing.T, got *httpexpect.Response) {
				got.Status(http.StatusOK)
				data := e2e.GetData(got).Object()
				testUser := e2e.GetObject(data, "createUser")
				testUser.Value("name").String().IsEqual("Bhuwan")
				testUser.Value("age").Number().IsEqual(34)
			},
			teardown: func(t *testing.T) {
				testutil.DropUser(t, client)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.arrange(t)
			got := tt.act(t)
			tt.assert(t, got)
			tt.teardown(t)
		})
	}
}

func TestUser_Updateuser(t *testing.T) {
	expect, client, teardown := e2e.Setup(t, e2e.SetupOption{
		TearDown: func(t *testing.T, client *ent.Client) {
			testutil.DropUser(t, client)
		},
	})

	defer teardown()

	tests := []struct {
		name    string
		arrange func(t *testing.T) (string, string)
		act     func(t *testing.T, userId string, accessToken string) *httpexpect.Response
		assert  func(t *testing.T, got *httpexpect.Response)
		args    struct {
			ctx context.Context
		}
		teardown func(t *testing.T)
	}{
		{
			name: "It should update test user",
			arrange: func(t *testing.T) (string, string) {
				// Create a user and return its id
				resp := expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": `

          mutation {
        createUser (input:{
        name :"Bhuwan",
          email:"test@mail.com",
          password:"sometestpassword"
        age : 34
        }) {
        id
        name
          email
        age
        createdAt
        updatedAt
        }
        }
        `,
				}).Expect().Status(http.StatusOK)
				// Extract the ID of the created user
				data := e2e.GetData(resp).Object()
				createdUser := e2e.GetObject(data, "createUser")
				userId := createdUser.Value("id").String().Raw()

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
				accessToken := loggedInUser.Value("accessToken").String().Raw()

				return string(userId), string(accessToken)
			},
			act: func(t *testing.T, userId string, accessToken string) *httpexpect.Response {
				// Construct the query string with the userID
				query := fmt.Sprintf(`
					mutation {
						updateUser(input: { id: "%s", name: "Nanda", age: 35 }) {
							id
							name
							age
							updatedAt
						}
					}
				`, userId)
				return expect.POST(router.QueryPath).
					WithHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
					WithJSON(map[string]string{
						"query": query,
					}).Expect()
			},
			assert: func(t *testing.T, got *httpexpect.Response) {
				// Assert: Check if the user's details were updated correctly

				got.Status(http.StatusOK)
				data := e2e.GetData(got).Object()
				updatedUser := e2e.GetObject(data, "updateUser")
				updatedUser.Value("age").Number().IsEqual(35)
				updatedUser.Value("name").String().IsEqual("Nanda")
			},
			teardown: func(t *testing.T) {
				// Teardown : Cleanup user after test
				testutil.DropUser(t, client)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userId, accessToken := tt.arrange(t)
			got := tt.act(t, userId, accessToken)
			tt.assert(t, got)
			tt.teardown(t)
		})
	}
}
