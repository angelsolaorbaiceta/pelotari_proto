package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/angelsolaorbaiceta/prototari/prototari"
)

func main() {
	var (
		config               = prototari.MakeDefaultConfig()
		manager              = prototari.MakeUDPManager(config)
		sigchan              = make(chan os.Signal, 1)
		privIP, broadIP, err = prototari.GetPrivateIPAndBroadcastAddr()
	)

	if err != nil {
		panic(err)
	}
	log.Println("========================= [Pelotari] =========================")
	log.Printf("Private IP: %s, Broadcast IP: %s\n", privIP, broadIP)

	manager.Start()
	defer func() {
		manager.Stop()
		log.Println("Done. Bye!")
	}()

	log.Println("Pelotari protocol starting... Press CTRL+C to exit.")

	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan

	log.Println("Closing connections...")
}
