package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	// help with tokenizing response - divide the page into small tokens
	"golang.org/x/net/html"
)

func getHref(t html.Token)(ok bool, href string){
	// range over the attributes of the token
	for _, a := range t.Attr{
		if a.Key == "href"{
			href = a.Val 
			ok = true
		}
	}

	return
}

// ch is chUrls - shortened for syntax
func crawl(url string, ch chan string, chFinished chan bool) {
	// make a request to a URL and the response will be in HTML
	res, err := http.Get(url)

	// defer is the keyword that you use if you want a function to run at the end of a particular function (towards the end of the callstack)
	defer func() {
		// publish to the finished channel, chFinished, that we're done processing this particular URL
		chFinished <- true
	// defer is a self calling function. like IIFE in JavaScript
	}()

	if err != nil {
		fmt.Println("error: failed to crawl:", url)
		return
	}

	// receive something in the body of the response 
	b := res.Body

	defer b.Close()

	// html method from x/net/html package
	// this will take b and divide it into small html tokens
	z := html.NewTokenizer(b)

	// for loop to process all of the tokens
	for {
		// goes over each of the html tokens 1 by 1
		tt := z.Next()

		switch {
		// if we get an ErrorToken that means the document has stopped proccessing - it's the end of the document
		case tt == html.ErrorToken:
			return
		case tt == html.StartTagToken:
			// put the value of the StartTagToken in a var called t
			t := z.Token()

		// check if t is an anchor tag
		isAnchor := t.Data == "a"
		// if it's not an <a> then we will continue
		if !isAnchor{
			continue
		}

		// if it is an <a> we'll pass it into getHref
		ok, url := getHref(t)

		// if we don't get a URL, just continue 
		if !ok {
			continue
		}

		// if ok is true it goes to this step
		// check if URL starts with http
		hasProto := strings.Index(url, "http") == 0
		if hasProto{
			// publish the URL to this channel - chUrls in main func
			ch <- url
		}
		}
	}
}

func main() {
	// pass multiple URLs to scrape and find unique URLs on those pages. those will be stored in foundUrls. bool is there for if it found URLs or not
	foundUrls := make(map[string]bool)
	// the URLs that the user will pass when the crawl function is called
	seedUrls := os.Args[1:]

	// output where all the URLs need to be scraped 
	// chUrls is all the unique links on a particular URL
	chUrls := make(chan string)
	// used when we're done processing a particular URL
	chFinished := make(chan bool)

	// range over seedUrls then pass them to the crawl function
	for _, url := range seedUrls{
		// called using a go routine
		go crawl(url, chUrls, chFinished)
	}

	// select case statement to subscribe to those channels
	for c := 0; c<len(seedUrls);{
		// when we get a response from chUrls, we want to perform a simultaneous action with <var name here>
		select{
		case url := <-chUrls:
			// once go crawl pulls out all the unique URLs, we'll store it in foundUrls map. true because we found a URL
			foundUrls[url] = true
		case <-chFinished:
			// when when the previous channel responds we want to go to the next channel -> c++
			c++
		}
	}

	// will print the amount unique URLs 
	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

	// range over over the found URLs map
	for url, _ := range foundUrls {
		fmt.Println("-" + url)
	}

	// the function where you receive a channel is where you have to close the channel. Not where you're writing something into the channel
	// have to close channels to remove deadlocks
	close(chUrls)
}