package pcd

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unsafe"
)

var (
	ErrUnsupportPointCloudFileType = errors.New("unsupport pointCloud fileType")
)

func TransFileToPcd(sourceFile string, Item io.Writer) error {
	f, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	switch strings.ToLower(filepath.Ext(sourceFile)) {
	case ".bin":
		bin, err := DecodeBin(f)
		if err != nil {
			return err
		}
		pcd, err := bin.ToPcd()
		if err != nil {
			return err
		}
		return (pcd.Encode(Item))
	default:
		return ErrUnsupportPointCloudFileType
	}
	// return nil
}

type Point struct {
	X, Y, Z float32
	R       float32
}

type PointCloud struct {
	Points []Point
}

func (p *PointCloud) AddPoint(pt Point) {
	p.Points = append(p.Points, pt)
}

func ByteSliceAsFloat32Slice(b []byte) []float32 {
	n := len(b) / 4

	up := unsafe.Pointer(&(b[0]))
	pi := (*[1]float32)(up)
	buf := (*pi)[:]
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Len = n
	sh.Cap = n

	return buf
}

func Float32SliceAsByteSlice(f []float32) []byte {
	n := len(f) * 4

	up := unsafe.Pointer(&(f[0]))
	pi := (*[1]byte)(up)
	buf := (*pi)[:]
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Len = n
	sh.Cap = n

	return buf
}

func IsShadowing(b []byte, f []float32) bool {
	return uintptr(unsafe.Pointer(&f[0])) != uintptr(unsafe.Pointer(&b[0]))
}
