// +build integration

package picard

import (
	"reflect"
	"testing"

	"github.com/ianschenck/envflag"
)

const (
	vdiName = "testVDI"
)

var (
	dataBytes = []byte{1, 2, 3, 4, 5}

	vdi *VDI
)

func setupCluster(t *testing.T) *Cluster {
	var (
		hostport = envflag.String("SHEEPDOG_HOSTPORT", "", "Host:port for Sheepdog server")
	)
	envflag.Parse()

	c, err := NewCluster(*hostport)
	if err != nil {
		t.Fatalf("unable to create cluster: %s", err)
	}

	return c
}

func tearDownCluster(c *Cluster, t *testing.T) {
	if err := c.Disconnect(); err != nil {
		t.Errorf("failed disconnecting: %s", err)
	}
}

func createVDI(c *Cluster, t *testing.T) {
	var err error
	vdi, err = c.CreateOpenVDI(vdiName, uint64(len(dataBytes)))
	if err != nil {
		t.Errorf("failed creating or opening vdi %q: %s", vdiName, err)
	}

	if vdi.Name() != vdiName {
		t.Errorf("vdi created with wrong name\n\texpected: %q\n\tactual: %q",
			vdiName,
			vdi.Name())
	}
}

func writeToVDI(vdi *VDI, t *testing.T) {
	if _, err := vdi.Write(dataBytes); err != nil {
		t.Errorf("failed writing to vdi %q: %s", vdi.Name(), err)
	}
}

func writeToVDIOffset(vdi *VDI, t *testing.T, offset int) {
	if _, err := vdi.WriteAt(dataBytes, offset); err != nil {
		t.Errorf("failed writing to vdi %q at offset %d: %s", vdi.Name(), offset, err)
	}
}

func readFromVDI(vdi *VDI, t *testing.T) {
	buf := make([]byte, len(dataBytes))
	if _, err := vdi.Read(buf); err != nil {
		t.Errorf("failed reading from vdi %q: %s", vdi.Name(), err)
	}

	if !reflect.DeepEqual(buf, dataBytes) {
		t.Errorf("read write mismatch on vdi\n\texpected: %+v\n\tactual: %+v",
			dataBytes,
			buf)
	}
}

func readFromVDIOffset(vdi *VDI, t *testing.T, offset int) {
	buf := make([]byte, len(dataBytes))
	if _, err := vdi.ReadAt(buf, offset); err != nil {
		t.Errorf("failed reading from vdi %q at offset %d: %s", vdi.Name(), offset, err)
	}

	if !reflect.DeepEqual(buf, dataBytes) {
		t.Errorf("read write mismatch on vdi\n\texpected: %+v\n\tactual: %+v",
			dataBytes,
			buf)
	}
}

func closeVDI(vdi *VDI, t *testing.T) {
	if err := vdi.Close(); err != nil {
		t.Errorf("error closing vdi %q: %s", vdiName, err)
	}
}

func deleteVDI(c *Cluster, t *testing.T) {
	if err := c.DeleteVDI(vdiName); err != nil {
		t.Errorf("failed deleting vdi %q: %s", vdiName, err)
	}
}

func TestVDIRun(t *testing.T) {
	c := setupCluster(t)
	defer tearDownCluster(c, t)

	createVDI(c, t)
	defer func() {
		closeVDI(vdi, t)
		deleteVDI(c, t)
	}()

	writeToVDI(vdi, t)
	readFromVDI(vdi, t)

	writeToVDIOffset(vdi, t, 5)
	readFromVDIOffset(vdi, t, 5)
}

/**
 * Because of lack of timeout, currently `sd_vdi_write()` hangs and we
 * cannot keep running this test for infinite amount of time.
 *
func TestVDIWriteFail(t *testing.T) {
	c := setupCluster(t)
	defer tearDownCluster(c, t)

	vdi = &VDI{cluster: c, name: "testArbitVDI", vdi: nil}
	n, err := vdi.Write(dataBytes)
	if err == nil || n != 0 {
		t.Errorf("vdi is not initialized correctly, write to it should fail")
	}
}
*/
