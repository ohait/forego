# WebSockets

This library allows you to bind multiple objects to a WebSocket, exposing some of their methods.

You could call it RPC over Websockets.


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

After a clients connect, it will instantiate a new Counter by sending:

```
{
  channel: "1234",
  type: "new",
  path: "counter",
}
```

Increment it by 7 with:

```
{
  channel: "1234",
  path: "inc",
  data: 7,
}
```

Query the current value:

```
{
  channel: "1234",
  path: "get",
}
```

And closing the channel with:
```
{
  channel: "1234",
  type "close",
}
```

Or instantiate another `Counter` (or any other object) by using different channels.

## Reflection

This library inspect the given object

### Fields

Any field which is public and not zero, is used to initialize the objects when instantiated by a new channel

### `Init()`

The special field `Init()` won't be exposed, but used as an extra step for initialization

### Methods

Any other public method that has the first argument of type `ws.C` and on optional argument, will be exposed.

The `data` field is expanded in the optional argument before calling the function

The method must return an error

`ws.C` can be used to send replies or close the channel

### Names

By default, names are obtained from methods and types directly, with the initial lower-case.

TODO: allow overrides


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
