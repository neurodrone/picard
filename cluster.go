package picard

// #cgo LDFLAGS: -lsheepdog
// #include <stdlib.h>
// #include <sheepdog/sheepdog.h>
import "C"

import (
	"fmt"
	"unsafe"
)

type Cluster struct {
	cluster *C.struct_sd_cluster
}

// NewCluster creates a new instance of cluster connection that will
// be used to issue vdi requests to Sheepdog backend. It also undertakes
// the work of connecting to the cluster so commands can be issued as
// soon as a Cluster instance is available.
//
// If there is an issue connecting to the cluster, an appropriate
// error is returned.
func NewCluster(hostport string) (c *Cluster, err error) {
	c_hostport := C.CString(hostport)
	defer C.free(unsafe.Pointer(c_hostport))

	c = &Cluster{}
	c.cluster, err = C.sd_connect(c_hostport)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to cluster: %s", err)
	}

	return c, nil
}

// Disconnect closes the connection to the Sheepdog cluster. Use this
// to cleanly relinquish all the claimed resources at the end of operation.
func (c *Cluster) Disconnect() (err error) {
	_, err = C.sd_disconnect(c.cluster)
	return err
}

// CreateVDI creates a sheepdog vdi with the vdiName name
// of the given size.
func (c *Cluster) CreateVDI(vdiName string, size uint64) error {
	c_vdiname := C.CString(vdiName)
	defer C.free(unsafe.Pointer(c_vdiname))

	_, err := C.sd_vdi_create(c.cluster, c_vdiname, C.uint64_t(size))
	return err
}

// OpenVDI opens the vdi handler to sheepdog with the given vdiName.
// The Cluster identifier should also be provided to open handler to
// the correct vdi.
func (c *Cluster) OpenVDI(vdiName string) (vdi *VDI, err error) {
	c_vdiname := C.CString(vdiName)
	defer C.free(unsafe.Pointer(c_vdiname))

	vdi = &VDI{
		cluster: c,
		name:    vdiName,
	}
	vdi.vdi, err = C.sd_vdi_open(c.cluster, c_vdiname)
	if err != nil {
		return nil, err
	}

	return vdi, nil
}

// CreateOpenVDI first creates a vdi with the given vdiName name and if
// that succeeds it opens the VDI for further operation and returns the
// vdi handler back.
//
// If there is an issue creating or opening the vdi an appropriate error
// is returned.
func (c *Cluster) CreateOpenVDI(vdiName string, size uint64) (*VDI, error) {
	if err := c.CreateVDI(vdiName, size); err != nil {
		return nil, err
	}

	return c.OpenVDI(vdiName)
}

// DeleteVDI will delete all VDIs present on sheepdog that are named
// by the name that vdiName contains.
func (c *Cluster) DeleteVDI(vdiName string) error {
	c_vdiname := C.CString(vdiName)
	defer C.free(unsafe.Pointer(c_vdiname))

	// This will delete all the VDIs that have this name.
	_, err := C.sd_vdi_delete(c.cluster, c_vdiname, nil)
	return err
}
