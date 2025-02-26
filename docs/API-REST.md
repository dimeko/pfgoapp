# [GoApp REST API](#goapp)

## GET /goapp

Returns an example websocket page.

## GET /goapp/ws

The message sent by the server containing the counter value:
```
{"iteration": 1, "value": "0ABDE3F5EB}
```
The message sent by the client to reset the counter:
```
{}
```

If the application has been initialised with csrf protection then this endpoint is changed to `/goapp/ws/{csrf_token}` and requests to `/goapp/ws` return `404 not found`.

## [GET /goapp/health](#health)
| _health_ |

Returns an HTTP code for health status.
