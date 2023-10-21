package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

var mutex = sync.Mutex{}

func main() {
	//TODO timeout?
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	//runtime.GOMAXPROCS

	logger := log.New(os.Stdout, "", 0)

	var path string
	flag.StringVar(&path, "f", "", "File path.")
	flag.Usage = func() {
		instruction()
	}
	flag.Parse()

	if path == "" {
		logger.Print("missing args")
		instruction()
		ctxCancel()
		os.Exit(1)
	}

	startTime := time.Now()
	fmt.Printf("Start %v \n", startTime)

	workerPoolSize := 100
	err := run(ctx, logger, path, workerPoolSize, http.DefaultClient)
	if err != nil {
		logger.Fatal(err)
	}

	endTime := time.Now()
	fmt.Printf("End %v \n", endTime)
	fmt.Printf("Duration %v \n", time.Since(startTime))
}

func run(ctx context.Context, logger logger, path string, size int, httpClient httpClient) error {
	if size <= 0 {
		return errors.New("invalid pool size")
	}
	if path == "" {
		return errors.New("invalid path")
	}
	if logger == nil || httpClient == nil {
		return errors.New("invalid params")
	}

	chanURLs := make(chan string, size)

	go getURLs(ctx, chanURLs, path)

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	_, err := fmt.Fprintln(w, "URL\tProcessingTime\tCODE\tSIZE\tERROR\t")
	if err != nil {
		return err
	}

	wg := &sync.WaitGroup{}
	runWorkerPool(ctx, logger, size, wg, chanURLs, w, httpClient)

	wg.Wait()
	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func runWorkerPool(ctx context.Context, logger logger, size int, wg *sync.WaitGroup, chanURLs <-chan string, w io.Writer, httpClient httpClient) {
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range chanURLs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				size, code, time, err := getData(httpClient, u)
				err = printLine(size, code, time, err, u, w)
				if err != nil {
					logger.Print(err.Error())
				}
			}
		}()
	}
}

func printLine(size, code int, time time.Time, err error, u string, w io.Writer) error {
	sizeStr := strconv.Itoa(size)
	errStr := "-"
	codeStr := strconv.Itoa(code)
	if err != nil {
		errStr = err.Error()
		sizeStr = "-"
		codeStr = "-"
	}

	line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t", u, time, codeStr, sizeStr, errStr)
	mutex.Lock()
	defer mutex.Unlock()
	_, err = fmt.Fprintln(w, line)
	if err != nil {
		return errors.New("error print line with")
	}
	return nil
}

func getURLs(ctx context.Context, chanURLs chan<- string, path string) {
	defer close(chanURLs)
	//TODO valid path?
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		url := strings.TrimSpace(scanner.Text())
		chanURLs <- url
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func getData(httpClient httpClient, u string) (int, int, time.Time, error) {
	_, err := url.Parse(u)
	if err != nil {
		return 0, 0, time.Now(), err
	}
	rsp, err := httpClient.Get(u)
	if err != nil {
		return 0, 0, time.Now(), err
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return 0, rsp.StatusCode, time.Now(), err
	}
	return len(body), rsp.StatusCode, time.Now(), nil
}

func instruction() {
	fmt.Printf("Usage of our program: \n")
	fmt.Printf("./go_office -f path/to/file\n")
}

type httpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type logger interface {
	Print(v ...any)
}
