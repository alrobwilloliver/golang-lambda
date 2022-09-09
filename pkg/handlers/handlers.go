package handlers

import (
	"net/http"

	"github.com/alrobwilloliver/aws-lambda-in-golang/pkg/user"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var ErrorMethodNotAllowed = "Error Method Not Allowed"

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

func GetUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
	email := req.QueryStringParameters["email"]
	if len(email) > 0 {
		// Get single user
		result, err := user.FetchUser(email, tableName, dynaClient)
		if err != nil {
			return apiResponse(http.StatusInternalServerError, ErrorBody{aws.String(err.Error())})
		}
		return apiResponse(http.StatusOK, result)
	}

	// Get all users
	result, err := user.FetchAllUsers(tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, result)
}

func CreateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
	newUser, err := user.CreateUser(req, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusCreated, newUser)
}

func UpdateUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
	newUser, err := user.UpdateUser(req, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusInternalServerError, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, newUser)
}

func DeleteUser(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*events.APIGatewayProxyResponse, error) {
	err := user.DeleteUser(req, tableName, dynaClient)
	if err != nil {
		return apiResponse(http.StatusBadRequest, ErrorBody{aws.String(err.Error())})
	}
	return apiResponse(http.StatusOK, nil)
}

func UnhandledMethod() (*events.APIGatewayProxyResponse, error) {
	return apiResponse(http.StatusMethodNotAllowed, ErrorMethodNotAllowed)
}
