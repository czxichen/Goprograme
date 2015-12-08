package tool

import (
	"io/ioutil"
	"net/http"
	"text/template"
)

func ExecuteReplace(w http.ResponseWriter, temp_path string, funcs map[string]string) error {
	T := template.New("")
	buf, err := ioutil.ReadFile(temp_path)
	if err != nil {
		return err
	}
	T, err = T.Parse(string(buf))
	if err != nil {
		return err
	}
	err = T.Execute(w, funcs)
	if err != nil {
		return err
	}
	return nil
}
