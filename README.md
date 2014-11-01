# AOI Proxy 青いプロキシ

A simple proxy for switching between blue and green environments.

The proxy is configured with an address to a blue environment and the address of a green environment.
The proxy will listen on a specified port and send all requests to the blue environment.

Using a web api, you can toggle the environment to green.
Then the proxy will start sending all requests to the green environment. 

## Roadmap
- [x] Add test handler to test the "s-out" environment on a different port
 - [ ] Do I need to lock the ENVIRONMENT flag with a mutex or something??
- [ ] Find a better way to toggle/check the ENVIRONMENT flag (something like an enum)
- [ ] Command Line flags to change the ports
- [ ] use Environment variables so its easy to put in Docker
- [ ] Improve the admin api 
 - [ ] json-ify the output
 - [ ] check status api
 - [ ] specify environment instead of toggle

