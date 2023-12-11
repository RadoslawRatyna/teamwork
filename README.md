# CustomerImporter

Simple CLI tool to counting a unique email domains from CSV file. It's job interview assignment for TeamWork company

# Usage

```
Usage of customerimporter:

    -f string
    Path to source CSV data
    
    -o string (Optional)
    Output path for result of customerimporter program
```

When flag '-o' is empty/unused, then result of program prints in terminal(STDOUT).

Result of CLI, is sorted ascending and presents in the following way:

```
    <domain>            <amount>
    another.mail.com    1
    example.com         1
    gmail.com           3
```

## Requirements for CSV file
Based on customer.csv file that I was given in email, I suppose this file structure:
- first line in CSV file includes names of columns
- one of column has name: email
- order of columns aren't matter
- other lines in file are records that I should be processed by my program

Example CSV file:
```
first_name,last_name,email,gender,ip_address
Bob,Smith,bob.smith@example.com,Male,38.194.51.128
```

# Build

```shell
    make build
```