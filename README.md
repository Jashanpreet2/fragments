# fragments

CCP555 Labs

## Resources

Gin Documentation (server framework): https://gin-gonic.com/docs/
Zap Logger Documentation: https://pkg.go.dev/go.uber.org/zap
cognitoJwtVerify Documentation (used to verify cognito tokens): https://pkg.go.dev/github.com/jhosan7/cognito-jwt-verify#section-readme

## Scripts

###### Start the server
go run server.go

###### Start server with hot reloading
1. air init
2. air

###### Generate coverage report
1. go test -coverprofile="c.out"
2. go tool cover -html="c.out"