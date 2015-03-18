package main

import(
	"os"
	"os/signal"
	"fmt"
	"spike"
	"time"
	"github.com/codegangsta/cli"
	"github.com/montanaflynn/stats"
)

var t0 time.Time
var samples []float64

func main() {
	t0  = time.Now()
	app := cli.NewApp()
	app.Name = "sping"
	app.Usage = "Pinging and latency measurement utility for Spike Engine 3 powered services."
	app.Flags = []cli.Flag {
		cli.StringFlag {
			Name: "interval",
			Value: "500ms",
			Usage: "The interval between two pings. Accepts a string as a sequence of decimal numbers, each with optional fraction and a unit suffix, such as '300ms', '-1.5h' or '2h45m'. Valid time units are 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'.",
		},
	}
	app.Action = func(c *cli.Context) {
		host := "127.0.0.1:8002"
		if len(c.Args()) > 0 {
			host = c.Args()[0]
		}

		// The interval
		interval, err := time.ParseDuration(c.String("interval"))
		if err != nil {
			panic(err)
		}

		// Connect to the service
		fmt.Println("Starting pinging a Spike Engine service", host)
		channel := new(spike.TcpChannel)
		channel.Connect(host, 8196)

		// Handle pong
		go onPing(channel)

		// Start pinging

		for {
			// Get the ping start
			now := int32(time.Now().Sub(t0).Nanoseconds() / 1000000)
			fmt.Print("Pinging ", host, " with 12 bytes of data: ")
			channel.Ping(now)
			time.Sleep(interval)
		}
		//go onPing(channel)
		// go channel.Ping()

		// So we don't exit the app
		//var input string
		//fmt.Scanln(&input)
	}

	// Hook CTRL+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
    	for sig := range c {
        	// sig is a ^C, handle it
    		fmt.Println("CTRL-C", sig, "received")
    		min := stats.Min(samples)
    		max := stats.Max(samples)
    		mean := stats.Mean(samples)
    		median := stats.Median(samples)
    		varp := stats.VarP(samples)
    		p75 := stats.Percentile(samples, 75)
    		p90 := stats.Percentile(samples, 90)
    		p95 := stats.Percentile(samples, 95)
    		p99 := stats.Percentile(samples, 99)

    		fmt.Println()
    		fmt.Println("Ping statistics:")
    		fmt.Println("   Min:      ", min, "ms.")
    		fmt.Println("   Max:      ", max, "ms.")
    		fmt.Println("   Mean:     ", mean, "ms.")
    		fmt.Println("   Median:   ", median, "ms.")
    		fmt.Println("   Variance: ", varp, "ms.")
    		fmt.Println()
    		fmt.Println("Percentiles:")
    		fmt.Println("   75th: ", p75, "ms.")
    		fmt.Println("   90th: ", p90, "ms.")
    		fmt.Println("   95th: ", p95, "ms.")
    		fmt.Println("   99th: ", p99, "ms.")
    		os.Exit(0)
    	}
	}()

	// Run the application
	app.RunAndExitOnError()
}

func onPing(channel *spike.TcpChannel){
	for{
    	msg := <- channel.OnPing
    	rtt := int32(time.Now().Sub(t0) / 1000000) - msg.Time
    	fmt.Println(rtt, " ms.")
    	samples = append(samples, float64(rtt))
	}
}
