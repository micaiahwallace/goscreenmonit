# Go Screen Monitor - Screen Monitor Software

Go Screen Monitor allows a remotely installed service to capture screenshots at a specified framerate for any logged in user and provide remote access to them via the web interface.

## Sample

The client and server are both located in the cmd directory. Use go build to create the executables, then you can use the CLI arguments specified below.

## Server

To run the server, do the following:
1. copy the `credentials.json.sample` to `credentials.json` in the same directory as the built executable, then modify to add user logins. 
2. run `genkeypair.sh` to generate a tls key pair used for the web server and for agent to server communication.
3. ensure you have a recent version of nodejs installed then run `npm install && npm run build` inside the ui directory.
4. before running, ensure you have the following directory structure setup:

```
/goscreenmonit        (root project directory)
  - smserver          (server binary built with `go build ./cmd/smserver`)
  - server.key        (tls key file)
  - server.crt        (tls certificate file)
  - credentials.json  (web server authentication credentials)
  - ui                (ui source directory)
    - build           (directory of the built ui source files)
```

To run the server, use the following:
```shell
$ ./smserver -mserver :3000 -wserver :8080
```

## Client

The client can be run with the following command.

```shell
smclient.exe -server 192.168.1.5:3000 -fps 5
```

You can also install the client on a windows pc with:

```shell
smclient.exe install watch -server 192.168.1.5:3000 -fps 5
```

This will install a watchdog service that will run on every subsequent user login with the specified parameters.

## Todo

- [ ] Increase security validation between agent and server
- [ ] Provide a better authentication mechanism
- [ ] Clean up the POC user interface
- [ ] Increase data transmission efficiency from web server to UI
- [ ] Look into a more efficient screen capture option