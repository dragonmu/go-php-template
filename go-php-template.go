package phptemplate

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os/exec"

	"github.com/tomasen/fcgi_client"
)

type PhpTemplater struct {
	DOCUMENT_ROOT string
	Cache         map[string][]byte
}

func GetJson(data interface{}) []byte {
	json, err := json.Marshal(data)
	if err != nil {
		return []byte{}
	}
	return json
}

func (this *PhpTemplater) Init(data map[string]string) {
	this.DOCUMENT_ROOT = data["DOCUMENT_ROOT"]
	this.Cache = make(map[string][]byte)
	return
}

func (this *PhpTemplater) sendCommandFastCgi(file string, jsonData string) ([]byte, error) {
	env := make(map[string]string)
	env["DOCUMENT_ROOT"] = this.DOCUMENT_ROOT
	env["SCRIPT_FILENAME"] = this.DOCUMENT_ROOT + "\\" + file
	log.Println(env["SCRIPT_FILENAME"])

	fcgi, err := fcgiclient.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		log.Println("err:", err)
	}

	resp, err := fcgi.PostForm(env, url.Values{"goData": {jsonData}})
	if err != nil {
		log.Println("err:", err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err:", err)
	}

	return content, err
}

func SendCommand(service string, file string, jsonData string) ([]byte, error) {
	return exec.Command(service, file, jsonData).Output()
}

func (this *PhpTemplater) TmplGetByte(file string, data interface{}) ([]byte, error) {
	return this.sendCommandFastCgi(file, string(GetJson(data)))
}

func (this *PhpTemplater) TmplGetByteCacheTag(file string, data interface{}, tag string) ([]byte, error) {
	if _, ok := this.Cache[tag]; ok {
		return this.Cache[tag], nil
	} else {
		result, err := this.sendCommandFastCgi(file, string(GetJson(data)))
		this.Cache[tag] = result
		return result, err
	}

}
