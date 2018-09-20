package main

// Import Go and NATS packages
import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	nats "github.com/nats-io/go-nats"
	stan "github.com/nats-io/go-nats-streaming"
)

func main() {
	var (
		err error
	)
	// Create server connection
	natsURL := "localhost:4222"
	metricsURL := "localhost:8080"
	clusterID := "test-cluster"
	clientID := "nats-consumer"
	nc, err := nats.Connect(
		natsURL,
		nats.MaxReconnects(30),
		nats.ReconnectWait(1*time.Second),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("[WARN] Client[%s] disconnected.\n", clientID)
			_, err := http.Get(metricsURL + "/condisconncount")
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("[WARN] Client[%s] reconnected to %v.\n", clientID, nc.ConnectedUrl())
			_, err := http.Get(metricsURL + "/conreconncount")
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			fmt.Printf("[WARN] Client[%s] connection to %v closed: %q\n", clientID, nc.ConnectedUrl(), nc.LastError())
			_, err := http.Get(metricsURL + "/conconnclosscount")
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}
		}),
	)
	natsConnection, err := stan.Connect(clusterID, clientID, stan.NatsConn(nc))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, natsURL)
	}
	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", natsURL, clusterID, clientID)
	// Subscribe to subject
	log.Printf("Subscribing to subject 'foo'\n")
	handle := func(msg *stan.Msg) {
		log.Printf("seq = %d [redelivered = %v]\n", msg.Sequence, msg.Redelivered)
		// time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
		msg.Ack()
		_, err := http.Get(metricsURL + "/conmesgcount")
		if err != nil {
			log.Printf("[ERROR] %s", err)
		}
	}
	natsConnection.Subscribe("foo",
		handle,
		stan.DurableName("i-will-remember"),
	)
	if err != nil {
		log.Printf("[ERROR] %s", err)
	}
	// Keep the connection alive
	runtime.Goexit()
}
