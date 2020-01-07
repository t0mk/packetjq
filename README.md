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

Using https://github.com/itchyny/gojq for the JSON parsing.

