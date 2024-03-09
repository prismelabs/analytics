<h1 align="center">
    <img height="250" src="./.github/images/logo.jpg">
</h1>

<p align="center">
    <a href="https://goreportcard.com/report/github.com/prismelabs/analytics">
        <img src="https://goreportcard.com/badge/github.com/prismelabs/analytics">
    </a>
    <a href="https://github.com/prismelabs/analytics/raw/master/LICENSE">
        <img src="https://img.shields.io/github/license/prismelabs/analytics">
    </a>
    <a href="https://hub.docker.com/r/prismelabs/analytics">
        <img alt="Docker Image Size (tag)" src="https://img.shields.io/docker/image-size/prismelabs/analytics/latest">
    </a>
    <img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/prismelabs/analytics">
</p>

# :gem: Prisme Analytics

An Open Source, privacy-focused and progressive analytics service.

[Documentation](https://www.prismeanalytics.com/docs)
|
[Live demo](https://app.prismeanalytics.com/grafana)

![grafana dashboard](.github/images/builtin-dashboard.jpg)

## Configuration

You can find configuration documentation on
[our website](https://www.prismeanalytics.com/docs/set-up/configuration/configure-server/server-modes).

## Performance

Prisme Analytics ingestion server is easily capable of ingesting ~8000 req/s on
AMD Ryzen 7 7840U w/ Radeon 780M Graphics using ~70% of a CPU core and ~200M of RAM.

```
$ cat /proc/cpuinfo | head | grep 'model name'
model name	: AMD Ryzen 7 7840U w/ Radeon  780M Graphics

$ cd tests/k6/events
$ make start test clean

          /\      |‾‾| /‾‾/   /‾‾/   
     /\  /  \     |  |/  /   /  /    
    /  \/    \    |     (   /   ‾‾\  
   /          \   |  |\  \ |  (‾)  | 
  / __________ \  |__| \__\ \_____/ .io

     execution: local
        script: /data/main.js
        output: -

     scenarios: (100.00%) 1 scenario, 4096 max VUs, 10m30s max duration (incl. graceful stop):
              * sharedIterations: 131072 iterations shared among 4096 VUs (maxDuration: 10m0s, gracefulStop: 30s)

     data_received..................: 28 MB  1.7 MB/s
     data_sent......................: 34 MB  2.0 MB/s
     http_req_blocked...............: avg=3.95ms   min=652ns    med=2.41µs   max=251.86ms p(90)=3.8µs    p(95)=5.7µs   
     http_req_connecting............: avg=1.81ms   min=0s       med=0s       max=188.89ms p(90)=0s       p(95)=0s      
     http_req_duration..............: avg=3.95ms   min=47.35µs  med=628.69µs max=111.45ms p(90)=5.45ms   p(95)=17.68ms 
       { expected_response:true }...: avg=3.91ms   min=47.35µs  med=620.01µs max=111.45ms p(90)=5.42ms   p(95)=17.04ms 
     http_req_failed................: 60.12% ✓ 78806       ✗ 52266 
     http_req_receiving.............: avg=285.59µs min=3.05µs   med=9.43µs   max=47.99ms  p(90)=20.94µs  p(95)=99.66µs 
     http_req_sending...............: avg=834.84µs min=3.41µs   med=9.53µs   max=89.07ms  p(90)=286.23µs p(95)=1.35ms  
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s       max=0s       p(90)=0s       p(95)=0s      
     http_req_waiting...............: avg=2.83ms   min=39.71µs  med=556.52µs max=78.53ms  p(90)=4.53ms   p(95)=13.22ms 
     http_reqs......................: 131072 7970.091392/s
     iteration_duration.............: avg=509.35ms min=500.09ms med=501.33ms max=786.3ms  p(90)=509.57ms p(95)=528.11ms
     iterations.....................: 131072 7970.091392/s
     vus............................: 4096   min=4096      max=4096
     vus_max........................: 4096   min=4096      max=4096


running (00m16.4s), 0000/4096 VUs, 131072 complete and 0 interrupted iterations
sharedIterations ✓ [ 100% ] 4096 VUs  00m16.4s/10m0s  131072/131072 shared iters
```

## Contributing

If you want to contribute to `prismeanalytics` to add a feature or improve the
code, open an [issue](https://github.com/prismelabs/analytics/issues)
or make a [pull request](https://github.com/prismelabs/analytics/pulls).

## :stars: Show your support

Please give a :star: if this project helped you!

## :scroll: License

AGPL-3.0 © [Prisme Analytics](https://www.prismeanalytics.com/)
