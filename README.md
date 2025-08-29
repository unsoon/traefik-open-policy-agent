# Traefik Open Policy Agent Plugin <img src="https://raw.githubusercontent.com/open-policy-agent/opa/main/logo/logo.png" align="right" width="150" height="150" style="margin: 0px 0px 10px 10px" >

A Traefik middleware plugin that integrates with Open Policy Agent (OPA) for request authorization. This plugin allows you to implement flexible and powerful authorization policies using OPA's policy language (Rego).

## Features

- Integrates Traefik with Open Policy Agent
- Validates requests against OPA policies
- Supports full request context (headers, path, method, query parameters)
- Customizable error responses with support for:
  - Custom HTTP status code
  - Custom headers
  - Custom response body
  - Multiple content types (JSON/Plain text)

## Configuration

### Static Configuration

To enable the plugin in your Traefik instance:

```yaml
experimental:
  plugins:
    open-policy-agent:
      moduleName: "github.com/unsoon/traefik-open-policy-agent"
      version: "v1.2.1"
```

### Dynamic Configuration

```yaml
http:
  middlewares:
    my-opa-middleware:
      plugin:
        open-policy-agent:
          url: "http://opa.kube-system:8181/v1/data/httpapi/authz"
          allowField: "allow"
          errorResponse:
            statusCode: 403
            contentType: "application/json"
            headers:
              X-Error-Type: "authorization_failed"
            body:
              error: "Access denied by policy"
```

### Configuration Options

| Option                      | Type                | Required | Default            | Description                                                         |
| --------------------------- | ------------------- | -------- | ------------------ | ------------------------------------------------------------------- |
| `url`                       | `string`            | Yes      | -                  | OPA server URL with policy path                                     |
| `allowField`                | `string`            | No       | `allow`            | Field name in OPA response for allow/deny                           |
| `errorResponse.statusCode`  | `int`               | No       | `403`              | HTTP status code for denied requests                                |
| `errorResponse.contentType` | `string`            | No       | `application/json` | Content type of error response (`text/plain` or `application/json`) |
| `errorResponse.headers`     | `map[string]string` | No       | `{}`               | Additional headers to include in error response                     |
| `errorResponse.body`        | `interface{}`       | No       | `nil`              | Custom response body                                                |

## How It Works

1. The plugin intercepts incoming HTTP requests
2. Sends request data to OPA server including:
   - HTTP method
   - Request path
   - Headers
   - Query parameters
3. OPA evaluates the request against defined policies
4. Based on OPA's response:
   - If allowed: request proceeds to the next middleware/handler
   - If denied: returns configured error response

## Example OPA Policy

Here's a simple example of an OPA policy that allows requests based on specific criteria:

```rego
package httpapi.authz

import data.io.jwt

default allow = false

env := opa.runtime().env

allow if {
    [token] := input.headers["Authorization"]
    substring(token, 0, 6, prefix)

    prefix == "Basic "
    substring(token, 6, -1, basic_token)

    base64url.decode(basic_token, decoded)
    split(decoded, ":", [username, password])

    crypto.hmac.equal(username, env.USERNAME)
    crypto.hmac.equal(crypto.md5(password), env.HASHED_PASSWORD)
}

allow if {
    [token] := input.headers["Authorization"]
    substring(token, 0, 7, prefix)

    prefix == "Bearer "
    substring(token, 7, -1, bearer_token)

    io.jwt.verify_rs256(bearer_token, env.JWKS_URL)
}
```

## Example Usage

### Basic Authorization Check

```yaml
http:
  middlewares:
    api-auth:
      plugin:
        open-policy-agent:
          url: "http://opa.kube-system:8181/v1/data/httpapi/authz"
          allowField: "allow"
```

### Custom Error Response

```yaml
http:
  middlewares:
    secure-api:
      plugin:
        open-policy-agent:
          url: "http://opa.kube-system:8181/v1/data/httpapi/authz"
          allowField: "allow"
          errorResponse:
            statusCode: 401
            contentType: "application/json"
            headers:
              WWW-Authenticate: 'Bearer realm="example"'
            body:
              message: "Authorization required"
              details: "Please provide valid credentials"
```

## Request Data Format

The plugin sends the following data structure to OPA:

```json
{
  "input": {
    "method": "POST",
    "path": ["api", "posts"],
    "headers": {
      "authorization": ["Bearer token"],
      "content-type": ["application/json"]
    },
    "query": {
      "filter": ["active"],
      "sort": ["desc"]
    },
    "body": {
      "title": "New Post",
      "author": "john.doe",
      "content": "Hello World!"
    }
  }
}
```

### Body-Based Authorization Example

Here's an example of using request body content for authorization decisions:

```rego
package httpapi.authz

# Default deny
default allow = false

# Allow users to modify only their own posts
allow if {
    # Check if this is a POST or PUT request
    input.method in ["POST", "PUT"]

    # Extract user from JWT token
    [token] := input.headers["Authorization"]
    
    # Check if the token is a Bearer token
    startswith(token, "Bearer ")

    # Extract the JWT token
    substring(token, 7, -1, jwt)

    # Decode the JWT token
    io.jwt.decode(jwt, [_, payload, _])

    # Extract the username from the JWT token
    username := payload.sub

    # Verify the username is not empty
    username

    # Extract the author from the request body
    author := input.body.author

    # Verify the author is not empty
    author

    # Verify the author in request body matches the authenticated user's username
    crypto.hmac.equal(author, username)
}
```

Configure the middleware to use this policy:

```yaml
http:
  middlewares:
    author-check:
      plugin:
        open-policy-agent:
          url: "http://opa.kube-system:8181/v1/data/httpapi/authz"
          errorResponse:
            statusCode: 403
            contentType: "application/json"
            body:
              error: "You can only create/modify your own posts"
```

This policy ensures users can only create or modify posts where they are the author.

## License

This plugin is distributed under the [MIT License](LICENSE).
