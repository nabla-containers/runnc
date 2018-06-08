1. Make binaries and copy to bin dir
```
make build
cp runnc /usr/local/bin/runnc
# In addition, copy nabla_run and ukvm_bin to /usr/local/bin and chmod +x them
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
