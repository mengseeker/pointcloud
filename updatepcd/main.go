package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pointcloud/pkg/pcd"

	"github.com/seqsense/pcgol/pc"
	"github.com/spf13/cobra"
)

var cfg struct {
	in  string
	out string
}

var cmd = &cobra.Command{
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return tranBinFiles()
	},
}

func init() {
	cmd.PersistentFlags().StringVarP(&cfg.in, "in", "i", "", "input zipFile or dir")
	cmd.PersistentFlags().StringVarP(&cfg.out, "out", "o", "", "output zipFile or dir")

	cmd.MarkPersistentFlagRequired("in")
}

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func tranBinFiles() (err error) {
	if cfg.out == "" {
		cfg.out = cfg.in
	}
	return TransDirBinToPcd(cfg.in, cfg.out)
}

func TransDirBinToPcd(sourceDir, OutDir string) (err error) {
	ds, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}
	for _, d := range ds {
		fn := d.Name()
		if filepath.Ext(fn) != ".pcd" {
			continue
		}
		src := filepath.Join(sourceDir, fn)
		out := filepath.Join(OutDir, fn)
		binf, err := os.Open(src)
		if err != nil {
			return err
		}

		pcf, err := pc.Unmarshal(binf)
		if err != nil {
			panic(err)
		}

		pcdf, err := os.Create(out)
		if err != nil {
			return err
		}
		defer pcdf.Close()

		ir, iw := io.Pipe()

		go func() {
			pc.Marshal(pcf, iw)
			iw.Close()
		}()

		pp, err := pcd.DecodePcd(ir)
		if err != nil {
			panic(err)
		}
		pp.Encode(pcdf)

		fmt.Printf("TransBinToPcd %s => %s\n", src, out)
	}
	return
}
