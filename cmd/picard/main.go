package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/ianschenck/envflag"

	"github.com/vaibhav/picard"
)

func main() {
	var (
		hostport     = envflag.String("SHEEPDOG_HOSTPORT", "", "host:port pair for sheepdog cluster")
		readHostPort = envflag.String("SHEEPDOG_READ_HOSTPORT", "", "host:port pair for only issuing reads to sheepdog cluster")
		readDelay    = envflag.Duration("SHEEPDOG_READ_SLEEP", 0, "time to sleep between each read of test")
		payloadCount = envflag.Int("SHEEPDOG_TEST_PAYLOAD_COUNT", 10, "payload count to issue reads and writes to sheepdog")
		vdiSize      = envflag.Int("SHEEPDOG_VDI_SIZE", 1<<22, "create vdi of given size")
		vdiName      = envflag.String("SHEEPDOG_VDI_NAME", "testvdi", "name of vdi to test read/writes across")
	)
	envflag.Parse()

	c, err := picard.NewCluster(*hostport)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Created connection to sheepdog successfully")

	defer func() {
		if err := c.Disconnect(); err != nil {
			log.Fatalln(err)
		}
		log.Printf("Successfully disconnected!")
	}()

	vdi, err := c.CreateOpenVDI(*vdiName, uint64(*vdiSize))
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Created and opened VDI successfully")

	defer func() {
		vdi.Close()
		if err := c.DeleteVDI(vdi.Name()); err != nil {
			log.Fatalln(err)
		}
		log.Printf("Successfully deleted vdi: %q", vdi.Name())
	}()

	rvdi := vdi
	if *readHostPort != "" {
		rc, err := picard.NewCluster(*readHostPort)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Created read connection to sheepdog successfully")

		defer func() {
			if err := rc.Disconnect(); err != nil {
				log.Fatalln(err)
			}
			log.Printf("Successfully disconnected for read connection!")
		}()

		rvdi, err = rc.OpenVDI(*vdiName)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Opened VDI successfully for reads")

		defer rvdi.Close()
	}

	vdiChan := make(chan *vdiData)

	go func() {
		if count, err := writeToVDI(vdi, vdiChan, *payloadCount); err != nil {
			log.Printf("Error while writing at %d: %s", count, err)
		}
	}()

	if count, failed, err := readFromVDI(rvdi, vdiChan, *readDelay); err != nil {
		log.Printf("Error occurred during reads:\n\tTotal: %d, Failures: %d\n\tError: %q", count, failed, err)
	}
}

type vdiData struct {
	payload []byte
}

// writeToVDI sends data packets to the vdi interface upto the
// count number that is defined within the function. If there is
// an error, it returns the no. of packets sent until the error
// occurred.
func writeToVDI(vdi *picard.VDI, vdiChan chan<- *vdiData, count int) (int, error) {
	defer close(vdiChan)

	var offset int
	for i := 0; i < count; i++ {
		buf := randPayload()
		if _, err := vdi.WriteAt(buf, offset); err != nil {
			return i, err
		}
		vdiChan <- &vdiData{buf}
		offset += len(buf)
	}
	return count, nil
}

// readFromVDI returns the count of total reads the corresponding VDI was subject to
// along with the no. of failed attempts, with the last seen error.
func readFromVDI(vdi *picard.VDI, vdiChan <-chan *vdiData, delay time.Duration) (int, int, error) {
	// size of the buffer should be equal to twice the size of
	// random payload.
	var buf [8]byte
	var total, failed, offset int
	var err error

	for vdat := range vdiChan {
		time.Sleep(delay)
		total++

		n, rerr := vdi.ReadAt(buf[:], offset)
		if rerr != nil {
			err = rerr
			failed++
		} else if !reflect.DeepEqual(vdat.payload, buf[:n]) {
			// This will register the error within `err` at a high level scope
			// and continue processing remaining reads increasing the count
			// of failures.
			err = fmt.Errorf("new payload does not match original. New: %+v, Old: %+v", vdat.payload, buf[:n])
			failed++
		} else {
			log.Printf("Complete reading %d payloads", total)
		}

		offset += n
	}
	return total, failed, err
}

// randPayload generates random byte slices to pass in as
// payload for writes to sheepdog.
func randPayload() []byte {
	var buf [4]byte
	n, err := rand.Read(buf[:])
	if err != nil {
		return []byte("rand")
	}
	return []byte(fmt.Sprintf("%X", buf[:n]))
}
