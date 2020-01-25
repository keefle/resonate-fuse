package main

import (
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	rs "git.iyi.cz/mo/resonatefuse"
)

func main() {
	mountpoint := "fake"

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("helloworld"),
		fuse.Subtype("hellofs"),
		fuse.LocalVolume(),
		fuse.VolumeName("Hello world!"),
	)

	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	fsy := rs.NewFS("fake")

	err = fs.Serve(c, fsy)
	defer func() {
		if err := fuse.Unmount(mountpoint); err != nil {
			log.Println(err)
		}
	}()

	if err != nil {
		log.Fatal(err)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}
