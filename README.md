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

## Why Prisme ?

Here's what makes Prisme a great alternative to other analytics:
* **Tailored analytics**: Prisme Analytics supports **custom dashboards and events**
so you can collect, visualize analyze metrics that matters to you the way you want.
* **Ready to use**: Prisme also comes with built-ins metrics and dashboards 
(see [demo](https://app.prismeanalytics.com/grafana)).
* **Privacy-focused**: Prisme is **GDPRC, CCPA, PECR compliant by design**, no
Personally Identifiable Information (PII) is collected. Tracking script doesn't
use cookies, you can ditch your cookie pop-up.
* **Lightweight**: Prisme tracking script is less than 1kB (~45x smaller than
Google Analytics).
* **Resource efficient**: Prisme is designed to be fast and resource efficient,
checkout our [benchmarks](#performance).
* **SPA support**: Prisme is built with modern web frameworks in mind and it works
automatically with any pushState based router on the frontend.
* **[Grafana](https://github.com/grafana/grafana) based**: Prisme integrates with 
[Grafana](https://github.com/grafana/grafana) that provides:
  * User managements
  * Team managements
  * Permissions managements
  * Multi organizations support
  * Custom dashboards

## Configuration

You can find configuration documentation on
[our website](https://www.prismeanalytics.com/docs/set-up/configuration/configure-server/server-modes).

## Performance

Prisme Analytics ingestion server is **easily** capable of ingesting more than 
50,000 req/s on my AMD Ryzen 7 7840U w/ Radeon 780M Graphics.

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

     scenarios: (100.00%) 2 scenarios, 8192 max VUs, 10m30s max duration (incl. graceful stop):
              * sharedIterationsCustomEvents: 524288 iterations shared among 4096 VUs (maxDuration: 10m0s, exec: customEvent, gracefulStop: 30s)
              * sharedIterationsPageViewEvents: 524288 iterations shared among 4096 VUs (maxDuration: 10m0s, exec: pageView, gracefulStop: 30s)

     data_received..................: 223 MB  11 MB/s
     data_sent......................: 314 MB  16 MB/s
     http_req_blocked...............: avg=1.3ms    min=350ns    med=2.13µs   max=452.4ms  p(90)=3.22µs   p(95)=3.92µs
     http_req_connecting............: avg=1.26ms   min=0s       med=0s       max=452.35ms p(90)=0s       p(95)=0s
     http_req_duration..............: avg=116.06ms min=65.58µs  med=102.9ms  max=737.66ms p(90)=205.62ms p(95)=237.97ms
       { expected_response:true }...: avg=116.48ms min=69.01µs  med=103.26ms max=737.66ms p(90)=206.26ms p(95)=238.63ms
     http_req_failed................: 33.35%  ✓ 349785       ✗ 698791
     http_req_receiving.............: avg=3.23ms   min=4.27µs   med=11.88µs  max=334.49ms p(90)=200.62µs p(95)=10.72ms
     http_req_sending...............: avg=431.43µs min=3.53µs   med=8.68µs   max=334.28ms p(90)=41.42µs  p(95)=123.74µs
     http_req_tls_handshaking.......: avg=0s       min=0s       med=0s       max=0s       p(90)=0s       p(95)=0s
     http_req_waiting...............: avg=112.4ms  min=41.87µs  med=102.49ms max=546.44ms p(90)=196.89ms p(95)=219.73ms
     http_reqs......................: 1048576 52278.070656/s
     iteration_duration.............: avg=135.96ms min=871.67µs med=118.11ms max=739.61ms p(90)=228.58ms p(95)=262.73ms
     iterations.....................: 1048576 52278.070656/s
     vus............................: 8192    min=8192       max=8192
     vus_max........................: 8192    min=8192       max=8192


running (00m20.1s), 0000/8192 VUs, 1048576 complete and 0 interrupted iterations
sharedIterationsCustomEvents   ✓ [ 100% ] 4096 VUs  00m20.1s/10m0s  524288/524288 shared iters
sharedIterationsPageViewEvents ✓ [ 100% ] 4096 VUs  00m20.1s/10m0s  524288/524288 shared iters

```

## Contributing

If you want to contribute to `prismeanalytics` to add a feature or improve the
code, open an [issue](https://github.com/prismelabs/analytics/issues)
or make a [pull request](https://github.com/prismelabs/analytics/pulls).

## :stars: Show your support

Please give a :star: if this project helped you!

## :scroll: License

[Prisme Analytics](https://www.prismeanalytics.com/) is distributed under 
[AGPL-3.0-only](LICENSE). For MIT exceptions, see [LICENSING.md](LICENSING.md)
