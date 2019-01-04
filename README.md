[![Build Status](https://travis-ci.org/cloverstd/tcping.svg?branch=master)](https://travis-ci.org/cloverstd/tcping)

# tcping

tcping is like [tcping.exe](https://elifulkerson.com/projects/tcping.php), but written with Golang.


## Usage

* The default count of ping is 4.

* If the port is omitted, the default port is 80.

* The default interval of ping is 1s.

* The default timeout of ping is 1s.

### ping tcp

```bash
> tcping google.com 443
Ping tcp://google.com:443 - Connected - time=15.425732ms
Ping tcp://google.com:443 - Connected - time=2.628025ms
Ping tcp://google.com:443 - Connected - time=2.400356ms
Ping tcp://google.com:443 - Connected - time=1.967587ms

Ping statistics tcp://google.com:443
	4 probes sent.
	4 successful, 0 failed.
Approximate trip times:
	Minimum = 1.967587ms, Maximum = 15.425732ms, Average = 5.605425ms
```

### ping http

```bash
> tcping -H hui.lu
Ping http://hui.lu:80 - http is open - time=232.880173ms method=GET status=200 bytes=10317
Ping http://hui.lu:80 - http is open - time=60.096446ms method=GET status=200 bytes=10317
Ping http://hui.lu:80 - http is open - time=56.750403ms method=GET status=200 bytes=10317
Ping http://hui.lu:80 - http is open - time=57.886907ms method=GET status=200 bytes=10317

Ping statistics http://hui.lu:80
	4 probes sent.
	4 successful, 0 failed.
Approximate trip times:
	Minimum = 56.750403ms, Maximum = 232.880173ms, Average = 101.903482ms
```
