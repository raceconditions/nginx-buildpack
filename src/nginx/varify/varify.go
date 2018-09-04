package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"fmt"
	"path/filepath"
	"encoding/json"
)

func main() {
	filename := os.Args[1]

	body, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Could not read config file: %s: %s", filename, err)
	}

	fileHandle, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Could not open config file for writing: %s", err)
	}
	defer fileHandle.Close()

	funcMap := template.FuncMap{
		"env": os.Getenv,
		"port": func() string {
			return os.Getenv("PORT")
		},
		"module": func(name string) string {
			return fmt.Sprintf("load_module %s.so;", filepath.Join(os.Getenv("NGINX_MODULES"), name))
		},
		"svcprop": func(args ...string) string {
			return getServiceProperty(args)
		},
	}

	t, err := template.New("conf").Funcs(funcMap).Parse(string(body))
	if err != nil {
		log.Fatalf("Could not parse config file: %s", err)
	}

	if err := t.Execute(fileHandle, nil); err != nil {
		log.Fatalf("Could not write config file: %s", err)
	}
}

func getServiceProperty(args []string) string {
	vcapservices := os.Getenv("VCAP_SERVICES")
	var services map[string][]interface{} 
	
	serviceType := args[0]
	serviceName := args[1]
	propKey := args[2]

	json.Unmarshal([]byte(vcapservices), &services)

	for i := 0; i < len(services[serviceType]); i++ {
		svc := services[serviceType][i].(map[string]interface{})
		if serviceName == svc["name"].(string) {
			if len(args) == 3 {
				prop := svc[propKey].(string)
				return prop
			} else if len(args) == 4 {
				subPropKey:= args[3]
				prop := svc[propKey].(map[string]interface{})
				return prop[subPropKey].(string)
			}
		}
	}
	return ""
}
