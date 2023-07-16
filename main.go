package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"strings"

	flags "github.com/jessevdk/go-flags"
	gochanbroker "github.com/krikus/go-chan-broker"
)

type options struct {
	Cname        []string     `long:"cname" description:"Pass a CNAME entry that domain needs to set as well"`
	IpParser     func(string) `long:"ip" description:"Pass an IP range that domain needs to be in"`
	IpCmp        []func(net.IP) bool
	TLS          bool   `long:"tls" description:"Check if domain has valid certificate"`
	Verbose      bool   `short:"v" long:"verbose" description:"Show more info on the console (same as Debug)"`
	Debug        bool   `short:"d" long:"debug" description:"Show more info on the console(same as Verbose)"`
	File         string `short:"f" long:"file" description:"File path with all domains to be checked.\nEach domain should be in a new line"`
	PrintValid   bool   `short:"p" long:"print-valid" description:"Prints valid domains"`
	PrintInvalid bool   `short:"i" long:"print-invalid" description:"Prints invalid domains"`
	Concurrency  int    `short:"c" long:"concurrency" description:"How many checks to perform in the same" default:"100"`
	ExitCode     bool   `short:"e" long:"exit" description:"Exit with code upon detecting invalid domain"`
}

var opts options

func debug(msg string) {
	if opts.Verbose || opts.Debug {
		fmt.Println(msg)
	}
}

func validate(domain string) bool {
	if opts.TLS == true {
		addr := strings.Join([]string{domain, "443"}, ":")
		conn, err := tls.Dial("tcp", addr, nil)
		debug(fmt.Sprintf("[%s] Checking for TLS errors %v", domain, err))

		if err != nil {
			debug(fmt.Sprintf("[%s] TLS checkup failed", domain))
			return false
		}
		conn.Close()
	}

	setTrue := false

	if len(opts.IpCmp) > 0 {
		ip, _ := net.LookupIP(domain)
		for _, netIp := range ip {
			for _, cmp := range opts.IpCmp {
				if cmp(netIp) {
					setTrue = true
					break
				}
			}
		}

		if setTrue == false {
			debug(fmt.Sprintf("[%s] IP checkup failed %v", domain, ip))
		}
	}

	if len(opts.Cname) > 0 {
		cname, _ := net.LookupCNAME(domain)
		debug(fmt.Sprintf("[%s] Checking CNAME %s agains %s", domain, cname, opts.Cname))

		for _, itemCname := range opts.Cname {
			if cname == itemCname {
				setTrue = true
			}
		}
		if setTrue == false {
			debug(fmt.Sprintf("[%s] CNAME checkup failed", domain))
		}
	}

	return setTrue || (len(opts.IpCmp) == 0 && len(opts.Cname) == 0)
}

func main() {
	opts.IpParser = func(ip string) {
		ipRange := strings.Split(ip, "/")
		if len(ipRange) == 2 {
			debug(fmt.Sprintf("Adding IP range to check pool: %s", ip))
			_, ipnet, err := net.ParseCIDR(ip)
			if err != nil {
				panic(fmt.Sprintf("Error during parsing the IP range %s - %v", ip, err))
			}
			opts.IpCmp = append(opts.IpCmp, ipnet.Contains)
		} else {
			debug(fmt.Sprintf("Adding IP to check pool: %s", ip))
			opts.IpCmp = append(opts.IpCmp, func(ipIn net.IP) bool {
				return ip == ipIn.String()
			})
		}
	}

	args, err := flags.Parse(&opts)

	if err != nil {
		panic(err)
	}

	debug(fmt.Sprintf("Cname: %v", opts.Cname))
	for i := range opts.Cname {
		if opts.Cname[i][len(opts.Cname[i])-1:] != "." {
			opts.Cname[i] = strings.ToLower(fmt.Sprintf("%s.", opts.Cname[i]))
			debug(fmt.Sprintf("Added missing dot at the end of the Cname parameter %v", opts.Cname[i]))
		} else {
			opts.Cname[i] = strings.ToLower(opts.Cname[i])
		}
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

	broker := gochanbroker.CreateChanBroker(opts.Concurrency, validate)

	go func() {
		for iterator.hasNext() {
			domainName := iterator.getNext()
			broker.AddJob(domainName)
		}

		broker.Finalize()
	}()

	for result := range broker.GetResultsChan() {
		debug(fmt.Sprintf("Got result %v", result))
		if result.Result {
			debug(fmt.Sprintf("%s is valid", result.Key))
			if opts.PrintValid {
				fmt.Println(result.Key)
			}
		} else {
			debug(fmt.Sprintf("%s is invalid", result.Key))
			if opts.PrintInvalid {
				fmt.Println(result.Key)
			}
			if opts.ExitCode {
				os.Exit(1)
			}
		}
	}
}
