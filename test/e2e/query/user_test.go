package query_test

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

func TestUser_GetUser(t *testing.T) {
	expect, client, teardown := e2e.Setup(t, e2e.SetupOption{
		TearDown: func(t *testing.T, client *ent.Client) {
			testutil.DropUser(t, client)
		},
	})
	defer teardown()

	tests := []struct {
		name    string
		arrange func(t *testing.T) (string, string)                                        // Return the ID of the created user as a string
		act     func(t *testing.T, userID string, accessToken string) *httpexpect.Response // Accept the user ID as a string
		assert  func(t *testing.T, got *httpexpect.Response, userID string)
		args    struct {
			ctx context.Context
		}
		teardown func(t *testing.T)
	}{
		{
			name: "It Should get a single user by ID",
			arrange: func(t *testing.T) (string, string) {
				// Arrange: Create a user and return its ID
				resp := expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": `
						mutation {
          createUser(input: {name: "Bhuwan", email:"test@mail.com",password:"sometestpass",age: 34}) {
								id
								name
								age
								createdAt
								updatedAt
							}
						}
					`,
				}).Expect().Status(http.StatusOK)

				// Extract the ID of the created user as a string
				data := e2e.GetData(resp).Object()
				createdUser := e2e.GetObject(data, "createUser")
				userID := createdUser.Value("id").String().Raw() // Extract as string

				loginResponse := expect.POST(router.QueryPath).WithJSON(map[string]string{
					"query": `

          mutation {
        login (input:{
          email:"test@mail.com",
          password:"sometestpass"
        
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
				return userID, accessToken
			},
			act: func(t *testing.T, userID string, accessToken string) *httpexpect.Response {
				// Construct the query string with the userID
				query := fmt.Sprintf(`
					query {
						user(id: "%s") {
							id
							name
							age
							createdAt
							updatedAt
						}
					}
				`, userID) // Use %s for string and wrap userID in quotes

				// Debug: Print the query to verify it
				fmt.Println("Generated Query:", query)

				// Act: Query the user's details using the ID from arrange
				return expect.POST(router.QueryPath).
					WithHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
					WithJSON(map[string]string{
						"query": query,
					}).Expect()
			},
			assert: func(t *testing.T, got *httpexpect.Response, userID string) {
				// Assert: Check if the user's details were retrieved correctly
				got.Status(http.StatusOK)
				data := e2e.GetData(got).Object()
				user := e2e.GetObject(data, "user")
				user.Value("id").String().IsEqual(userID) // Ensure the ID matches
				user.Value("name").String().IsEqual("Bhuwan")
				user.Value("age").Number().IsEqual(34)
			},
			teardown: func(t *testing.T) {
				// Teardown: Clean up the user after the test
				testutil.DropUser(t, client)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, accessToken := tt.arrange(t)  // Arrange and get the user ID
			got := tt.act(t, userID, accessToken) // Act using the user ID
			tt.assert(t, got, userID)             // Assert
			tt.teardown(t)                        // Teardown
		})
	}
}

func TestUser_ListUser(t *testing.T) {
	expect, client, teardown := e2e.Setup(t, e2e.SetupOption{
		TearDown: func(t *testing.T, client *ent.Client) {
			testutil.DropUser(t, client)
		},
	})
	defer teardown()

	tests := []struct {
		name    string
		arrange func(t *testing.T) string                                   // Create multiple users
		act     func(t *testing.T, accessToken string) *httpexpect.Response // Query the list of users
		assert  func(t *testing.T, got *httpexpect.Response)
		args    struct {
			ctx context.Context
		}
		teardown func(t *testing.T)
	}{
		{
			name: "It Should get a list of users using Relay Connection Specification",
			arrange: func(t *testing.T) string {
				// Arrange: Create 3 random users
				users := []struct {
					name     string
					email    string
					password string
					age      int
				}{
					{name: "Alice", age: 25, email: "test@mail.com", password: "sometestpassword"},
					{name: "Bob", age: 30, email: "test1@mail.com", password: "somestssdfsdf"},
					{name: "Charlie", age: 35, email: "test2@mail.com", password: "someststssdf"},
				}

				for _, user := range users {
					resp := expect.POST(router.QueryPath).WithJSON(map[string]string{
						"query": fmt.Sprintf(`
							mutation {
              createUser(input: {name: "%s", age: %d, email:"%s",password:"%s"}) {
									id
									name
									age
									createdAt
									updatedAt
								}
							}
						`, user.name, user.age, user.email, user.password),
					}).Expect().Status(http.StatusOK)

					// Debug: Print the created user
					data := e2e.GetData(resp).Object()
					createdUser := e2e.GetObject(data, "createUser")
					fmt.Printf(
						"Created User: %s (ID: %s)\n",
						createdUser.Value("name").String().Raw(),
						createdUser.Value("id").String().Raw(),
					)
				}

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
				return accessToken
			},
			act: func(t *testing.T, accessToken string) *httpexpect.Response {
				// Act: Query the list of users using Relay Connection Specification
				query := `
					query {
						users(first: 3) {
							edges {
								node {
									id
									name
									age
									createdAt
									updatedAt
								}
							}
							pageInfo {
								hasNextPage
								hasPreviousPage
								startCursor
								endCursor
							}
						}
					}
				`

				// Debug: Print the query
				fmt.Println("Generated Query:", query)

				return expect.POST(router.QueryPath).
					WithHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
					WithJSON(map[string]string{
						"query": query,
					}).Expect()
			},
			assert: func(t *testing.T, got *httpexpect.Response) {
				// Assert: Check if the list of users is returned correctly
				got.Status(http.StatusOK)

				// Validate the response structure
				data := e2e.GetData(got).Object()
				usersConnection := e2e.GetObject(data, "users")
				edges := usersConnection.Value("edges").Array()

				// Ensure 3 users are returned
				edges.Length().IsEqual(3)

				// Validate each user in the list
				expectedUsers := []struct {
					name string
					age  int
				}{
					{name: "Alice", age: 25},
					{name: "Bob", age: 30},
					{name: "Charlie", age: 35},
				}

				for i, edge := range edges.Iter() {
					node := edge.Object().Value("node").Object()
					node.Value("name").String().IsEqual(expectedUsers[i].name)
					node.Value("age").Number().IsEqual(expectedUsers[i].age)
				}

				// Validate pageInfo
				pageInfo := usersConnection.Value("pageInfo").Object()
				pageInfo.Value("hasNextPage").Boolean().IsFalse() // Assuming only 3 users exist
				pageInfo.Value("hasPreviousPage").Boolean().IsFalse()
				pageInfo.Value("startCursor").String().NotEmpty()
				pageInfo.Value("endCursor").String().NotEmpty()
			},
			teardown: func(t *testing.T) {
				// Teardown: Clean up the users after the test
				testutil.DropUser(t, client)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken := tt.arrange(t)  // Arrange: Create users
			got := tt.act(t, accessToken) // Act: Query the list of users
			tt.assert(t, got)             // Assert: Validate the response
			tt.teardown(t)                // Teardown: Clean up
		})
	}
}
