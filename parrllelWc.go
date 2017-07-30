// word-counter counts the words in a text file.
package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "net/http"
    "regexp"
    "strings"
    "io/ioutil"
    "os"
    "golang.org/x/net/html"
    "sync"
)

var wordRegExp = regexp.MustCompile(`\pL+('\pL+)*`)
var wordCnts = make(map[string]int)

// cntWords counts the words in a file.
func cntWords(url string, c chan<- string, wg *sync.WaitGroup) {

    defer wg.Done()

    hdl, err := http.Get(url)
    if err != nil {
        fmt.Println("Failed to get the url :", url)
        log.Fatal(err)
        return
    }
    defer hdl.Body.Close()
    body, err := ioutil.ReadAll(hdl.Body)
    if err != nil {
        log.Fatal(err)
        return
    }
    responseString :=  string(body)
    line := strings.ToLower(responseString)
    words := wordRegExp.FindAllString(line, -1)
    for _, word := range words {
        c <- word
    }
    fmt.Printf("Done reading words from url %v\n", url)
}

func main() {
    flag.Parse()
    if len(flag.Args()) < 1 {
        log.Fatal("Missing filename argument")
    }

    url := flag.Arg(0)
    fmt.Printf("looking for url: %s \n", url)

    resp, _ := http.Get(url)
    defer resp.Body.Close()

    z := html.NewTokenizer(resp.Body)
    urls :=  make ([]string, 0)
    loop := true
    for {
        tt := z.Next()

        switch {
        case tt == html.ErrorToken:
            // End of the document, we're done
            loop = false
        case tt == html.StartTagToken:
            t := z.Token()

            isAnchor := t.Data == "a"
            if isAnchor {
                //~~~~~~~~~~~~~~~~~~~//
                // Get Tag Attribute //
                //~~~~~~~~~~~~~~~~~~~//
                for _, a := range t.Attr {
                    if a.Key == "href" {
                        pathUrl := url+ a.Val
                        urls = append(urls, pathUrl)
                    }
                }
            }
        }
        if (loop == false) {
            break
        }
    }

    var c chan string  = make(chan string)

    var wg sync.WaitGroup

    wg.Add(len(urls))

    // Run a go routine for each url
    for _, u := range urls {
        go cntWords(u, c, &wg)
    }

    //var mux sync.Mutex

    go func() {
        for word := range c {
            //mux.Lock()
            wordCnts[word]++
            //mux.Unlock()
        }
    }()

    wg.Wait()

    f,err := os.Create("result.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    w := bufio.NewWriter(f)

    // Print the word counts.
    for k,v := range wordCnts {
        //fmt.Printf("%s,%d\n", k, v)
        fmt.Fprintf(w, "%s,%d\n", k, v)
        w.Flush()
    }
}
