1. Make binaries and copy to bin dir
```
# Get the neceesary binaries for the runtime
make build

# Copy the binaries to /usr/local/bin
sudo hack/copy_bins.sh
```

2. Modify to add runtime to `/etc/docker/daemon.json`, for example:
```
{
    "default-runtime": "runc",
    "runtimes": {
        "runsc": {
            "path": "/usr/local/bin/runsc",
            "runtimeArgs": [
                "--network=sandbox"
            ]
       },
        "runn": {
                "path": "/usr/local/bin/runnc",
                "runtimeArgs": [
                ]
        }
    }
}
```

3. Restart docker 

```systemctl restart docker```

4. Run with runtime:

```docker run --rm --runtime=runnc lumjjb/node-nable:latest```
