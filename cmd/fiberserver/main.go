package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.con/napat/gofiber/internal/fiberserver"
)

func main() {
	fmt.Printf("GOMAXPROCS is %d\n", runtime.GOMAXPROCS(0))
	serv := fiberserver.NewServHandler()

	runGracefulServer(serv)
}

func runGracefulServer(serv *fiberserver.Handler) {
	go func() {
		if err := serv.Listen(":3000"); err != nil {
			log.Panic(err)
		}
	}()
	log.Print("Server Started")

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

	<-c // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")
	_ = serv.App.Shutdown()

	fmt.Println("Running cleanup tasks...")

	// Your cleanup tasks go here
	// db.Close()
	// redisConn.Close()
	fmt.Println("Fiber was successful shutdown.")
}
