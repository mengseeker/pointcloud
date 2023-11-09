package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"pointcloud/pkg/pcd"
)

type PCD struct {
	PCDFile string
	Labels  [][]float32
}

type Result struct {
	Error      string
	Area       float32
	LabelCount []int
}

func Cal() {
	var req PCD
	var res Result
	var err error
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	for {
		res.Error = ""
		err = decoder.Decode(&req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			res.Error = fmt.Sprintf("read request err: %v", err)
			encoder.Encode(res)
			continue
		}
		err = func() (err error) {
			pcdFile, err := os.Open(req.PCDFile)
			if err != nil {
				return fmt.Errorf("open pcd file err: %v", err)
			}
			defer pcdFile.Close()
			p, err := pcd.DecodePcd(pcdFile)
			if err != nil {
				return fmt.Errorf("decode pcd file err: %v", err)
			}
			res.Area = p.XYArea(0.08)
			res.LabelCount = []int{}
			for i := range req.Labels {
				if len(req.Labels[i]) != 7 {
					err = errors.New("invalid labels")
					return
				}
				res.LabelCount = append(res.LabelCount, p.XYAreaPointCount(
					req.Labels[i][0], req.Labels[i][1], req.Labels[i][2], req.Labels[i][3], req.Labels[i][4], req.Labels[i][5], req.Labels[i][6]))
			}
			encoder.Encode(res)
			return
		}()
		if err != nil {
			res.Error = err.Error()
			encoder.Encode(res)
			continue
		}

	}

}

func main() {
	Cal()
}
