# WebSockets

This library allows you to bind multiple objects to a WebSocket, exposing some of their methods.

You could call it RPC over WebSockets.

## How It Works

The WebSocket protocol uses **channels** to manage multiple concurrent RPC sessions over a single WebSocket connection. Each channel represents an independent instance of a registered object.

### Message Flow

1. **Opening a channel**: Client sends a message with `type: "open"` (or `"new"`) to instantiate an object
2. **Calling methods**: Client sends messages with `path: "methodName"` to invoke methods on that instance
3. **Receiving replies**: Server sends messages with `path: "methodName"` containing return data
4. **Closing a channel**: Client sends `type: "close"` to destroy the instance

All messages include a `channel` field (arbitrary string ID) to route to the correct object instance.


Example:

```go
   // maps to `counter` unless overridden
   type Counter struct {
     MinIncrement int
     ct int
   }

   // exposes `inc` to the client, allowing for a int payload
   func (this *Counter) Inc(c ws.C, amt int) error {
     if amt < this.MinIncrement {
       amt = this.MinIncrement
     }
     this.ct+=amt
     return nil
   }

   // exposes `get` so the client can check the current value of the counter
   func (this *Counter) Get(c ws.C) error {
     return c.Reply("ct", this.ct)
   }
```   

Bound to an http server like:

```go
	h := ws.Handler{}
	h.Register(c, &Counter{MinIncrement: 1}) // the fields are shallow copied into any new instance
	s.Mux().Handle("/counter/ws/v1", h.Server()) // `h.Server()` returns an http.Handler
```

### Client Protocol

After connecting to the WebSocket, a client instantiates a new Counter by sending:

```json
{
  "channel": "c1",
  "type": "open",
  "path": "counter"
}
```

**Note**: The `path` field when opening must match the lowercase struct name (e.g., `Counter` → `"counter"`).

Increment it by 7 with:

```json
{
  "channel": "c1",
  "path": "inc",
  "data": 7
}
```

**Note**: Method names have their first letter lowercased (e.g., `Inc()` → `"inc"`, `IngestFile()` → `"ingestFile"`).

Query the current value:

```json
{
  "channel": "c1",
  "path": "get"
}
```

The server will reply with:

```json
{
  "channel": "c1",
  "path": "ct",
  "data": 7
}
```

Close the channel with:

```json
{
  "channel": "c1",
  "type": "close"
}
```

You can instantiate another `Counter` (or any other registered object) by using a different channel ID:

```json
{
  "channel": "c2",
  "type": "open",
  "path": "counter"
}
```

## Reflection

This library inspects the given object using reflection to determine what to expose.

### Fields

Any public field with a non-zero value is shallow-copied to new instances when a channel is opened. This allows you to configure shared settings (like `MinIncrement` in the example).

### `Init()` Method

The special method `Init(c ws.C, data T)` is called during channel instantiation but not exposed as a regular method. Use it for custom initialization logic when a channel opens.

The `data` parameter receives the `data` field from the opening message, allowing clients to pass initialization parameters.

### Exposed Methods

Any public method with signature `func (receiver *T) MethodName(c ws.C, [data T]) error` is exposed:

- **First parameter**: Must be `ws.C` (WebSocket context)
- **Second parameter** (optional): Receives the `data` field from the client message (unmarshaled to the parameter type)
- **Return**: Must return `error`

The `data` field from client messages is automatically unmarshaled into the method's second parameter type before calling.

### Using `ws.C`

The `ws.C` context provides:
- `c.Reply(path string, obj any)`: Send a reply message to the client
- `c.Close()`: Close the WebSocket connection
- All methods from `ctx.C` (logging, cancellation, etc.)

### Naming Convention

Names are derived automatically using `toLowerFirst()` which lowercases only the first character:
- **Struct name** → lowercase first letter (e.g., `Counter` → `"counter"`, `Board` → `"board"`)
- **Method name** → lowercase first letter, rest unchanged (e.g., `Inc()` → `"inc"`, `Get()` → `"get"`, `IngestFile()` → `"ingestFile"`)

**Important**: CamelCase method names preserve their internal capitalization. Only the first character is lowercased.


## Test

Since all the bindings to the WebSocket is automatic, you could test the object directly for functionalities and ignore the bindings entirely

Alternatively you can use the provided test client:

```go
	h := ws.Handler{}
	_ = h.Register(c, &Example{})
	cli := h.NewTest(t)

	send, err := cli.Open(c, "example", nil, func(c ctx.C, f ws.Frame) error {
		switch f.Type {
		case "close":
			t.Logf("recv CLOSED")
		default:
			t.Logf("recv %+v", f.Data)
		}
		return nil
	})
```

## `Handler{}.Server()`

calling `.Server()` on a handler, will create an `http.Handler` which can be used on any http server.

Internally it uses `golang.org/x/net/websocket` which implements `http.Handler`

you might need to modify the `.Handshake` function pointer, by default it accept from the same origin
