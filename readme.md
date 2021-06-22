# Go Screen Monitor - Screen Monitor Software

Go Screen Monitor allows a remotely installed service to capture screenshots at a specified framerate for any logged in user and provide remote access to them via the web interface.

## Sample

The client and server are both located in the cmd directory. Use go build to create the executables, then you can use the CLI arguments specified below.

## Server

To run the server, copy the `credentials.json.sample` to `credentials.json` in the same directory as the built executable, then modify to add user logins. To run the server, use the following:
```shell
smserver.exe -mserver :3000 -wserver :8080
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