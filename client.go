package eventsource

import (
	"net"
	"time"
)

// A client holds the actual connection to the browser, the channels names the
// client has subscribed to, a queue to receive events and a done channel for
// syncronization with pending events.
type client struct {
	events   chan payload
	done     chan bool
	channels []string
	conn     net.Conn
}

// A payload contains the event data that must be written to the client
// connection and a done channel to signalize the end of the write process
type payload struct {
	data []byte
	done chan status
}

// A status tracks the client connection write state
type status struct {
	start time.Time
	end   time.Time
	sent  bool
}

// The listen function receives incoming events on the events channel, writing
// them to its underlining connection. If there is an error, the client send a
// message to remove itself from the pool through the remove channel passed in
// and notifies pending events by closing the done channel.
func (c *client) listen(remove chan<- client) {
	for {
		e, ok := <-c.events
		if !ok {
			c.conn.Close()
			return
		}
		c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		start := time.Now()
		_, err := c.conn.Write(e.data)
		if err == nil && e.done != nil {
			e.done <- status{start: start, end: time.Now(), sent: true}
		} else {
			remove <- *c
			c.conn.Close()
			close(c.done)
			break
		}
	}
}
