package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	dir 			= 		"/Users/shine/Desktop"
	FILE 			= 		"File"
	FOLDER 			= 		"Folder"
)

type File struct {
	ID				int 	`json:"id"`
	Title			string	`json:"title"`
	Type			string	`json:"type"`
}

func createFileJSON(id int, title string, fileType string) *File {
	file := new(File)
	file.ID = id
	file.Title = title
	file.Type = fileType
	return file
}

func (file File) addToJSON(files []File) []File {
	files = append(files, file)
	return files
}

func defineFileOrFolder(filename string) string {
	if strings.Contains(filename, ".") {
		return FILE
	} else {
		return FOLDER
	}
}

func copyFile(c echo.Context, src multipart.File, path string) error {
	dst, err := os.Create(path)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func getFiles(c echo.Context) {
	files, err := ioutil.ReadDir(dir + c.Request().RequestURI)
	if err != nil {
		log.Print(err)
		c.Response().WriteHeader(http.StatusNotFound)
	}

	var foldersFiles []File
	var file File

	for index, f := range files {
		file.ID = index
		file.Title = f.Name()
		file.Type = defineFileOrFolder(f.Name())
		foldersFiles = file.addToJSON(foldersFiles)
	}

	c.JSON(http.StatusOK, foldersFiles)
}

func getFile(c echo.Context) {
	c.File(dir + c.Request().RequestURI)
	c.Response().WriteHeader(http.StatusOK)
}

func handleGETMethod(c echo.Context) error {
	path := dir + c.Request().RequestURI
	fi, err := os.Stat(path)
	if err != nil {
		c.Response().WriteHeader(http.StatusNotFound)
		return nil
	}

	if fi.IsDir() {
		getFiles(c)
	} else {
		getFile(c)
	}
	return nil
}

func handlePUTMethod(c echo.Context) error {

	_, err := os.Stat(dir + c.Request().RequestURI)
	if err != nil {
		c.Response().WriteHeader(http.StatusNotFound)
		return nil
	}

	file, err := c.FormFile(FILE)
	if err != nil {
		c.Response().WriteHeader(http.StatusNotFound)
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	path := dir + c.Request().RequestURI + file.Filename

	err = copyFile(c, src, path)
	if err != nil {
		return nil
	}

	json := createFileJSON(0,  file.Filename, FILE)
	return c.JSON(http.StatusCreated, json)
}

func handleDELETEMethod(c echo.Context) error {
	path := dir + c.Request().RequestURI

	_, err := os.Stat(path)
	if err != nil {
		c.Response().WriteHeader(http.StatusNotFound)
		return nil
	}

	err = os.RemoveAll(path)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func handleHEADMethod(c echo.Context) error {
	path := dir + c.Request().RequestURI

	fi, err := os.Stat(path)
	if err != nil {
		c.Response().WriteHeader(http.StatusNotFound)
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}

	fileSize := strconv.FormatInt(fi.Size(), 10)

	c.Response().Header().Set("Connection", "Closed")
	c.Response().Header().Add(echo.HeaderContentLength, fileSize)
	c.Response().Header().Add(echo.HeaderContentType, http.DetectContentType(buffer[:n]))
	c.Response().Header().Add(echo.HeaderContentDisposition, "attachment; filename=\"" + fi.Name() + "\"")
	c.Response().Header().Add("File Server", "shine")
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func handlePOSTMethod(c echo.Context) error {
	currPath := dir + c.Request().RequestURI;
	nextPath := dir + c.Request().Header.Get("X-Copy-From")

	file, err := os.Open(currPath)
	if err != nil {
		c.Response().WriteHeader(http.StatusNotFound)
		return err
	}

	defer file.Close()

	err = copyFile(c, file, nextPath)
	if err != nil {
		c.Response().WriteHeader(http.StatusInternalServerError)
		return nil
	}
	c.Response().WriteHeader(http.StatusOK)
	return nil
}

func main() {
	e := echo.New()
	e.GET("*", handleGETMethod)
	e.PUT("*", handlePUTMethod)
	e.DELETE("*", handleDELETEMethod)
	e.HEAD("*", handleHEADMethod)
	e.POST("*", handlePOSTMethod)
	e.Logger.Fatal(e.Start(":1323"))
}