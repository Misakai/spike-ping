package main

import(
	"os"
	"os/signal"
	"fmt"
	"spike"
	"time"
	"strings"
	"strconv"
	"github.com/codegangsta/cli"
	"github.com/montanaflynn/stats"
)

func main() {
	app := cli.NewApp()
	app.Name = "sping"
	app.Usage = "Pinging and latency measurement utility for Spike Engine 3 powered services."
	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "interval",
			Value: "250ms",
			Usage: "Sets the interval between two pings. Accepts a string as a sequence of decimal numbers, each with optional fraction and a unit suffix, such as '300ms', '-1.5h' or '2h45m'. Valid time units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'.",
		},
		cli.StringFlag {
			Name: "out",
			Value: "",
			Usage: "Sets the output file to write out the latency values instead of calculating them in-memory.",
		},
	}
	app.Action = func(c *cli.Context) {
		// Recover and print a nicer message
		defer func() {
        	if r := recover(); r != nil {
            	fmt.Println("Error:", r)
        	}
    	}()

		// Host and port
		host := "127.0.0.1:8002"
		if len(c.Args()) > 0 {
			host = c.Args()[0]
		}
		if !strings.Contains(host, ":"){
			host += ":80"
		}

		// Variables we need
		var samples []float64
		var out *os.File
		t0 := time.Now()

		// The interval
		interval, err := time.ParseDuration(c.String("interval"))
		if err != nil {
			panic(err)
		}

		// Output file
		if c.String("out") != "" {
			out, err = os.OpenFile(c.String("out"), os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				panic(err)
			}
		}

		// Connect to the service
		fmt.Println("Starting pinging a Spike Engine service", host)
		channel := new(spike.TcpChannel)
		channel.Connect(host, 8196)

		// Handle pong
		go func (){
			for{
		    	msg := <- channel.OnPing
		    	rtt := int32(time.Now().Sub(t0) / 1000000) - msg.Time
		    	fmt.Println("Pinging", host, "with 12 bytes of data:", rtt, "ms.")
		    	if out != nil {
		    		if _, err = out.WriteString(strconv.Itoa(int(rtt)) + "\r\n"); err != nil {
					    panic(err)
					}
		    	} else {
		    		samples = append(samples, float64(rtt))
		    	}
			}
		}()


		// Hook CTRL+C
		schan := make(chan os.Signal, 1)
		signal.Notify(schan, os.Interrupt)
		go func(){
	    	for sig := range schan {
	        	// sig is a ^C, handle it
	    		fmt.Println("CTRL-C", sig, "received")
	    		if out != nil {
	    			os.Exit(0)
	    		}

	    		min := stats.Min(samples)
	    		max := stats.Max(samples)
	    		mean := stats.Mean(samples)
	    		median := stats.Median(samples)
	    		varp := stats.VarP(samples)
	    		p1 := stats.Percentile(samples, 1)
	    		p25 := stats.Percentile(samples, 25)
	    		p50 := stats.Percentile(samples, 50)
	    		p75 := stats.Percentile(samples, 75)
	    		p90 := stats.Percentile(samples, 90)
	    		p95 := stats.Percentile(samples, 95)
	    		p99 := stats.Percentile(samples, 99)

	    		fmt.Println()
	    		fmt.Println("Ping statistics:")
	    		fmt.Println("   Samples:  ", len(samples), "events")
	    		fmt.Println("   Min:      ", min, "ms.")
	    		fmt.Println("   Max:      ", max, "ms.")
	    		fmt.Println("   Mean:     ", mean, "ms.")
	    		fmt.Println("   Median:   ", median, "ms.")
	    		fmt.Println("   Variance: ", varp, "ms.")
	    		fmt.Println()
	    		fmt.Println("Percentiles:")
	    		fmt.Println("    1st: ", p1, "ms.")
	    		fmt.Println("   25th: ", p25, "ms.")
	    		fmt.Println("   50th: ", p50, "ms.")
	    		fmt.Println("   75th: ", p75, "ms.")
	    		fmt.Println("   90th: ", p90, "ms.")
	    		fmt.Println("   95th: ", p95, "ms.")
	    		fmt.Println("   99th: ", p99, "ms.")
	    		os.Exit(0)
	    	}
		}()

		// Ping loop
		for {
			// Get the ping start
			now := int32(time.Now().Sub(t0).Nanoseconds() / 1000000)
			channel.Ping(now)
			time.Sleep(interval)
		}
	}

	// Run the application
	app.RunAndExitOnError()
}
