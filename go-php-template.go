package phptemplate

import (
	"os/exec"

	"github.com/pquerna/ffjson/ffjson"
)

func TmplGetByte(path string, data interface{}) ([]byte, error) {
	json, _ := ffjson.Marshal(data)
	return exec.Command("php", path, string(json)).Output()
}
