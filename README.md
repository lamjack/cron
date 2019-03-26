# Golang Cron

Package cron implements a cron spec parser and job runner.
Base on [https://godoc.org/github.com/robfig/cron](https://godoc.org/github.com/robfig/cron), added JOB delete function.

## Getting Started

### Installing
```bash
go get -u github.com/lamjack/cron
```

### Using
```golang
c := cron.New()
c.AddFunc("* * * * * *", func() {
    fmt.Println(time.Now().Format("15:04:05"), "every second")
})
c.Start()

// Dynamically add a scheduled task after 5 seconds, run once every 2 seconds
time.Sleep(5 * time.Second)
id, _ := c.AddFunc("*/2 * * * * *", func() {
    fmt.Println(time.Now().Format("15:04:05"), "every 2 seconds")
})

// Delete the newly added task after 10 seconds
time.Sleep(10 * time.Second)
c.DeleteJob(string(id))

select {}
```

### Document
Documentation here: [https://godoc.org/github.com/robfig/cron](https://godoc.org/github.com/robfig/cron)
