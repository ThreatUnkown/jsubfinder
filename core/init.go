package core

import (
	"fmt"
	"strconv"
	"sync"
)

var (
	Urls        []string
	Threads     int
	InputFile   string
	Url         string
	OutputFile  string
	Greedy      bool
	Debug       bool
	Crawl       bool
	FindSecrets bool
	Sig         string
	Silent      bool
)

func ExecSearch(concurrency int, outputFile string) {

	//fmt.Print(Urls)
	var data []UrlData
	var wg = sync.WaitGroup{}
	maxGoroutines := concurrency
	guard := make(chan struct{}, maxGoroutines)

	results := make(chan UrlData, len(Urls))
	for _, url := range Urls {
		guard <- struct{}{}
		wg.Add(1)
		go func(url string) {

			results <- NewURLData(url)
			<-guard
			wg.Done()
		}(url)
	}

	wg.Wait()
	close(guard)
	close(results)

	for result := range results {
		if result.Content != "" { //the urladdr will be blank if the page can't be reached. Thus don't add it.
			data = append(data, result)
		}
	}

	var newSubdomains []string
	var newSecrets []string
	if Debug {
		for _, url := range data {
			fmt.Println("url: " + url.UrlAddr.string)
			fmt.Println("\ttld: " + url.UrlAddr.tld)
			for _, js := range url.JSFiles {
				fmt.Println("\tjs: " + js.UrlAddr.string)
				fmt.Println("\t\tcontent length: " + strconv.Itoa(len(js.Content)))
				for _, subdomain := range js.subdomains {
					fmt.Println("\t\tsubdomain: " + subdomain)
					_, found := Find(newSubdomains, subdomain)
					if !found {
						newSubdomains = append(newSubdomains, subdomain)
					}
				}
				for _, secret := range js.secrets {
					fmt.Println("\t\tsecret: " + secret)
					_, found := Find(newSecrets, secret)
					if !found {
						newSecrets = append(newSecrets, secret+" of "+js.UrlAddr.string)
					}
				}
			}
		}
	} else {
		for _, url := range data {
			for _, js := range url.JSFiles {
				for _, subdomain := range js.subdomains {
					_, found := Find(newSubdomains, subdomain)
					if !found {
						fmt.Println(subdomain)
						newSubdomains = append(newSubdomains, subdomain)
					}
				}
				for _, secret := range js.secrets {
					_, found := Find(newSecrets, secret)
					if !found {

						newSecrets = append(newSecrets, secret+" of "+js.UrlAddr.string)
					}
				}
			}
		}
	}

	if PrintSecrets {
		for _, secret := range newSecrets {
			fmt.Println(secret)
		}
	}

	if outputFile != "" {
		SaveResults(outputFile, newSubdomains)
		SaveResults("secrets_"+outputFile, newSecrets)
	}
}
