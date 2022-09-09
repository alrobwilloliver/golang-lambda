package handlers

import (
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type mockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
	fetchUser *dynamodb.GetItemOutput
	fetchErr  error
	scanRes   *dynamodb.ScanOutput
	scanErr   error
}

func (m mockDynamoDBClient) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return m.fetchUser, m.fetchErr
}

func (m mockDynamoDBClient) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func (m mockDynamoDBClient) Scan(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return m.scanRes, m.scanErr
}

func TestGetUser(t *testing.T) {
	t.Run("should return a 500 response when failure to fetch record", func(t *testing.T) {
		mockDb := mockDynamoDBClient{
			fetchErr: errors.New("user not found"),
		}
		resp, _ := GetUser(events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"email": "alan.oliver@ecs.co.uk",
			},
		}, "test", mockDb)
		if resp.StatusCode != 500 {
			t.Errorf("expected status code to be %d, got %d", 500, resp.StatusCode)
		}
		if resp.Body != "{\"error\":\"failed to fetch record\"}" {
			t.Errorf("expected body to be %q, got %q", "{\"error\":\"failed to fetch record\"}", resp.Body)
		}
	})
	t.Run("should return a user", func(t *testing.T) {
		mockDb := mockDynamoDBClient{
			fetchUser: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"email": {
						S: aws.String("alan.oliver@ecs.co.uk"),
					},
					"firstName": {
						S: aws.String("Alan"),
					},
					"lastName": {
						S: aws.String("Oliver"),
					},
				},
			},
		}
		resp, _ := GetUser(events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"email": "alan.oliver@ecs.co.uk",
			},
		}, "test", mockDb)
		if resp.StatusCode != 200 {
			t.Errorf("expected status code to be %d, got %d", 200, resp.StatusCode)
		}
		if resp.Body != "{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Oliver\"}" {
			t.Errorf("expected body to be %q, got %q", "{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Oliver\"}", resp.Body)
		}
	})
	t.Run("should fail to find all users when no users are found", func(t *testing.T) {
		mockDb := mockDynamoDBClient{
			scanErr: errors.New("no users found"),
		}
		resp, _ := GetUser(events.APIGatewayProxyRequest{}, "test", mockDb)
		if resp.StatusCode != 500 {
			t.Errorf("expected status code to be %d, got %d", 500, resp.StatusCode)
		}
		if resp.Body != "{\"error\":\"failed to fetch record\"}" {
			t.Errorf("expected body to be %q, got %q", "{\"error\":\"failed to fetch record\"}", resp.Body)
		}
	})
	t.Run("should return all users", func(t *testing.T) {
		mockDb := mockDynamoDBClient{
			scanRes: &dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{
						"email": {
							S: aws.String("alan.oliver@ecs.co.uk"),
						},
						"firstName": {
							S: aws.String("Alan"),
						},
						"lastName": {
							S: aws.String("Oliver"),
						},
					},
					{
						"email": {
							S: aws.String("alan.shearer@ecs.co.uk"),
						},
						"firstName": {
							S: aws.String("Alan"),
						},
						"lastName": {
							S: aws.String("Shearer"),
						},
					},
				},
			},
		}
		resp, _ := GetUser(events.APIGatewayProxyRequest{}, "test", mockDb)
		if resp.StatusCode != 200 {
			t.Errorf("expected status code to be %d, got %d", 200, resp.StatusCode)
		}
		if resp.Body != "[{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Oliver\"},{\"email\":\"alan.shearer@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Shearer\"}]" {
			t.Errorf("expected body to be %q, got %q", "[[{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Oliver\"},{\"email\":\"alan.shearer@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Shearer\"}]", resp.Body)
		}
	})
}

func TestCreateUser(t *testing.T) {
	t.Run("should return a 500 error response when the request body is invalid", func(t *testing.T) {
		resp, _ := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "1"`,
		}, "test", nil)

		if resp == nil {
			t.Fatalf("expected a response, got nil")
		}
		if resp.StatusCode != 500 {
			t.Fatalf("expected status code 500, got %d", resp.StatusCode)
		}
		if resp.Body != "{\"error\":\"invalid user data\"}" {
			t.Fatalf("expected body to be %q, got %q", "{\"error\":\"invalid user data\"}", resp.Body)
		}
		if resp.Headers["Application-Type"] != "application/json" {
			t.Fatalf("expected header to be %q, got %q", "application/json", resp.Headers["Application-Type"])
		}
	})
	t.Run("should return a 201 response when the request body is valid", func(t *testing.T) {
		mockDb := mockDynamoDBClient{
			fetchUser: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{},
			},
		}
		resp, _ := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstName": "Alan", "lastName": "Oliver"}`,
		}, "test", mockDb)

		if resp == nil {
			t.Fatalf("expected a response, got nil")
		}
		if resp.StatusCode != 201 {
			t.Fatalf("expected status code 201, got %d", resp.StatusCode)
		}
		if resp.Body != "{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Oliver\"}" {
			t.Fatalf("expected body to be %q, got %q", "{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Oliver\"}", resp.Body)
		}
		if resp.Headers["Application-Type"] != "application/json" {
			t.Fatalf("expected header to be %q, got %q", "application/json", resp.Headers["Application-Type"])
		}
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("should return a 500 error response when the request body is invalid", func(t *testing.T) {
		resp, _ := UpdateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "1"`,
		}, "test", nil)

		if resp == nil {
			t.Fatalf("expected a response, got nil")
		}
		if resp.StatusCode != 500 {
			t.Fatalf("expected status code 500, got %d", resp.StatusCode)
		}
		if resp.Body != "{\"error\":\"invalid user data\"}" {
			t.Fatalf("expected body to be %q, got %q", "{\"error\":\"invalid user data\"}", resp.Body)
		}
		if resp.Headers["Application-Type"] != "application/json" {
			t.Fatalf("expected header to be %q, got %q", "application/json", resp.Headers["Application-Type"])
		}
	})
	t.Run("should return a 200 response when the request body is valid", func(t *testing.T) {
		mockDb := mockDynamoDBClient{
			fetchUser: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"email": {
						S: aws.String("alan.oliver@ecs.co.uk"),
					},
					"firstName": {
						S: aws.String("Alan"),
					},
					"lastName": {
						S: aws.String("Oliver"),
					},
				},
			},
		}

		resp, _ := UpdateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstName": "Al", "lastName": "O"}`,
		}, "test", mockDb)

		if resp == nil {
			t.Fatalf("expected a response, got nil")
		}
		if resp.StatusCode != 200 {
			t.Fatalf("expected status code 200, got %d", resp.StatusCode)
		}
		if resp.Body != "{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Al\",\"lastName\":\"O\"}" {
			t.Fatalf("expected body to be %q, got %q", "{\"email\":\"alan.oliver@ecs.co.uk\",\"firstName\":\"Alan\",\"lastName\":\"Oliver\"}", resp.Body)
		}
		if resp.Headers["Application-Type"] != "application/json" {
			t.Fatalf("expected header to be %q, got %q", "application/json", resp.Headers["Application-Type"])
		}
	})
}
