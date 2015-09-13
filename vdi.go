package picard

// #cgo LDFLAGS: -lsheepdog
// #include <stdlib.h>
// #include <sheepdog/sheepdog.h>
import "C"

import "unsafe"

type VDI struct {
	cluster *Cluster
	vdi     *C.struct_sd_vdi

	name string
}

// Name returns the name of the current vdi.
func (v *VDI) Name() string {
	return v.name
}

// Close will close the vdi handler.
func (v *VDI) Close() error {
	_, err := C.sd_vdi_close(v.cluster.cluster, v.vdi)
	return err
}

// Read reads the information from the vdi into the buffer provided.
// This helps *VDI implement the io.Reader interface. Thus, once the read
// completes it will return the no. of bytes that have been successfully
// read into the buffer from the vdi.
func (v *VDI) Read(buf []byte) (int, error) {
	return v.ReadAt(buf, 0)
}

// ReadAt reads the information from the vdi starting from the
// given offset. We implemented the io.ReaderAt interface here.
// The returned values match the returned values of Read.
func (v *VDI) ReadAt(buf []byte, offset int) (int, error) {
	_, err := C.sd_vdi_read(
		v.cluster.cluster,
		v.vdi,
		unsafe.Pointer(&buf[0]),
		C.size_t(len(buf)),
		C.off_t(offset))
	return len(buf), err
}

// Write will copy the information from the buffer into the vdi.
func (v *VDI) Write(buf []byte) (int, error) {
	return v.WriteAt(buf, 0)
}

// WriteAt helps implement the io.WriterAt interface by allowing
// an additional offset to start writing to within the VDI.
func (v *VDI) WriteAt(buf []byte, offset int) (int, error) {
	if _, err := C.sd_vdi_write(
		v.cluster.cluster,
		v.vdi,
		unsafe.Pointer(&buf[0]),
		C.size_t(len(buf)),
		C.off_t(offset),
	); err != nil {
		return 0, err
	}

	return len(buf), nil
}
