# tests/integration/post-fragments.hurl
# Authenticated POST to /v1/fragments
POST http://localhost:8080/v1/fragments
# We're sending a plain text fragment
Content-Type: text/plain
# Include HTTP Basic Auth credentials
[BasicAuth]
user1@email.com:password1
# Body of the request goes in ```...``` when it's a string
{ "service": "DynamoDB" }

HTTP/1.1 201
[Captures]
fragment1_url: header "Location"
fragment1_id: jsonpath "$.fragment.id"

# 2. Try to GET the fragment we just posted by its URL
GET {{fragment1_url/info}}
[BasicAuth]
user1@email.com:password1

HTTP/1.1 200
[Asserts]
header "Location" matches "^http:\/\/localhost:8080\/v1\/fragment\/[A-Za-z0-9_-]+$"
jsonpath "$.status" == "ok"
jsonpath "$.fragment.id" matches /\d+/
jsonpath "$.fragment.ownerId" matches "^[0-9a-fA-F]+$"
jsonpath "$.fragment.created" isString
jsonpath "$.fragment.updated" isString
jsonpath "$.fragment.fragmentType" == "text/json"
jsonpath "$.fragment.size" == 21

