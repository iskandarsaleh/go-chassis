package main

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"

	"fmt"
	"github.com/go-chassis/go-chassis"
	_ "github.com/go-chassis/go-chassis/bootstrap"
	"github.com/go-chassis/go-chassis/client/rest"
	"github.com/go-chassis/go-chassis/core"
	"github.com/go-chassis/go-chassis/core/lager"
	"io/ioutil"
)

//if you use go run main.go instead of binary run, plz export CHASSIS_HOME=/{path}/{to}/fileupload/client/
func main() {
	//Init framework
	if err := chassis.Init(); err != nil {
		lager.Logger.Error("Init failed." + err.Error())
		return
	}

	// file / form to upload
	uploadfile("file.input")
	uploadform("form.input")
}

func uploadfile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		lager.Logger.Error("Error in opening file" + err.Error())
		return
	}
	defer f.Close()

	body := &bytes.Buffer{}

	_, err = io.Copy(body, f)
	if err != nil {
		lager.Logger.Error("Copy failed." + err.Error())
		return
	}

	req, err := rest.NewRequest("POST", "cse://FileUploadServer/uploadfile", body.Bytes())
	if err != nil {
		lager.Logger.Error("new request failed." + err.Error())
		return
	}
	defer req.Close()

	req.SetHeader("Content-Type", "application/octet-stream")

	resp, err := core.NewRestInvoker().ContextDo(context.TODO(), req)
	if err != nil {
		lager.Logger.Error("do request failed." + err.Error())
		return
	}
	defer resp.Close()
	lager.Logger.Info("FileUploadServer Response: " + string(resp.ReadBody()))

}

func uploadform(filename string) {
	//Form part
	headBuf := bytes.NewBufferString("")
	headBufWriter := multipart.NewWriter(headBuf)
	_, err := headBufWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		lager.Logger.Error("Error in create form file" + err.Error())
		return
	}

	f, err := os.Open(filename)
	if err != nil {
		lager.Logger.Error("Error in opening file" + err.Error())
		return
	}
	defer f.Close()

	fs, err := f.Stat()
	if err != nil {
		lager.Logger.Error("Error in stat file" + err.Error())
		return
	}

	lastBoundary := bytes.NewBufferString(fmt.Sprintf("\r\n--%s--\r\n", headBufWriter.Boundary()))

	bodyReader := io.MultiReader(headBuf, f, lastBoundary)

	req, err := rest.NewRequest("POST", "cse://FileUploadServer/uploadform")
	if err != nil {
		lager.Logger.Error("new request failed." + err.Error())
		return
	}
	req.Req.Body = ioutil.NopCloser(bodyReader)
	req.SetHeader("Content-Type", headBufWriter.FormDataContentType())
	req.Req.ContentLength = int64(headBuf.Len()) + fs.Size() + int64(lastBoundary.Len())

	defer req.Close()

	resp, err := core.NewRestInvoker().ContextDo(context.TODO(), req)
	if err != nil {
		lager.Logger.Error("do request failed." + err.Error())
		return
	}
	defer resp.Close()
	lager.Logger.Info("FileUploadServer Response: " + string(resp.ReadBody()))

}
