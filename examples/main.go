package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bazil.org/fuse"
	rs "git.nightcrickets.space/keefleoflimon/resonatefuse"
)

func main() {
	done := make(chan struct{})
	volume := rs.NewVolume("fake", func(path string, req *fuse.CreateRequest) { fmt.Println("Hello", path) })
	closeHandler(volume, done)

	if err := volume.Serve(); err != nil {
		log.Println(err)
	}

	<-done
}

func closeHandler(v *rs.Volume, done chan<- struct{}) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if err := v.Stop(); err != nil {
			log.Println(err)
		}
		done <- struct{}{}
		os.Exit(0)
	}()
}
