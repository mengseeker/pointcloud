package pointcloud

import (
	"os"
	"pointcloud/pkg/pcd"
	"testing"
)

func TestDecodePCD(t *testing.T) {
	var pcdf = ".dev/in/mul.pcd"
	pcdio, err := os.Open(pcdf)
	if err != nil {
		t.Fatal(err)
	}
	d, err := pcd.DecodePcd(pcdio)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("size:", len(d.Points))
}
