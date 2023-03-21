package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pointcloud/pkg/pcd"
	"strings"

	"github.com/spf13/cobra"
)

var cfg struct {
	in  string
	out string
}

var cmd = &cobra.Command{
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if strings.HasSuffix(cfg.in, ".zip") {
			return tranZipFile()
		}
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

func tranZipFile() (err error) {
	if cfg.out == "" {
		base := filepath.Base(cfg.in)
		ext := filepath.Ext(base)
		cfg.out = strings.TrimSuffix(base, ext) + "-pcd" + ext
	}
	if cfg.out == cfg.in {
		return errors.New("input file can not sample as output file")
	}
	outFile, err := os.Create(cfg.out)
	if err != nil {
		return err
	}
	defer outFile.Close()
	outZip := zip.NewWriter(outFile)
	defer outZip.Close()
	inZip, err := zip.OpenReader(cfg.in)
	if err != nil {
		return err
	}
	for _, f := range inZip.File {
		if filepath.Ext(f.Name) == ".bin" {
			err = func() (err error) {
				binr, err := f.Open()
				if err != nil {
					return err
				}
				defer binr.Close()
				bin, err := pcd.DecodeBin(binr)
				if err != nil {
					return
				}
				pcdFile, _ := bin.ToPcd()
				newName := strings.TrimSuffix(f.Name, ".bin") + ".pcd"
				w, err := outZip.Create(newName)
				if err != nil {
					return
				}
				err = pcdFile.Encode(w)
				if err != nil {
					return
				}
				return
			}()
			if err != nil {
				return
			}
		} else {
			// w, err := outZip.CreateHeader(&f.FileHeader)
			w, err := outZip.CreateRaw(&f.FileHeader)
			if err != nil {
				return err
			}
			r, err := f.OpenRaw()
			if err != nil {
				return err
			}
			_, err = io.Copy(w, r)
			if err != nil {
				return err
			}
		}
	}
	return
}

func TransDirBinToPcd(sourceDir, OutDir string) (err error) {
	ds, err := os.ReadDir(sourceDir)
	if err != nil {
		return err
	}
	for _, d := range ds {
		fn := d.Name()
		if filepath.Ext(fn) != ".bin" {
			continue
		}
		src := filepath.Join(sourceDir, fn)
		out := filepath.Join(OutDir, strings.TrimSuffix(fn, ".bin")+".pcd")
		binf, err := os.Open(src)
		if err != nil {
			return err
		}
		bin, err := pcd.DecodeBin(binf)
		if err != nil {
			return err
		}
		pcd, err := bin.ToPcd()
		if err != nil {
			return err
		}
		pcdf, err := os.Create(out)
		if err != nil {
			return err
		}
		defer pcdf.Close()
		err = pcd.Encode(pcdf)
		if err != nil {
			return err
		}
		fmt.Printf("TransBinToPcd %s => %s\n", src, out)
	}
	return
}
