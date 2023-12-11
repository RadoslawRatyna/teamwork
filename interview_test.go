package customerimporter

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestFindEmailFieldPosition(t *testing.T) {
	headers := []string{
		"first_name,last_name,email",
		"first_name,last_name,Email",
		"first_name,last_name,eMail",
		"first_name,last_name,emAil",
		"first_name,last_name,emaIl",
		"first_name,last_name,emaiL",
	}

	for _, h := range headers {
		t.Run(h, func(t *testing.T) {
			s := bufio.NewReader(strings.NewReader(h))

			position, err := findEmailFieldPosition(s)
			if err != nil {
				t.Fatal(err)
			}

			if position != 2 {
				t.Fatalf("Email field position is invalid. Expected: 2, Actually: %d", position)
			}
		})
	}
}

func TestFindEmailFieldPositionButMissingEmailField(t *testing.T) {
	headers := []string{
		"first_name,last_name",
		"",
		" ",
		",,,,",
		",ema il",
		",email ,",
		",email   ",
		"first_name,lastname\nemail",
	}

	for _, h := range headers {
		t.Run(h, func(t *testing.T) {
			s := bufio.NewReader(strings.NewReader(h))

			position, err := findEmailFieldPosition(s)
			if position != 0 || err == nil {
				t.Fatal("Find email, when field is missing")
			}
		})
	}
}

func TestCountEmailDomains(t *testing.T) {
	f, err := createTempDataFile(t, "email,ip\ntest@example.com,127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	domains, err := CountEmailDomains(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(domains) != 1 {
		t.Fatalf("Invalid number of email domains. Expected: 1, Actually: %d", len(domains))
	}

	if len(domains["example.com"]) != 1 {
		t.Fatalf("Invalid amount of example.com domain. Expected: 1, Actually: %d", domains.CountDomains("example.com"))
	}
}

func TestCountEmailDomainsWithDuplicates(t *testing.T) {
	contents := []string{
		`email,ip
test@example.com,127.0.0.1
test@example.com,127.0.0.1
test2@test.com,127.0.0.1`,
		`email,ip
test@example.com,127.0.0.1
test2@test.com,127.0.0.1
test@example.com,127.0.0.1`,
		`email,ip
test@example.com,127.0.0.1
test@test.com,127.0.0.1
test2@test.com,127.0.0.1
test@example.com,127.0.0.1`,
	}

	for _, c := range contents {
		t.Run(c, func(t *testing.T) {
			f, err := createTempDataFile(t, c)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			domains, err := CountEmailDomains(f.Name())
			if err != nil {
				t.Fatal(err)
			}

			if len(domains) != 2 {
				t.Fatalf("Invalid number of email domains. Expected: 2, Actually: %d", len(domains))
			}

			if len(domains["example.com"]) != 1 {
				t.Fatalf("Invalid amount of example.com domain. Expected: 1, Actually: %d", domains.CountDomains("example.com"))
			}
		})
	}
}

func TestCountEmailDomainsWithOneInvalidEmail(t *testing.T) {
	data := []string{
		`email,ip
test@example.com,127.0.0.1
,127.0.0.1`,
		`email,ip
test@example.com,127.0.0.1
invalid@@test.com,127.0.0.1`,
		`email,ip
test@example.com,127.0.0.1
invalid@!test.com,127.0.0.1`,
		`email,ip
test@example.com,127.0.0.1
invalidtest.com,127.0.0.1`,
	}

	for _, d := range data {
		t.Run(d, func(t *testing.T) {
			f, err := createTempDataFile(t, d)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			domains, err := CountEmailDomains(f.Name())
			if err != nil {
				t.Fatal(err)
			}

			if len(domains) != 1 {
				t.Fatalf("Invalid number of email domains. Expected: 1, Actually: %d", len(domains))
			}
		})
	}
}

func TestCountEmailDomainsButLoadDataFromRealFile(t *testing.T) {
	f, err := os.Open("./customers.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	domains, err := CountEmailDomains(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(domains) != 501 {
		t.Fatalf("Invalid number of email domains. Expected: 501, Actually: %d", len(domains))
	}
}

func BenchmarkCountEmailDomains(b *testing.B) {
	temp, err := os.CreateTemp(b.TempDir(), "benchmark")
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(temp)

	for i := 0; i < b.N; i++ {
		_, err := CountEmailDomains("./customers.csv")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestReadEmailFromRecord(t *testing.T) {
	type EmailRecord struct {
		record     string
		emailIndex uint
	}

	data := []EmailRecord{
		{
			"test@example.com",
			0,
		},
		{
			"Name,test@example.com",
			1,
		},
		{
			"Name,test@example.com,LastName",
			1,
		},
		{
			"Name,LastName,test@example.com",
			2,
		},
	}

	for _, d := range data {
		t.Run(d.record, func(t *testing.T) {
			email := readEmailFromRecord([]byte(d.record), d.emailIndex)

			if string(email) != "test@example.com" {
				t.Fatalf("Invalid email. Expected test@example.com, Actually: %s", email)
			}
		})
	}
}

func TestLoadLines(t *testing.T) {
	var (
		wg             sync.WaitGroup
		recordCh, done = make(chan []byte), make(chan bool)
		data           = "Dennis,Henry,dhenry2@hubpages.com,Male,155.75.186.217"
		reader         = bufio.NewReader(strings.NewReader(data))
	)

	go loadLines(reader, &wg, recordCh, done)

	wg.Wait()

	select {
	case record := <-recordCh:
		if string(record) != data {
			t.Fatalf("Invalid record data: %s", record)
		}
	}
}

func createTempDataFile(t *testing.T, data string) (f *os.File, err error) {
	f, err = os.CreateTemp(t.TempDir(), "result")
	if err != nil {
		return
	}

	_, err = f.WriteString(data)

	return
}
