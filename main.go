package main

import (
	"fmt"
	"sync"
	"golang.org/x/net/html"
	"net/http"
)


type Fetcher interface{
	Fetch(url string) (body string, urls []string, err error)
}


type SafeCache struct{
	mu sync.Mutex
	cache map[string]bool
}

func NewSafeCache()  *SafeCache{
	return &SafeCache{cache: make(map[string]bool)}
}


func (c *SafeCache) Exists(url string) bool{
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache[url] {
		return true
	}

	c.cache[url] =  true
	return false
}

func Crawl(url string, depth int, fetcher Fetcher, cache *SafeCache, wg *sync.WaitGroup){

	defer wg.Done()

	if depth <= 0 || cache.Exists(url){
		return 
	}

	body, urls, err :=  fetcher.Fetch(url)
	if err != nil{
		fmt.Println(err)
		return
	}

	fmt.Printf("found : %s %q\n", url , body)
	for _, u := range urls {
		wg.Add(1)
		go Crawl(u , depth -1 , fetcher, cache, wg)
	}

}

func main() {
	cache := NewSafeCache()
	var wg sync.WaitGroup
	wg.Add(1)
	go Crawl("http://golang.org/", 2, RealFetcher{}, cache, &wg)
	wg.Wait()
	}


type RealFetcher struct{}


func (f RealFetcher) Fetch(urlStr string) (string, [] string , error){
	resp, err := http.Get(urlStr)
	if err != nil{
		return "", nil , fmt.Errorf("error fetching URL %s: %v", urlStr, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		return "", nil, fmt.Errorf("HTTP error %d for URL %s", resp.StatusCode, urlStr)
	}

	doc, err := html.Parse(resp.Body)

	if err != nil{
		return "", nil , fmt.Errorf("error parsing HTML of %s: %v", urlStr, err)
	}

	

	var urls []string
	var crawler func(*html.Node)
	crawler = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a"{
			for _, a := range n.Attr{
				if a.Key == "href"{
					urls = append(urls, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling{
			crawler(c)
		}
	}
	crawler(doc)
	return urlStr, urls, nil
}
