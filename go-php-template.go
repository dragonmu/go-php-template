package phptemplate

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/tomasen/fcgi_client"
)

type PhpTemplater struct {
	DOCUMENT_ROOT string
	Cache         map[string][]byte
	Request       *http.Request
	IpPort        string
}

func GetJson(data interface{}) []byte {
	json, err := json.Marshal(data)
	if err != nil {
		return []byte{}
	}
	return json
}

func (this *PhpTemplater) SetRequest(request *http.Request) {
	this.Request = request
	return
}

func (this *PhpTemplater) Init(data map[string]string) {
	this.DOCUMENT_ROOT = data["DOCUMENT_ROOT"]
	this.Cache = make(map[string][]byte)
	if value, ok := data["IP_PORT"]; ok {
		this.IpPort = value
	} else {
		this.IpPort = "127.0.0.1:9000"
	}

	return
}

func (this *PhpTemplater) sendCommandFastCgi(file string, data url.Values) ([]byte, error) {
	env := make(map[string]string)
	env["DOCUMENT_ROOT"] = this.DOCUMENT_ROOT
	env["SCRIPT_FILENAME"] = this.DOCUMENT_ROOT + "\\" + file
	if this.Request != nil {
		env["SCRIPT_NAME"] = this.Request.URL.Path
		env["REQUEST_URI"] = this.Request.URL.String()
		env["DOCUMENT_ROOT"] = this.DOCUMENT_ROOT
		env["DOCUMENT_URI"] = this.Request.URL.Path
		env["SCRIPT_FILENAME"] = this.DOCUMENT_ROOT + "\\" + file
		env["HTTP_COOKIE"] = this.Request.Header.Get("Cookie")
		env["REMOTE_ADDR"] = this.Request.RemoteAddr
		env["QUERY_STRING"] = this.Request.URL.RawQuery
	}

	fcgi, err := fcgiclient.Dial("tcp", this.IpPort)
	if err != nil {
		log.Println("err:", err)
	}

	resp, err := fcgi.PostForm(env, data)
	if err != nil {
		log.Println("err:", err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err:", err)
	}

	return content, err
}

//Функция отправляет запрос к fcgi и возвращает контент и результат запроса
func (this *PhpTemplater) sendCommandFastCgiWithRequest(file string, request *http.Request) ([]byte, *http.Response, error) {
	//Перменные окружения
	env := make(map[string]string)

	env["SCRIPT_NAME"] = request.URL.Path
	env["REQUEST_URI"] = request.URL.String()
	env["DOCUMENT_ROOT"] = this.DOCUMENT_ROOT
	env["DOCUMENT_URI"] = request.URL.Path
	env["SCRIPT_FILENAME"] = this.DOCUMENT_ROOT + "\\" + file
	env["HTTP_COOKIE"] = request.Header.Get("Cookie")
	env["REMOTE_ADDR"] = request.RemoteAddr
	env["QUERY_STRING"] = request.URL.RawQuery

	//Подключение к fcgi
	fcgi, err := fcgiclient.Dial("tcp", this.IpPort)
	if err != nil {
		log.Println("err:", err)
	}

	//Выполняем запрос скрипта с окружением
	resp, err := fcgi.PostForm(env, request.Form)
	if err != nil {
		log.Println("err:", err)
	}

	//Читаем тело ответа в переменную
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("err:", err)
	}

	//Возвращаем контент и результат запроса
	return content, resp, err
}

func (this *PhpTemplater) TmplGetByteByRequest(file string, request *http.Request) ([]byte, *http.Response, error) {
	return this.sendCommandFastCgiWithRequest(file, request)
}

func SendCommand(service string, file string, jsonData string) ([]byte, error) {
	return exec.Command(service, file, jsonData).Output()
}

func (this *PhpTemplater) TmplGetByte(file string, data url.Values) ([]byte, error) {
	return this.sendCommandFastCgi(file, data)
}

func (this *PhpTemplater) TmplGetByteCacheTag(file string, data url.Values, tag string) ([]byte, error) {
	if _, ok := this.Cache[tag]; ok {
		return this.Cache[tag], nil
	} else {
		result, err := this.sendCommandFastCgi(file, data)
		this.Cache[tag] = result
		return result, err
	}

}
