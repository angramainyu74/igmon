package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
	"strconv"
)

func getBodyString( url string ) string{
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("unable to get "+url)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("unable to read body of "+url)
	}
	return string(body)
}

func downloadFile( url string ){
	parts := strings.Split(url,"/")
	name := parts[len(parts)-1]
	if _, err := os.Stat(name); os.IsNotExist(err) {
		out, err := os.Create(name)
		if err != nil {
			log.Fatal("unable to create file "+name)
		}
		defer out.Close()

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("unable to download "+url)
		}
		defer resp.Body.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Fatal("unable to write file "+name)
		}
		fmt.Println("Downloaded "+name)
	}else{
		fmt.Println("Already have "+name)
	}
}

func main() {
	napsec := 300
	argsWithoutProg := os.Args[1:]
	
	if len(argsWithoutProg) < 1 || len(argsWithoutProg) > 2{
		fmt.Println("Usage:")
		fmt.Println("  igmon <username> [timeoutsecs]")
		os.Exit(0)
	}
	username := argsWithoutProg[0]
	if len(argsWithoutProg) == 2 {
		n, err := strconv.Atoi(argsWithoutProg[1])
		if err != nil {
			log.Fatal("unable to parse sleep value from "+argsWithoutProg[1])
		}
		napsec = n
	}
	
	for ;;{
		sb := getBodyString("https://www.instagram.com/"+username+"/")
		
		r, _ := regexp.Compile("\"display_url\":\"([^\"]*)\"")
		matches := r.FindAllStringSubmatch(sb, -1)
		for _,i := range matches{
			downloadFile(i[1])
		}
		
		fmt.Println("checking sidecars")
		r, _ = regexp.Compile("\"__typename\":\"GraphSidecar\".*?\"shortcode\":\"([^\"]*)\"")
		matches = r.FindAllStringSubmatch(sb, -1)
		for _,i := range matches{
			fmt.Println("Sidecar "+i[1])
			sbs := getBodyString("https://www.instagram.com/p/"+i[1])
			rs, _ := regexp.Compile("\"display_url\":\"([^\"]*)\"")
			smatches := rs.FindAllStringSubmatch(sbs, -1)
			for _,si := range smatches{
				downloadFile(si[1])
			}
		}

		fmt.Println("checking videos")
		r, _ = regexp.Compile("\"__typename\":\"GraphVideo\".*?\"shortcode\":\"([^\"]*)\"")
		matches = r.FindAllStringSubmatch(sb, -1)
		for _,i := range matches{
			fmt.Println("Video "+i[1])
			sbv := getBodyString("https://www.instagram.com/p/"+i[1])
			rv, _ := regexp.Compile("property=\"og:video\" content=\"([^\"]*)\"")
			smatches := rv.FindAllStringSubmatch(sbv, -1)
			for _,vi := range smatches{
				downloadFile(vi[1])
			}
		}

		fmt.Printf("Sleeping for %d seconds.\n\n",napsec)
		time.Sleep(time.Duration(napsec) * time.Second)
	}
}