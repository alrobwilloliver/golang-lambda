# Overview
This is a quick simple lambda project from a tutorial I found online. I can do CRUD operations on users using Lambda functions and saved in a DynamoDB table. I also used this as an opportunity to mock and test Lamdba functions.

# Lambda URLs

https://4uirkr2nw9.execute-api.eu-west-2.amazonaws.com/staging

## Endpoints

### GET By Email
```bash
curl -X GET https://4uirkr2nw9.execute-api.eu-west-2.amazonaws.com/staging\?email=$EMAIL
```

### GET All

```bash
curl -X GET https://4uirkr2nw9.execute-api.eu-west-2.amazonaws.com/staging
```

### POST

```bash
curl --header "Content-Type: application/json" --request POST --data '{"email": "alan.oliver@ecs.co.uk", "firstName": "Al", "lastName": "Oliver"}' https://4uirkr2nw9.execute-api.eu-west-2.amazonaws.com/staging
```

### UPDATE
```bash
curl --header "Content-Type: application/json" --request PUT --data '{"email": "alan.oliver@ecs.co.uk", "firstName": "Alan", "lastName": "Oliver"}' https://4uirkr2nw9.execute-api.eu-west-2.amazonaws.com/staging
```

### DELETE
```bash
curl -X DELETE https://4uirkr2nw9.execute-api.eu-west-2.amazonaws.com/staging\?email\=alan.oliver@ecs.co.uk 
```