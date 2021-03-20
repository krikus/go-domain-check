package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
)

type options struct {
	Cname string `long:"cname" description:"Pass a CNAME entry that domain needs to set as well"`
	// Caa string `long:"caa" description:"Pass a CAA entry that domain needs to set as well"`
	TLS bool `long:"tls" description:"Check if domain has valid certificate"`
	Verbose bool `short:"v" long:"verbose" description:"Show more info on the console"`
	File string `short:"f" long:"file" description:"File path with all domains to be checked.\nEach domain should be in a new line"`
	PrintValid bool `short:"p" long:"print-valid" description:"Prints valid domains"`
	PrintInvalid bool `short:"i" long:"print-invalid" description:"Prints invalid domains"`
	ExitCode bool `short:"e" long:"exit" description:"Exit with code upon detecting invalid domain"`
}

var opts options;

func debug(msg string) {
	if opts.Verbose {
		fmt.Println(msg)
	}
}

func validate(domain string) bool {
	if opts.TLS == true {
		addr := strings.Join([]string{domain, "443"}, ":");
		conn, err := tls.Dial("tcp", addr, nil);
		debug(fmt.Sprintf("[%s] Checking for TLS errors %v", domain, err))

		if err != nil {
			debug(fmt.Sprintf("[%s] TLS checkup failed", domain))
			return false;
		}
		conn.Close();
	}

	if opts.Cname != "" {
		cname, _ := net.LookupCNAME(domain);
		debug(fmt.Sprintf("[%s] Checking CNAME %s agains %s", domain, cname, opts.Cname))
		if cname != opts.Cname {
			debug(fmt.Sprintf("[%s] CNAME checkup failed", domain))
			return false
		}
	}

	return true;
}

func main() {
	args, err := flags.Parse(&opts);

	if err != nil {
		panic(err);
	}

	if (opts.Cname != "" && opts.Cname[len(opts.Cname)-1:] != ".") {
		opts.Cname = fmt.Sprintf("%s.", opts.Cname);
		debug("Added missing dot at the end of the Cname parameter")
	}

	var iterator StringIterator
	if opts.File != "" {
		file, err := os.Open(opts.File)
		if err != nil {
			panic(fmt.Sprintf("Error during opening the file %s - %v", opts.File, err))
		}
		defer file.Close()
		iterator = CreateReaderIterator(file)
	
	} else if len(args) == 0 {
		iterator = CreateReaderIterator(os.Stdin)
	} else {
		iterator = CreateArrayIterator(args)
	}

	for iterator.hasNext() {
		element := iterator.getNext()
		if validate(element) {
			debug(fmt.Sprintf("%s is valid", element))
			if (opts.PrintValid) {
				fmt.Println(element)
			}
		} else {
			debug(fmt.Sprintf("%s is invalid", element))
			if (opts.PrintInvalid) {
				fmt.Println(element)
			}
			if (opts.ExitCode) {
				os.Exit(1)
			}
		}
	}
}
