package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
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
		err = decoder.Decode(&req)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			res.Error = err.Error()
			encoder.Encode(res)
			continue
		}
		err = func() (err error) {
			pcdFile, err := os.Open(req.PCDFile)
			if err != nil {
				return
			}
			defer pcdFile.Close()
			p, err := pcd.DecodePcd(pcdFile)
			if err != nil {
				return
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

func Run() {
	binFile := "./pcdmeta-dawin-arm64"
	cmd := exec.Command(binFile)

	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	// 关闭输入pip时，计算进程会自动退出
	defer stdin.Close()

	runOK := make(chan int)

	// 启动
	go func() {
		err := cmd.Start()
		if err != nil {
			panic(err)
		}
		close(runOK)
		cmd.Wait()
	}()
	// defer cmd.Process.Kill()

	// 等待进程启动成功
	<-runOK

	// 请求
	inputEncoder := json.NewEncoder(stdin)
	// 获取结果
	outputDecoder := json.NewDecoder(stdout)

	// 计算
	for i := 0; i < 100; i++ {
		req := PCD{
			PCDFile: ".dev/92c82d69-4033-49b9-9cdd-17004a32f7f1.pcd",
			Labels: [][]float32{{-1.48, -7.96, -1.62, 4.62, 1.76, 1.98, -7.7908},
				{29.69, -2.25, -1.555, 10.32, 4.15, 3.26, -7.8208}},
		}

		err := inputEncoder.Encode(req)
		if err != nil {
			panic("请求计算失败：" + err.Error())
		}
		res := Result{}
		err = outputDecoder.Decode(&res)
		if err != nil {
			panic("获取计算结果失败：" + err.Error())
		}
		fmt.Println(res)
	}

}

func main() {
	// Run()
	Cal()
}
