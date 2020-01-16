# packetjq

A tool to parse JSON from the Packet API

It's just a shortcut to

```
curl -X GET -H "X-Auth-Token: $PACKET_AUTH_TOKEN" https://api.packet.net/projects | jq ".projects[].name"
```

instead, you can do just

```
./packetjq -p projects -q ".projects[].name"
```

## Usage

You need to set your Packet API token to envvar `PACKET_AUTH_TOKEN`.
 
### List devices in all projects

```
./packetjq -p "projects?include=devices" -q ".projects[].devices[].hostname"
```

### Create new read-only API key

```
./packetjq -p "user/api-keys" -m POST -r '{"description": "newKey", "read_only": true}'
```

### Create Project API key

```
tomk@xps ~/packetjq ±master » ./packetjq -p "projects/1ff4bd4e-5901-4b39-9b19-58d619132322/api-keys" -m POST -r '{"description": "aaaa2", "read_only": true}'
```

### Delete API key

```
./packetjq -p "user/api-keys/958c3495-0331-40c5-bb80-d946a8e6df05" -m DELETE
```

Using https://github.com/itchyny/gojq for the JSON parsing.

