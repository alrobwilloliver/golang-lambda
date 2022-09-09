package user

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
	deleteErr   error
	fetchedUser *dynamodb.GetItemOutput
	fetchErr    error
	putErr      error
	scanRes     *dynamodb.ScanOutput
	scanErr     error
}

func (m *mockDynamoDBClient) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	return m.fetchedUser, m.fetchErr
}

func (m *mockDynamoDBClient) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, m.putErr
}

func (m *mockDynamoDBClient) Scan(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return m.scanRes, m.scanErr
}

func (m *mockDynamoDBClient) DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	return nil, m.deleteErr
}

func TestCreateUser(t *testing.T) {
	t.Run("expect error when invalid body is provided", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}

		_, err := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": , "name": ""}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorInvalidUserData {
			t.Errorf("Expected error %s, got %s", ErrorInvalidUserData, err.Error())
		}
	})
	t.Run("expect error when invalid email is provided", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}

		_, err := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "invalid-email", "firstName": "test", "lastName": "test"}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorInvalidEmail {
			t.Errorf("Expected error %s, got %s", ErrorInvalidEmail, err.Error())
		}
	})
	t.Run("expect error when fetching user to see if it already exists fails", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.fetchErr = errors.New("test error")

		_, err := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.shearer@ecs.co.uk", "firstName": "Alan", "lastName": "Shearer"}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorFailedToFetchRecord {
			t.Errorf("Expected error %s, got %s", ErrorFailedToFetchRecord, err.Error())
		}
	})
	t.Run("expect error when user already exists", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.fetchedUser = &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{
				"email": {
					S: aws.String("alan.shearer@ecs.co.uk"),
				},
			},
		}

		_, err := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.shearer@ecs.co.uk", "firstName": "Alan", "lastName": "Shearer"}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorUserAlreadyExists {
			t.Errorf("Expected error %s, got %s", ErrorUserAlreadyExists, err.Error())
		}
	})
	t.Run("expect error when creating user fails", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.fetchedUser = &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{},
		}
		mockDb.putErr = errors.New("test error")

		_, err := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstName": "Alan", "lastName": "Oliver"}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorCouldNotDynamoPutItem {
			t.Errorf("Expected error %s, got %s", ErrorCouldNotDynamoPutItem, err.Error())
		}
	})
	t.Run("expect user to be created", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.fetchedUser = &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{},
		}

		createdUser, err := CreateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstName": "Alan", "lastName": "Oliver"}`,
		}, "test", mockDb)
		if err != nil {
			t.Fatalf("Expected nil, got %s", err.Error())
		}
		if createdUser.Email != "alan.oliver@ecs.co.uk" {
			t.Errorf("Expected email %s, got %s", "alan.oliver@ecs.co.uk", createdUser.Email)
		}
		if createdUser.FirstName != "Alan" {
			t.Errorf("Expected firstName %s, got %s", "Alan", createdUser.FirstName)
		}
		if createdUser.LastName != "Oliver" {
			t.Errorf("Expected lastName %s, got %s", "Oliver", createdUser.LastName)
		}
	})
}

func TestFetchAllUsers(t *testing.T) {
	t.Run("expect error when fetching users fails", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.scanErr = errors.New("scan error")

		_, err := FetchAllUsers("test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorFailedToFetchRecord {
			t.Errorf("Expected error %s, got %s", ErrorFailedToFetchRecord, err.Error())
		}
	})
	t.Run("should return empty list when no users are found", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.scanRes = &dynamodb.ScanOutput{
			Items: []map[string]*dynamodb.AttributeValue{},
		}

		users, err := FetchAllUsers("test", mockDb)
		if err != nil {
			t.Fatalf("Expected nil, got %s", err.Error())
		}
		if len(*users) != 0 {
			t.Errorf("Expected length %d, got %d", 0, len(*users))
		}
	})
	t.Run("should return list of users", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.scanRes = &dynamodb.ScanOutput{
			Items: []map[string]*dynamodb.AttributeValue{
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
				{
					"email": {
						S: aws.String("alan.oliver@ecs.co.uk"),
					},
					"firstName": {
						S: aws.String("Al"),
					},
					"lastName": {
						S: aws.String("Oliver"),
					},
				},
			},
		}
		users, err := FetchAllUsers("test", mockDb)
		if err != nil {
			t.Fatalf("Expected nil, got %s", err.Error())
		}
		if len(*users) != 2 {
			t.Errorf("Expected length %d, got %d", 2, len(*users))
		}
		if (*users)[0].Email != "alan.shearer@ecs.co.uk" {
			t.Errorf("Expected email %s, got %s", "alan.shearer@ecs.co.uk", (*users)[0].Email)
		}
		if (*users)[0].FirstName != "Alan" {
			t.Errorf("Expected firstName %s, got %s", "Alan", (*users)[0].FirstName)
		}
		if (*users)[0].LastName != "Shearer" {
			t.Errorf("Expected lastName %s, got %s", "Shearer", (*users)[0].LastName)
		}
		if (*users)[1].Email != "alan.oliver@ecs.co.uk" {
			t.Errorf("Expected email %s, got %s", "alan.oliver@ecs.co.uk", (*users)[1].Email)
		}
		if (*users)[1].FirstName != "Al" {
			t.Errorf("Expected firstName %s, got %s", "Al", (*users)[1].FirstName)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Run("expect error when request body is invalid", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}

		_, err := UpdateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstNam": , "lastName": "Oliver"}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorInvalidUserData {
			t.Errorf("Expected error %s, got %s", ErrorInvalidUserData, err.Error())
		}
	})
	t.Run("expect error when there is an error fetching the user", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.fetchErr = errors.New("fetch error")

		_, err := UpdateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstName": "Allen", "lastName": "Oliver"}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorFailedToFetchRecord {
			t.Errorf("Expected error %s, got %s", ErrorFailedToFetchRecord, err.Error())
		}
	})
	t.Run("expect error when there is an error updating the user", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.fetchedUser = &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{
				"email": {
					S: aws.String("alan.oliver@ecs.co.uk"),
				},
				"firstName": {
					S: aws.String("Al"),
				},
				"lastName": {
					S: aws.String("Oliver"),
				},
			},
		}
		mockDb.putErr = errors.New("update error")

		_, err := UpdateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstName": "Allen", "lastName": "Oliver"}`,
		}, "test", mockDb)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorCouldNotDynamoPutItem {
			t.Errorf("Expected error %s, got %s", ErrorCouldNotDynamoPutItem, err.Error())
		}
	})
	t.Run("expect to update the user", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.fetchedUser = &dynamodb.GetItemOutput{
			Item: map[string]*dynamodb.AttributeValue{
				"email": {
					S: aws.String("alan.oliver@ecs.co.uk"),
				},
				"firstName": {
					S: aws.String("Al"),
				},
				"lastName": {
					S: aws.String("Oliver"),
				},
			},
		}

		response, err := UpdateUser(events.APIGatewayProxyRequest{
			Body: `{"email": "alan.oliver@ecs.co.uk", "firstName": "Allen", "lastName": "Oliver"}`,
		}, "test", mockDb)
		if err != nil {
			t.Fatalf("Expected nil, got %s", err.Error())
		}
		if response.Email != "alan.oliver@ecs.co.uk" {
			t.Errorf("Expected email %s, got %s", "alan.oliver@ecs.co.uk", response.Email)
		}
		if response.FirstName != "Allen" {
			t.Errorf("Expected firstName %s, got %s", "Allen", response.FirstName)
		}
		if response.LastName != "Oliver" {
			t.Errorf("Expected lastName %s, got %s", "Oliver", response.LastName)
		}
	})
}

func TestDeleteUser(t *testing.T) {
	t.Run("expect error when there is an error deleting the user", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}
		mockDb.deleteErr = errors.New("delete error")

		err := DeleteUser(events.APIGatewayProxyRequest{}, "test", mockDb)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if err.Error() != ErrorFailedToDeleteRecord {
			t.Errorf("Expected error %s, got %s", ErrorFailedToDeleteRecord, err.Error())
		}
	})
	t.Run("expect to delete the user", func(t *testing.T) {
		mockDb := &mockDynamoDBClient{}

		err := DeleteUser(events.APIGatewayProxyRequest{
			PathParameters: map[string]string{
				"email": "alan.oliver@ecs.co.uk",
			},
		}, "test", mockDb)

		if err != nil {
			t.Fatalf("Expected nil, got %s", err.Error())
		}
	})
}
