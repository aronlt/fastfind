package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

var BUFFERSIZE = 1024

func IsEmpty(name string) (bool, error) {
	if name == "" {
		return false, errors.New("empty check path")
	}
	f, err := os.Open(name)
	if err != nil {
		log.Printf("can't open fileutil:%s, error:%s\n", name, err.Error())
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func IsExist(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		log.Printf("check fileutil:%s exist fail, error:%s\n", path, err.Error())
		return false, err
	}
}

func IsFile(path string) (bool, error) {
	ret, err := IsExist(path)
	if err != nil {
		return false, err
	}
	if ret == false {
		return false, err
	}
	status, err := os.Stat(path)
	if err != nil || status.Mode().IsRegular() == false {
		return false, err
	}
	return true, nil
}

func Copy(src string, dst string, force bool) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return errors.New(src + " is not a regular fileutil.")
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if !force && err == nil {
		return errors.New(dst + " is already exists.")
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}

func HashFileSha1(filePath string) (string, error) {
	var returnSHA1String string

	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}
	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}
	hashInBytes := hash.Sum(nil)
	//Convert the bytes to a string
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil

}

func HashFileMd5(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	md5String := md5.New()

	_, err = io.Copy(md5String, f)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(md5String.Sum(nil)), nil
}

func CreateFile(filepath string) error {
	_, err := os.OpenFile(filepath, os.O_CREATE, 0755)
	if err != nil {
		log.Printf("can't create fileutil:%s, error:%s\n", filepath, err.Error())
	}
	return err
}

func WriteContent(filepath string, content string) error {
	f, err := os.Create(filepath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		log.Printf("write fileutil:%s content fail, error:%s\n", filepath, err.Error())
		return err
	}
	return nil
}

func ReadContent(path string) ([]byte, error) {
	dat, err := ioutil.ReadFile(path)
	return dat, err
}

func Files(path string) []string{
	files := make([]string, 0)
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return files
	}

	for _, f := range infos {
		// 跳过被挂载的自身目录
		files = append(files, path + "/"+ f.Name())
	}
	return files
}