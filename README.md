# hetzner-server-market-exporter

[![Build Status](https://travis-ci.org/andrexus/hetzner-server-market-exporter.svg?branch=master)](https://travis-ci.org/andrexus/hetzner-server-market-exporter)

Prometheus exporter for Hetzner Server Market API

Start locally in docker:
```
docker run --rm -p 8080:8080 -v $(pwd)/creds.json:/etc/creds.json andrexus/hetzner-server-market-exporter -robot-api-credentials /etc/creds.json
```


## Licence

[MIT License](https://raw.githubusercontent.com/andrexus/hetzner-server-market-exporter/master/LICENSE.txt)

## Author

[andrexus](https://github.com/andrexus)
