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
)

type UniqueValue[T comparable] map[T]interface{}

type EmailResult map[string]UniqueValue[uint32]

func (e EmailResult) CountDomain(domain string) int {
	return len(e[domain])
}

func findEmailFieldPosition(s *bufio.Scanner) (uint, error) {
	s.Scan()
	line := s.Text()
	if line == "" {
		return 0, errors.New("file is empty")
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

	s := bufio.NewScanner(f)

	emailPosition, err := findEmailFieldPosition(s)
	if err != nil {
		return nil, err
	}

	res := make(map[string]UniqueValue[uint32])
	emailRegex := regexp.MustCompile(`^[\w\-\\.]+@([\w-]+\.)+[\w-]{2,}$`)

	for s.Scan() {
		buf := bytes.Split(s.Bytes(), []byte(","))
		if len(buf) <= 1 {
			log.Printf("Invalid record: %v\n", buf)
			continue
		}

		email := buf[emailPosition]
		if !emailRegex.Match(email) {
			log.Printf("Invalid email: %s", email)
			continue
		}

		hash := fnv.New32()
		hash.Write(email)

		domain := string(bytes.Split(email, []byte("@"))[1])

		if res[domain] == nil {
			res[domain] = UniqueValue[uint32]{
				hash.Sum32(): nil,
			}
		} else {
			res[domain][hash.Sum32()] = nil
		}
	}

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
