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
	"runtime"
	"sort"
	"strings"
	"sync"
)

var emailRegex = regexp.MustCompile(`^[\w\-\\.]+@([\w-]+\.)+[\w-]{2,}$`)

type UniqueValue[T comparable] map[T]interface{}

type EmailResult map[string]UniqueValue[uint32]

func (e EmailResult) CountDomain(domain string) int {
	return len(e[domain])
}

func findEmailFieldPosition(s *bufio.Reader) (uint, error) {
	line, err := s.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
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

	reader := bufio.NewReader(f)

	emailIndex, err := findEmailFieldPosition(reader)
	if err != nil {
		return nil, err
	}

	domains := make(map[string]UniqueValue[uint32])
	recordCh, done := make(chan []byte), make(chan bool)
	wg := sync.WaitGroup{}
	mutex := sync.Mutex{}

	wg.Add(1)
	go loadLines(reader, &wg, recordCh, done)

	coresNum := runtime.NumCPU()

	for i := 0; i < coresNum; i++ {
		go func() {
			for {
				select {
				case record := <-recordCh:
					{
						addEmailDomain(&wg, &mutex, record, emailIndex, domains)
					}
				case <-done:
					break
				}
			}
		}()
	}

	wg.Wait()

	return domains, nil
}

func loadLines(r *bufio.Reader, wg *sync.WaitGroup, recordCh chan []byte, done chan bool) {
	for {
		line, err := r.ReadBytes('\n')
		if line != nil && !bytes.Equal(line, []byte{}) {
			wg.Add(1)
			recordCh <- line
		}

		if err != nil && errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			log.Printf("Failed to read line from the file. %s", err)
			continue
		}
	}

	wg.Done()
	done <- true
}

func readEmail(data []byte, emailIndex uint) []byte {
	count := bytes.Count(data, []byte(","))

	if count == 0 {
		return data
	}

	first, last := bytes.Index(data, []byte(",")), bytes.LastIndex(data, []byte(","))

	if emailIndex == 0 || first == last {
		return data[0:first]
	} else if emailIndex == uint(count) {
		return data[last:]
	}

	return readEmail(data[first+1:last], emailIndex-1)
}

func addEmailDomain(wg *sync.WaitGroup, mutex *sync.Mutex, record []byte, emailIndex uint, res map[string]UniqueValue[uint32]) {
	defer wg.Done()

	if record == nil {
		return
	}

	email := readEmail(record, emailIndex)
	if !emailRegex.Match(email) {
		log.Printf("Invalid email: %s", email)
		return
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
