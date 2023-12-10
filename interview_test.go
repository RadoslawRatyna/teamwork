package customerimporter

import (
	"bufio"
	"os"
	"strings"
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
			s := bufio.NewScanner(strings.NewReader(h))

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
			s := bufio.NewScanner(strings.NewReader(h))

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
		t.Fatalf("Invalid amount of example.com domain. Expected: 1, Actually: %d", domains.CountDomain("example.com"))
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
				t.Fatalf("Invalid amount of example.com domain. Expected: 1, Actually: %d", domains.CountDomain("example.com"))
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

func createTempDataFile(t *testing.T, data string) (f *os.File, err error) {
	f, err = os.CreateTemp(t.TempDir(), "result")
	if err != nil {
		return
	}

	_, err = f.WriteString(data)

	return
}
