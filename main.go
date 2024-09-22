package main

import (
	"context"
	"dd/client"
	"dd/config"
	"dd/db"
	"dd/models"
	"dd/worker"
	"errors"
	"flag"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.Init()
	if err != nil {
		panic(err)
	}

	mysql, err := db.MakeMysqlConnect(cfg.Database)
	if err != nil {
		panic(err)
	}
	defer mysql.Close()

	requestPtr := flag.Int("requests", 1000, "number of requests")
	threadPtr := flag.Int("threads", 5, "number of threads")
	urlPtr := flag.String("url", "", "number of threads")

	pool := worker.NewWorkerPool(*threadPtr, false)
	pool.Start()

	go func() {
		for i := 0; i < *requestPtr; i++ {
			job := worker.NewJob(i, func(i interface{}, ctx context.Context) worker.JobRs {
				if d, ok := i.(int); ok {
					log.Print("Run job number: ", d)
					var proxy models.Proxy
					err = mysql.NewSelect().
						Model(&proxy).
						OrderExpr("RAND()").
						Limit(1).
						Scan(ctx)
					if err != nil {
						return worker.JobRs{Err: err}
					}
					proxyUrl := client.BuildURL(proxy)
					cli, err := client.CreateProxyClient(proxyUrl)
					if err != nil {
						return worker.JobRs{Err: err}
					}
					resp, err := cli.Get(*urlPtr)
					if err != nil {
						log.Fatalf("Failed to make request: %v", err)
					}
					return worker.JobRs{Value: resp.Status}
				}
				return worker.JobRs{Err: errors.New("data input no match")}
			})
			err = pool.AddJobNonBlocking(job)
			if err != nil {
				fmt.Printf("Failed to submit job %d: %v\n", i, err)
				break
			}
		}
	}()

	go func() {
		for result := range pool.Results() {
			if result.Err != nil {
				fmt.Printf("Job failed: %v\n", result.Err)
			} else {
				if _, ok := result.Value.(int); ok {
					log.Printf("Job %v finished\n", result.Value)
				} else {
					fmt.Printf("Result not match")
				}
			}
		}
	}()

	pool.Wait()
	fmt.Println("All jobs processed, worker pool shut down")
}
