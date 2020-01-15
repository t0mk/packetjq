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
{
  "id": "6973b150-3a8d-471a-860e-m423io4no234",
  "token": "f049vm309m09mve0FDdsgdfg04GFgdfg",
  "created_at": "2020-01-15T14:39:06Z",
  "updated_at": "2020-01-15T14:39:06Z",
  "description": "newKey",
  "user": {
    "href": "/users/ef43523e-8800-44ff-a31f-edd1a2cbf86d"
  },
  "read_only": true
}
```

### Delete API key

```
./packetjq -p "user/api-keys/958c3495-0331-40c5-bb80-d946a8e6df05" -m DELETE
```

Using https://github.com/itchyny/gojq for the JSON parsing.

