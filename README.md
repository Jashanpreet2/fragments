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
1. Install air using "go install github.com/air-verse/air@latest"
2. air init
3. air

###### Generate coverage report
1. go test -coverprofile="c.out"
2. go tool cover -html="c.out"
