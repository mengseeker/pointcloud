package pcd

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"pointcloud/pkg/lzf"
	"strconv"
	"strings"
)

var (
	ErrUnsupportPcdVersion   = errors.New("unsupport pcd version")
	ErrUnsupportPcdFieldSize = errors.New("unsupport pcd field size")
	ErrUnsupportPcdFieldType = errors.New("unsupport pcd field type")
	ErrUnsupportPcdDataType  = errors.New("unsupport pcd data type")
	ErrInvalidPcdFormat      = errors.New("invalid pcd format")
)

type Pcd struct {
	PointCloud
}

func DecodePcd(r io.Reader) (pcd *Pcd, err error) {
	bio := bufio.NewReader(r)
	var version string
	for {
		version, err = bio.ReadString('\n')
		if err != nil {
			return
		}
		if !strings.HasPrefix(version, "#") {
			break
		}
	}

	if !strings.HasPrefix(version, "VERSION 0.7") {
		return nil, ErrUnsupportPcdVersion
	}

	var headers = map[string][]string{}
	for i := 0; i < 9; i++ {
		header, err := bio.ReadString('\n')
		if err != nil {
			return nil, err
		}
		h := strings.Split(strings.TrimSuffix(header, "\n"), " ")
		if len(h) < 1 {
			return nil, ErrInvalidPcdFormat
		}
		headers[h[0]] = h[1:]
	}

	var fields = map[string]int{}
	for i, f := range headers["FIELDS"] {
		fields[f] = i
	}
	sizes, err := getIntHeaders(headers, "SIZE")
	if err != nil {
		return
	}
	if len(fields) != len(sizes) {
		return nil, ErrInvalidPcdFormat
	}

	types := headers["TYPE"]
	if len(fields) != len(types) {
		return nil, ErrInvalidPcdFormat
	}

	counts, err := getIntHeaders(headers, "COUNT")
	if err != nil {
		return
	}
	if len(fields) != len(counts) {
		return nil, ErrInvalidPcdFormat
	}

	if len(headers["DATA"]) != 1 {
		return nil, ErrInvalidPcdFormat
	}
	dataType := strings.ToLower(headers["DATA"][0])

	if len(headers["WIDTH"]) != 1 || len(headers["HEIGHT"]) != 1 {
		return nil, ErrInvalidPcdFormat
	}
	width, _ := strconv.Atoi(headers["WIDTH"][0])
	height, _ := strconv.Atoi(headers["HEIGHT"][0])

	pcd = &Pcd{
		PointCloud: PointCloud{
			Points: []Point{},
		},
	}
	if dataType == "binary" {
		err = pcd.LoadBinPoints(bio, width, height, fields, sizes, counts, types)
		if err != nil {
			return
		}
	} else if dataType == "ascii" {
		err = pcd.LoadAsciiPoints(bio, fields, sizes, counts, types)
		if err != nil {
			return
		}
	} else if dataType == "binary_compressed" {
		err = pcd.LoadBinCompressedPoints(bio, width, height, fields, sizes, counts, types)
		if err != nil {
			return
		}
	} else {
		return nil, ErrUnsupportPcdDataType
	}

	return
}

const (
	BinaryCompressedSize = 8
)

func (pcd *Pcd) LoadBinCompressedPoints(r io.Reader, width, height int, fields map[string]int, sizes, counts []int, types []string) (err error) {
	compressedSizesRaw := make([]byte, BinaryCompressedSize)
	n, err := io.ReadFull(r, compressedSizesRaw)
	if err != nil {
		return
	}
	if n != BinaryCompressedSize {
		return ErrInvalidPcdFormat
	}
	compressedSize := binary.LittleEndian.Uint32(compressedSizesRaw[:4])
	uncompressedSize := binary.LittleEndian.Uint32(compressedSizesRaw[4:])
	// 计算单点数据大小
	var pointSize int
	for i := range sizes {
		pointSize += sizes[i] * counts[i]
	}
	if uncompressedSize != uint32(width*height*pointSize) {
		return ErrInvalidPcdFormat
	}

	raw := make([]byte, compressedSize)
	_, err = io.ReadFull(r, raw)
	if err != nil {
		return
	}
	uncompressed := make([]byte, uncompressedSize)

	n, err = lzf.Decompress(raw, uncompressed)
	if err != nil {
		return
	}
	if n != int(uncompressedSize) {
		return ErrInvalidPcdFormat
	}

	return pcd.LoadBinPoints(bytes.NewReader(uncompressed), width, height, fields, sizes, counts, types)
}

func (pcd *Pcd) LoadBinPoints(r io.Reader, width, height int, fields map[string]int, sizes, counts []int, types []string) (err error) {
	if err = checkfield(fields, sizes, counts, types, "x"); err != nil {
		return
	}
	if err = checkfield(fields, sizes, counts, types, "y"); err != nil {
		return
	}
	if err = checkfield(fields, sizes, counts, types, "z"); err != nil {
		return
	}

	xi, xb, xe := getfieldIndexAndOffset(fields, sizes, counts, "x")
	yi, yb, ye := getfieldIndexAndOffset(fields, sizes, counts, "y")
	zi, zb, ze := getfieldIndexAndOffset(fields, sizes, counts, "z")
	if !(xi >= 0 && yi >= 0 && zi >= 0) {
		return ErrInvalidPcdFormat
	}

	var w int
	for i := range counts {
		w += counts[i] * sizes[i]
	}
	bs := make([]byte, w)
	for {
		_, err = io.ReadFull(r, bs)
		if err != nil {
			if errors.Is(io.EOF, err) {
				break
			}
			return err
		}

		pcd.AddPoint(Point{
			X: math.Float32frombits(binary.LittleEndian.Uint32(bs[xb:xe])),
			Y: math.Float32frombits(binary.LittleEndian.Uint32(bs[yb:ye])),
			Z: math.Float32frombits(binary.LittleEndian.Uint32(bs[zb:ze])),
			R: 1,
		})

	}

	return nil
}

func (pcd *Pcd) LoadAsciiPoints(r *bufio.Reader, fields map[string]int, sizes, counts []int, types []string) error {
	var fs []float64
	var err error
	xi, _, _ := getfieldIndexAndOffset(fields, sizes, counts, "x")
	yi, _, _ := getfieldIndexAndOffset(fields, sizes, counts, "y")
	zi, _, _ := getfieldIndexAndOffset(fields, sizes, counts, "z")
	if !(xi >= 0 && yi >= 0 && zi >= 0) {
		return ErrInvalidPcdFormat
	}
	var l int
	for _, i := range counts {
		l += i
	}
	fs = make([]float64, l)
	for {
		fs = fs[:0]
		err = AsciiGetFloats(r, &fs)
		if err != nil {
			if errors.Is(io.EOF, err) {
				break
			}
			return err
		}
		if len(fs) != l {
			return ErrInvalidPcdFormat
		}

		pcd.AddPoint(Point{
			X: float32(fs[xi]),
			Y: float32(fs[yi]),
			Z: float32(fs[zi]),
			R: 1,
		})

	}
	return nil
}

func (pcd *Pcd) Encode(w io.Writer) error {
	byf := bytes.NewBuffer(make([]byte, 0, 120))
	byf.WriteString("# .PCD v0.7 - Point Cloud Data file format\n")
	byf.WriteString("VERSION 0.7\n")
	byf.WriteString("FIELDS x y z intensity\n")
	byf.WriteString("SIZE 4 4 4 4\n")
	byf.WriteString("TYPE F F F F\n")
	byf.WriteString("COUNT 1 1 1 1\n")
	byf.WriteString(fmt.Sprintf("WIDTH %d\n", len(pcd.Points)))
	byf.WriteString("HEIGHT 1\n")
	byf.WriteString("VIEWPOINT 0 0 0 1 0 0 0\n")
	byf.WriteString(fmt.Sprintf("POINTS %d\n", len(pcd.Points)))
	byf.WriteString("DATA binary\n")
	// fmt.Print(byf.String())
	_, err := w.Write(byf.Bytes())
	if err != nil {
		return err
	}
	for _, p := range pcd.Points {
		pio := bytes.NewBuffer(make([]byte, 0, 16))
		binary.Write(pio, binary.LittleEndian, p.X)
		binary.Write(pio, binary.LittleEndian, p.Y)
		binary.Write(pio, binary.LittleEndian, p.Z)
		binary.Write(pio, binary.LittleEndian, p.R)
		_, err := w.Write(pio.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (pcd *Pcd) PointCount() int {
	return len(pcd.Points)
}

// precision 精度, 越大时，精度越高， 1则精度为m2
func (pcd *Pcd) XYArea(precision float32) float32 {
	areas := make(map[int]map[int]bool)
	var x, y, sum int
	var l map[int]bool
	var ok bool
	for _, p := range pcd.Points {
		x = int(p.X * precision)
		y = int(p.Y * precision)
		l, ok = areas[x]
		if !ok {
			areas[x] = make(map[int]bool)
			l = areas[x]
		}
		l[y] = true
	}
	for _, l := range areas {
		for range l {
			sum++
		}
	}
	return float32(sum) / precision / precision
}

// 计算区域内点数量, 物体必须平行于xy平面
func (pcd *Pcd) XYAreaPointCount(cx, cy, cz, height, width, depth, rx float32) int {
	var count int
	var sin, cos, x64, y64, d1, d2 float64
	var cx64, cy64 = float64(cx), float64(cy)
	var width64, height64 = float64(width), float64(height)
	for _, p := range pcd.Points {
		if p.Z > cz+depth/2 || p.Z < cz-depth/2 {
			continue
		}
		// 判断点和物体方向的水平距离和垂直距离是否在h/2和w/2内
		x64, y64 = float64(p.X), float64(p.Y)
		sin = math.Sin(float64(rx))
		cos = math.Cos(float64(rx))
		d1 = math.Abs(-sin*x64 + cos*y64 + (sin*cx64 - cos*cy64))
		sin = math.Sin(float64(rx + math.Pi/2))
		cos = math.Cos(float64(rx + math.Pi/2))
		d2 = math.Abs(-sin*x64 + cos*y64 + (sin*cx64 - cos*cy64))
		// fmt.Printf("d: %f, %f\n", d1, d2)
		if d1 > height64/2 || d2 > width64/2 {
			continue
		}
		count++
	}
	return count
}

func getIntHeaders(headers map[string][]string, field string) ([]int, error) {
	vals := []int{}
	for _, v := range headers[field] {
		vi, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid int field %s", field)
		}
		vals = append(vals, vi)
	}
	return vals, nil
}

func checkfield(fields map[string]int, sizes, counts []int, types []string, field string) error {
	i, ok := fields[field]
	if !ok {
		return ErrInvalidPcdFormat
	}
	if sizes[i] != 4 {
		return ErrUnsupportPcdFieldSize
	}
	if types[i] != "F" {
		return ErrUnsupportPcdFieldType
	}
	return nil
}

func getfieldIndexAndOffset(fields map[string]int, sizes, counts []int, field string) (idx, begin, end int) {
	idx = -1
	id, ok := fields[field]
	if !ok {
		return
	}
	idx = 0
	for i := 0; i < id; i++ {
		idx += counts[i]
		begin += sizes[i] * counts[i]
	}
	end = begin + sizes[id]*counts[id]
	return
}

func AsciiGetFloats(r *bufio.Reader, fs *[]float64) (err error) {
	line, _, err := r.ReadLine()
	if err != nil {
		return
	}
	var v float64
	for _, r := range strings.Split(string(line), " ") {
		v, err = strconv.ParseFloat(r, 64)
		if err != nil {
			return
		}
		*fs = append(*fs, v)
	}
	return
}
