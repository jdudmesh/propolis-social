package user

import (
	"fmt"
	"os"
	"path"
	"testing"
)

func TestMain(m *testing.M) {
	dataDir := os.Getenv("DATA_DIR")
	files, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Println(err)
	}

	for _, file := range files {
		fmt.Println(path.Join(dataDir, file.Name()))
		os.Remove(path.Join(dataDir, file.Name()))
	}

	os.Exit(m.Run())
}
