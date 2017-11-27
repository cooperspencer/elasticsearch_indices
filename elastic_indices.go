package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"os"
	"strings"
	"regexp"
	"time"
	"log"
	"github.com/spf13/viper"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if strings.Contains(b, a) {
			return false
		}
	}
	return true
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	viper.SetConfigFile("./config.yml")
	err := viper.ReadInConfig()
	check(err)

	elasticurl := viper.GetString("elasticurl")
	protocol := viper.GetString("protocol")
	port := viper.GetString("port")

	elasticsearch := protocol + "://" + elasticurl + ":" + port

	resp, err := http.Get(elasticsearch + "/_cat/indices?v&pretty")
	if err != nil {
		fmt.Println("%s", err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("%s", err)
			os.Exit(1)
		}
		//fmt.Println("%s", string(contents))
		indices := strings.Split(string(contents), "\n")

		indices_tomorrow := []string{}
		index_map := make(map[string]string)
		for _, index := range indices[1:len(indices)-1] {
			index = strings.Fields(index)[2]

			match, _ := regexp.MatchString("[A-Z,a-z,-_]{1,}[-.][0-9]{4,}[-.][0-9]{2,}[-.][0-9]{2,}", index)
			if match == true {
				//fmt.Println(strings.Split(index, "-")[0])
				re_name := regexp.MustCompile("^[^0-9]*")
				re_date := regexp.MustCompile("[^a-z,A-Z,-]([0-9,-.]+)")
				name := re_name.FindAllString(index, -1)[0]
				date := re_date.FindAllString(index, -1)[0]
				if stringInSlice(name, indices_tomorrow) {
					indices_tomorrow = append(indices_tomorrow, index)
					//fmt.Println(indices_tomorrow)
					index_map[name] = date
				}
			}
		}
		//fmt.Println(index_map)
		tomorrow := time.Now().AddDate(0, 0, +1)
		for key, value := range index_map {
			tomorrow_index := ""
			if strings.Contains(value, "-") {
				tomorrow_index = key + tomorrow.Format("2006-01-02")
			} else {
				tomorrow_index = key + tomorrow.Format("2006.01.02")
			}
			fmt.Println(tomorrow_index)
			url := elasticsearch + "/" + tomorrow_index + "?pretty"
			fmt.Println(url)
			client := &http.Client{}
			request, err := http.NewRequest(http.MethodPut,url, strings.NewReader(""))
			response, err := client.Do(request)
			if err != nil {
				log.Fatal(err)
			} else {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("The calculated length is:", len(string(contents)), "for the url:", url)
				fmt.Println("   ", response.StatusCode)
				hdr := response.Header
				for key, value := range hdr {
					fmt.Println("   ", key, ":", value)
				}
				fmt.Println(string(contents))
			}
		}
	}
}