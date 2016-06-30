package main

import (
	"fmt"
	"gopkg.in/gcfg.v1"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"git.unixvoid.com/mfaltys/glogger"
)

type Config struct {
	Beaconclient struct {
		Loglevel   string
		Endpoint   string
		AuthFile   string
		HostDevice string
	}
}

var (
	config = Config{}
)

func main() {
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		fmt.Printf("Could not load config.gcfg, error: %s\n", err)
		return
	}
	// init logger
	if config.Beaconclient.Loglevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if config.Beaconclient.Loglevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if config.Beaconclient.Loglevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}

	credentials, _ := ioutil.ReadFile(config.Beaconclient.AuthFile)
	lines := strings.Split(string(credentials), "\n")
	beaconId := lines[0]
	beaconSec := lines[1]
	beaconAddress := "127.0.0.x"
	// TODO rant.. to resolve the ip we either get the host ip (if run directly on the host)
	// or (if run in docker) the dockerhost..
	// for instance if we run with --add-host dockerhost: <host ip>
	//   we can get the host ip into the docker container..
	//   regular os: --add-host dockerhost:`/sbin/ip route|awk '/default/ { print  $3}'`
	//   if used in ec2: --add-host dockerhost: `curl http://169.254.169.254/latest/meta-data/local-ipv4`
	if config.Beaconclient.HostDevice == "docker" {
		hostIp, err := net.LookupIP("dockerhost")
		if err != nil {

		}
		for _, s := range hostIp {
			beaconAddress = s.String()
		}
	} else {
		list, err := net.Interfaces()
		if err != nil {
			glogger.Error.Println("unknown beacon registration error occured:", err)
		}

		for _, iface := range list {
			if iface.Name == config.Beaconclient.HostDevice {
				addrs, err := iface.Addrs()
				if err != nil {
					panic(err)
				}
				for _, addr := range addrs {
					switch ip := addr.(type) {
					case *net.IPNet:
						if ip.IP.DefaultMask() != nil {
							ipNonCidr, _, _ := net.ParseCIDR(addr.String())
							beaconAddress = fmt.Sprintf("%s", ipNonCidr)
						}
					}
				}
			}
		}
	}

	// post to update registration
	postData := url.Values{}
	postData.Set("id", beaconId)
	postData.Add("sec", beaconSec)
	postData.Add("address", beaconAddress)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", config.Beaconclient.Endpoint, strings.NewReader(postData.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)

	if err != nil {
		glogger.Error.Println("unknown beacon registration error occured:", err)
	} else {
		switch resp.StatusCode {
		case 200:
			glogger.Info.Printf("beacon id %s updated to %s", beaconId, beaconAddress)
		case 403:
			glogger.Error.Println("beacon authenication failed")
		case 400:
			glogger.Error.Println("beacon id not found")
		default:
			glogger.Error.Println("unknown beacon registration error occured")
		}
	}
}
