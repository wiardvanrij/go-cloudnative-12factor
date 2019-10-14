package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/kelseyhightower/envconfig"
)

//Redis struct includes the client and healthcheck info
type Redis struct {
	Client             *redis.Client
	Pong               bool
	ErrorCounter       int
	ExponentialBackoff bool
}

//Settings struct to define base settings
type Settings struct {
	RecheckAmount   int    `default:"20"`
	HealthCheckTime int    `default:"500"`
	RedisHost       string `default:"localhost:6379"`
}

func main() {

	// Init of the env settings based on the struct settings
	var settings Settings
	err := envconfig.Process("counter", &settings)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Welcome the user...
	fmt.Printf("Redishost: %s \n", settings.RedisHost)
	fmt.Printf("Healthcheck every %d ms\n", settings.HealthCheckTime)
	fmt.Printf("and waiting %d times when redis is down before trying again\n", settings.RecheckAmount)
	fmt.Println("Good flight.\n\n")

	// Redis init definition
	redisObject := redisInit(settings.RedisHost)
	// Close on bye bye
	defer redisObject.Client.Close()

	// Main ticker loop to check Redis health status
	ticker := time.NewTicker(time.Millisecond * time.Duration(settings.HealthCheckTime)).C
	go func() {
		for {
			select {
			case <-ticker:

				// If we know Redis is down, ignore the Ping/Pong test for X ticks
				if redisObject.ExponentialBackoff == true {
					settings.RecheckAmount = settings.RecheckAmount * 2

					// Dont exceed 5 minutes (30.000 ms)
					if settings.RecheckAmount*settings.HealthCheckTime > 300000 {
						settings.RecheckAmount = 300000 / settings.HealthCheckTime
					}
					// Unset this, so we don't get stuck in a loop. Our healthcheck will set this var correctly again.
					redisObject.ExponentialBackoff = false
				}

				if redisObject.ErrorCounter > 0 && redisObject.ErrorCounter < settings.RecheckAmount {
					fmt.Printf("Redis is unavailable, trying again in %d ticks\n", settings.RecheckAmount-redisObject.ErrorCounter)
					redisObject.ErrorCounter++
					continue
				}

				// Let's play ping/pong with Redis
				redisObject.ping()
			}
		}
	}()

	// And just start a http listen/server
	http.HandleFunc("/counter", redisObject.handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func redisInit(redisHost string) *Redis {
	redisObject := new(Redis)
	redisObject.Client = redis.NewClient(&redis.Options{
		Addr: redisHost,
	})
	redisObject.Pong = true
	redisObject.ErrorCounter = 0
	redisObject.ExponentialBackoff = false

	return redisObject
}

func (redisObject *Redis) handler(w http.ResponseWriter, r *http.Request) {

	// Increment our counter
	redisObject.Client.Incr("counter")

	if !redisObject.Pong {
		fmt.Fprintf(w, "The counter is unavailable at this time")
	} else {
		// Get our counter...
		val, err := redisObject.Client.Get("counter").Result()
		if err != nil {
			fmt.Println("got an error before healthcheck")
			fmt.Fprintf(w, "The counter is unavailable at this time")
		} else {
			// And write out
			fmt.Fprintf(w, "The counter is: %s", val)
		}
	}
}

func (redisObject *Redis) ping() {
	// Get our ping result
	_, err := redisObject.Client.Ping().Result()

	// If the error is not empty, we have an error.. so Pong is false.
	// Lets initiate this further by setting our error to our first tick.
	if err != nil {
		// It was already false, lets increase our backoff
		if redisObject.Pong == false {
			redisObject.ExponentialBackoff = true
		}
		redisObject.Pong = false
		redisObject.ErrorCounter = 1
	} else {
		// Otherwise we get a pong, and we can reset our error counter

		// Let's check if it was false before, so we don't spam our logs. Should hit on init..
		if redisObject.Pong == false {
			fmt.Println("Redis is available!")
		}
		redisObject.ExponentialBackoff = false
		redisObject.Pong = true
		redisObject.ErrorCounter = 0
	}
}
