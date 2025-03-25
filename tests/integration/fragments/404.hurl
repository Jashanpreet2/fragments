# Get a non-existent root
GET http://localhost:8080/NonExistentRoute

# Ensure it returns 404
HTTP/1.1 404