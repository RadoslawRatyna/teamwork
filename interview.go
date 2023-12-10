// Package customerimporter reads from a CSV file and returns a sorted (data
// structure of your choice) of email domains along with the number of customers
// with e-mail addresses for each domain. This should be able to be ran from the
// CLI and output the sorted domains to the terminal or to a file. Any errors
// should be logged (or handled). Performance matters (this is only ~3k lines,
// but could be 1m lines or run on a small machine).
package customerimporter

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/exp/maps"
	"hash/fnv"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type UniqueValue[T comparable] map[T]interface{}

type EmailResult map[string]UniqueValue[uint32]

func (e EmailResult) CountDomain(domain string) int {
	return len(e[domain])
}

func findEmailFieldPosition(s *bufio.Reader) (uint, error) {
	line, err := s.ReadString('\n')
	if err != nil {
		return 0, err
	}

	if line == "" {
		return 0, errors.New("file is empty")
	}

	if line == "email" {
		return 0, nil
	}

	meta := strings.Split(line, ",")
	if len(meta) <= 1 {
		return 0, errors.New("missing 'email' field")
	}

	for i, v := range meta {
		if strings.ToLower(v) == "email" {
			return uint(i), nil
		}
	}

	return 0, errors.New("missing 'email' field")
}

func CountEmailDomains(path string) (EmailResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := bufio.NewReader(f)

	emailIndex, err := findEmailFieldPosition(s)
	if err != nil {
		return nil, err
	}

	res := make(map[string]UniqueValue[uint32])
	emailRegex := regexp.MustCompile(`^[\w\-\\.]+@([\w-]+\.)+[\w-]{2,}$`)

	recordCh := make(chan []byte)
	done := make(chan bool)

	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}

	wg.Add(1)

	go func() {
		for {
			line, err := s.ReadBytes('\n')
			if err != nil && errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				log.Printf("Failed to read line from the file. %s", err)
				continue
			}

			recordCh <- line
		}

		close(recordCh)
		done <- true
	}()

	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case record := <-recordCh:
					{
						data := bytes.Split(record, []byte(","))

						if len(data) <= 1 {
							log.Printf("Invalid record: %v\n", data)
							continue
						}

						email := data[emailIndex]
						if !emailRegex.Match(email) {
							log.Printf("Invalid email: %s", email)
							continue
						}

						hash := fnv.New32()
						hash.Write(email)
						emailHash := hash.Sum32()

						atIndex := bytes.Index(email, []byte("@"))
						domain := string(email[atIndex+1:])

						mutex.Lock()
						if res[domain] == nil {
							res[domain] = UniqueValue[uint32]{
								emailHash: nil,
							}
						} else {
							res[domain][emailHash] = nil
						}
						mutex.Unlock()
					}
				case <-done:
					wg.Done()
					return
				}
			}
		}()
	}

	wg.Wait()

	return res, nil
}

func SaveResultToFile(path string, result EmailResult) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	SaveResultToOutput(f, result)
}

func SaveResultToOutput(w io.Writer, result EmailResult) {
	domains := maps.Keys(result)
	sort.Strings(domains)

	for _, k := range domains {
		_, err := fmt.Fprintln(w, fmt.Sprintf("%s %d", k, result.CountDomain(k)))
		if err != nil {
			log.Printf("Failed to save element of result: %s\n", err)
		}
	}
}
