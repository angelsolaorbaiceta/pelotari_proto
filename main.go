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
		peersCh              = manager.PeersCh()
	)

	if err != nil {
		panic(err)
	}
	log.Println("========================= [Pelotari] =========================")
	log.Printf("Private IP: %s, Broadcast IP: %s\n", privIP, broadIP)

	manager.Start()
	defer func() {
		log.Println("Defer function: calling manger.Close()...")
		manager.Close()
		log.Println("Done. Bye!")
	}()

	log.Println("Pelotari protocol starting... Press CTRL+C to exit.")

	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

loop:
	for {
		select {
		case <-sigchan:
			// User pressed CTRL+C. Exit!
			log.Println("Closing connections...")
			break loop
		case peers := <-peersCh:
			log.Println("----- [Peers] -----")
			for _, peer := range peers {
				log.Printf("\t> %s\n", peer.Address())
			}
		}
	}

	log.Println("Closing connections...")
}
