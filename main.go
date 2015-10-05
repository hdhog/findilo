package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/crackcomm/go-clitable"
	"github.com/parnurzeal/gorequest"
	"net"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type RIMP struct {
	XMLName xml.Name `xml:"RIMP"`
	SBSN    string   `xml:"HSI>SBSN"`
	SPN     string   `xml:"HSI>SPN"`
	PN      string   `xml:"MP>PN"`
	FWRI    string   `xml:"MP>FWRI"`
	HWRI    string   `xml:"MP>HWRI"`
}

func (r *RIMP) GetHW() string {
	var iloRevision = regexp.MustCompile(`\((.*)\)`)
	if len(r.PN) == 0 {
		return "N/A"
	}
	return iloRevision.FindAllStringSubmatch(r.PN, 1)[0][1]
}
func (r *RIMP) GetModel() string {
	if len(r.SPN) == 0 {
		return "N/A"
	}
	return strings.TrimSpace(r.SPN)
}
func (r *RIMP) GetFW() string {
	if len(r.FWRI) == 0 {
		return "N/A"
	}
	return strings.TrimSpace(r.FWRI)
}

const (
	ILO_PORT = 17988
)

var (
	IP_NETWORK   = ""
	IP_NETPARSED []string
)

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func IsOpen(host string, port int) bool {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false
	}
	conn, err := net.DialTimeout("tcp", tcpAddr.String(), 250*time.Millisecond)

	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func init() {
	flag.StringVar(&IP_NETWORK, "net", "", "Scan network, format 10.0.0.0/24")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(IP_NETWORK) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	ip, ipnet, err := net.ParseCIDR(IP_NETWORK)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	IP_NETPARSED = ips
}

func scan(in chan string, out chan string) {
	host := <-in
	for host != "" {
		if IsOpen(host, ILO_PORT) {
			out <- host
		}
		host = <-in
	}
	out <- "done"
}
func main() {

	fmt.Printf("Scaning %d hosts...\n\n", len(IP_NETPARSED))
	c := make(chan string, 50)
	in := make(chan string, 50)

	//Запуск воркеров
	for i := 0; i < 50; i++ {
		go scan(in, c)
	}

	for _, ip := range IP_NETPARSED {
		in <- ip
	}

	for i := 0; i < 50; i++ {
		in <- ""
	}

	var iloip = make([]string, 0)
	donecnt := 0
	for donecnt <= 49 {
		ip := <-c
		if ip != "done" {
			iloip = append(iloip, ip)
		} else {
			donecnt++
		}
	}

	request := gorequest.New()
	table := clitable.New([]string{"iLO IP Address", "iLO HW", "iLO FW", "Server S/N", "Server Model"})
	sort.Strings(iloip)

	for _, ipilo := range iloip {
		var info RIMP
		_, body, errs := request.Get(fmt.Sprintf("http://%s/xmldata?item=all", ipilo)).End()

		if errs != nil {
			panic(errs)
		}
		xml.Unmarshal([]byte(body), &info)
		table.AddRow(map[string]interface{}{
			"iLO IP Address": strings.TrimSpace(ipilo),
			"iLO HW":         info.GetHW(),
			"iLO FW":         info.GetFW(),
			"Server S/N":     strings.TrimSpace(info.SBSN),
			"Server Model":   info.GetModel(),
		})
	}

	table.Markdown = true
	table.Print()
	fmt.Println("")
	fmt.Printf("%d iLOs found on network target %s.\n", len(iloip), IP_NETWORK)
}
