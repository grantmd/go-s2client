package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func normalizeName(name string) string {
	if name == "" {
		name = "Smart"
	}
	if []byte(name)[0] >= '0' && []byte(name)[0] <= '9' {
		name = "A" + name
	}
	return strings.Replace(strings.Replace(strings.Title(name), " ", "", -1), "_", "", -1)
}

func main() {
	// todo: support all OS
	file, err := ioutil.ReadFile(os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH") + "/Documents/StarCraft II/stableid.json")
	if err != nil {
		panic(err)
	}

	data := map[string]interface{}{}
	if err = json.Unmarshal(file, &data); err != nil {
		panic(err)
	}

	for key, dir := range map[string]string{
		"Units":     "unit",
		"Abilities": "ability",
		"Upgrades":  "upgrade",
		"Buffs":     "buff",
		"Effects":   "effect"} {
		var typeName = strings.Title(dir)
		if err = os.MkdirAll("sc2ids/"+dir, 0755); err != nil {
			panic(err)
		}

		f, err := os.Create("sc2ids/" + dir + "/" + dir + ".go")
		if err != nil {
			panic(err)
		}
		w := bufio.NewWriter(f)

		maxNameLen := 0
		items := [][]interface{}{}
		itemsNames := map[string]bool{}
		fmt.Fprint(w, "// DO NOT EDIT! Generated automatically\n")
		fmt.Fprint(w, "package "+dir+"\n\ntype "+typeName+" uint32\n\nconst (\n")
		for _, itemInterface := range data[key].([]interface{}) {
			item := itemInterface.(map[string]interface{})
			name := normalizeName(item["name"].(string))
			if itemsNames[name] != false {
				if item["friendlyname"] != nil {
					name = normalizeName(item["friendlyname"].(string))
					if itemsNames[name] != false {
						continue
					}
				} else {
					continue
				}
			}
			itemsNames[name] = true
			items = append(items, []interface{}{name, item["id"]})
			if len(name) > maxNameLen {
				maxNameLen = len(name)
			}
		}
		for _, item := range items {
			pad := strings.Repeat(" ", maxNameLen-len(item[0].(string)))
			fmt.Fprintln(w, "\t"+item[0].(string)+pad, typeName, "=", item[1])
		}
		fmt.Fprint(w, ")\n")

		if err = w.Flush(); err != nil {
			panic(err)
		}
		if err = f.Close(); err != nil {
			panic(err)
		}
	}
}
