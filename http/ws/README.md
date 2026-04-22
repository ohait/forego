# WebSockets

This library binds objects to a WebSocket and exposes selected methods.

You can think of it as RPC over WebSockets.

## How It Works

The protocol uses **channels** to multiplex RPC sessions over one WebSocket connection. Each channel is an independent instance of a registered object. The same object type can be instantiated multiple times using different channel IDs.

### Message Flow

1. **Open a channel**: Client sends `type: "open"` (or `"new"`) to instantiate an object
2. **Call methods**: Client sends messages with `path: "methodName"` to invoke methods on that instance
3. **Receive replies**: Server sends messages with `path: "methodName"` containing data
4. **Cancel a request**: Client may send `type: "cancel"` with the matching `rid` to cancel in-flight work
5. **Close a channel**: Client sends `type: "close"` for that channel to destroy the instance

All messages include a `channel` field (arbitrary string ID) to route to the correct object instance.

Messages may also include an optional `rid` field. When present on an incoming request, the server copies it to any reply or `error` frame triggered by that request so the client can correlate responses. Within a channel, requests may run concurrently, so `rid` is the only reliable way to match replies to calls.


Example:

```go
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
	h.MustRegister(c, &Counter{MinIncrement: 1}) // the fields are shallow copied into any new instance
	s.Mux().Handle("/counter/ws/v1", h.Server()) // `h.Server()` returns an http.Handler
```

**Note**: you can specify the default values of the field. When a new channel is opened, the fields are shallow copied to the new instance.

### Client Protocol

#### Init

After connecting to the WebSocket, a client instantiates a new Counter by sending:

```json
{
  "channel": "c1",
  "rid": "req-1", // optional
  "type": "open",
  "path": "counter"
}
```

**Note**: The `path` field when opening must match the lowercase struct name (for example, `Counter` → `"counter"`).

If the object has an `Init(c ws.C)` method, it will be called during instantiation.

If the object has instead a `Init(c ws.C, data T)` method, the `data` field from the opening message will be unmarshaled into `T` and passed to `Init`.


#### Method Calls

Increment it by 7 with:

```json
{
  "channel": "c1",
  "path": "inc",
  "data": 7
}
```

**Note**: Method names have their first letter lowercased (for example, `Inc()` → `"inc"`, `IngestFile()` → `"ingestFile"`).

Query the current value:

```json
{
  "channel": "c1",
  "rid": "req-3",  // optional
  "path": "get"
}
```

The server will reply with:

```json
{
  "channel": "c1",
  "rid": "req-3", // <- same as request
  "path": "ct",
  "data": 7
}
```

If a handler returns an error, the emitted `error` frame will include the same `rid` when the request carried one.

#### Cancel

To cancel a request that is still running, send the same `channel` and `rid` with `type: "cancel"`:

```json
{
  "channel": "c1",
  "rid": "req-3",
  "type": "cancel"
}
```

Cancellation is cooperative: the handler must observe `c.Done()` or otherwise respect context cancellation. Replies are suppressed once the request context has been cancelled.

#### Close

When the client is done with an instance, it can close it by sending:

```json
{
  "channel": "c1",
  "type": "close"
}
```

This will fire the `Close(ws.C)` method if it exists, remove the channel from the connection, and leave other channels untouched.

**Note**: The `Close(ws.C)` method can never have secondary parameters, and it will be called whenever the channel or the whole connection is closed (`c.Close()` or a dropped WebSocket).


#### Multiple Instances

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

Any public field with a non-zero value is shallow-copied to new instances when a channel is opened. This lets you configure shared settings like `MinIncrement` in the example.


### Exposed Methods

Any public method (excluding `Init` and `Close` with signature `func (receiver *T) MethodName(c ws.C, [data T]) error` is exposed:

- **First parameter**: Must be `ws.C` (WebSocket context)
- **Second parameter** (optional): Receives the `data` field from the client message (unmarshaled to the parameter type)
- **Return**: Must return `error`

The `data` field from client messages is automatically unmarshaled into the method's second parameter type before calling. Calls on the same channel may execute concurrently, so handlers must be safe for concurrent use if clients issue overlapping requests.

At registration, the list of exposed methods is logged for reference, and the one rejected (if any) with the reason.

### Using `ws.C`

`ws.C` wraps a normal `ctx.C` and also provides:
- `c.Reply(path string, obj any)`: Send a reply message to the client, preserving the incoming `rid`
- `c.Close()`: Close the WebSocket connection
- All methods from `ctx.C` (logging, cancellation, etc.)


## Test

Since the WebSocket binding is automatic, you can usually test the object directly and ignore the transport layer.

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

This client allows you to open channels, call methods, and receive replies in a test environment without a real network connection.


## `Handler{}.Server()`

Calling `.Server()` on a handler returns an `http.Handler` that can be used on any HTTP server.

Internally it uses `golang.org/x/net/websocket`, which implements `http.Handler`.

You might need to modify the `.Handshake` function pointer; by default it accepts requests from the same origin.
