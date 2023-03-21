package pcd

import (
	"errors"
	"io"
)

const (
	BinPointDataLen = 4 * 4
)

var (
	ErrInvalidDataFormat = errors.New("invalid data")
)

type Bin struct {
	PointCloud
}

func DecodeBin(r io.Reader) (bin *Bin, err error) {
	bin = &Bin{}
	var data = make([]byte, BinPointDataLen)
	var n int
	for {
		n, err = r.Read(data)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return
		}
		if n != BinPointDataLen {
			err = ErrInvalidDataFormat
			return
		}
		pointFlots := ByteSliceAsFloat32Slice(data)
		if len(pointFlots) != 4 {
			panic(ErrInvalidDataFormat)
		}
		bin.AddPoint(Point{
			X: pointFlots[0],
			Y: pointFlots[1],
			Z: pointFlots[2],
			R: pointFlots[3],
		})
	}
	return
}

func (bin *Bin) ToPcd() (*Pcd, error) {
	return &Pcd{
		PointCloud: bin.PointCloud,
	}, nil
}
