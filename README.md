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
      version: "v1.0.0"
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

allow {
    token := input.headers["Authorization"][0]
    prefix := substring(token, 0, 6)

    prefix == "Basic "
    basic_token := substring(token, 6, -1)

    decoded := base64url.decode(basic_token)
    [username, password] := split(decoded, ":")

    username == env.USERNAME
    crypto.md5(password) == env.HASHED_PASSWORD
}

allow {
    token := input.headers["Authorization"][0]
    prefix := substring(token, 0, 7)

    prefix == "Bearer "
    bearer_token := substring(token, 7, -1)

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
          allowField: "allowed"
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
    "method": "GET",
    "path": ["api", "posts", "7903c777-cbee-4b9e-b886-d9c4ffebd8a3"],
    "headers": {
      "authorization": ["Bearer token"],
      "content-type": ["application/json"]
    },
    "query": {
      "filter": ["active"],
      "sort": ["desc"]
    }
  }
}
```

## License

This plugin is distributed under the [MIT License](LICENSE).
