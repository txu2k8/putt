# retry
A simple, stateless, functional mechanism to perform actions repetitively until successful.

## Install

`go get github.com/xxx/retry`

## Examples

### Basic

```go
retry.Retry(func(attempt uint) error {
	return nil // Do something that may or may not cause an error
})
```

### File Open

```go
const logFilePath = "/var/log/myapp.log"

var logFile *os.File

err := retry.Retry(func(attempt uint) error {
	var err error

	logFile, err = os.Open(logFilePath)

	return err
})

if nil != err {
	log.Fatalf("Unable to open file %q with error %q", logFilePath, err)
}

logFile.Chdir() // Do something with the file
```

### HTTP request with strategies and backoff

```go
var response *http.Response

action := func(attempt uint) error {
	var err error

	response, err = http.Get("https://api.github.com/repos/xxx/retry")

	if nil == err && nil != response && response.StatusCode > 200 {
		err = fmt.Errorf("failed to fetch (attempt #%d) with status code: %d", attempt, response.StatusCode)
	}

	return err
}

err := retry.Retry(
	action,
	strategy.Limit(5),
	strategy.Backoff(backoff.Fibonacci(10*time.Millisecond)),
)

if nil != err {
	log.Fatalf("Failed to fetch repository with error %q", err)
}
```

### Retry with backoff jitter

```go
action := func(attempt uint) error {
	return errors.New("something happened")
}

seed := time.Now().UnixNano()
random := rand.New(rand.NewSource(seed))

retry.Retry(
	action,
	strategy.Limit(5),
	strategy.BackoffWithJitter(
		backoff.BinaryExponential(10*time.Millisecond),
		jitter.Deviation(random, 0.5),
	),
)
```
