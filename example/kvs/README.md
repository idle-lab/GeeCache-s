

## introduce

In this example, we will create a simple in-memory kv cache server by geecache-s.

## how to use

You can start three node in a machine by command follow:

```shell
./start.sh
```

Kvs server provide two API interfaces.

### Get a kv

Method: `GET`

URL: `http://localhost:9999/get`

Query parameter: 

- key：string

eg:

```shell
curl http://localhost:9999/get?key=abc
```

### Add a kv

Method: `POST`

URL: `http://localhost:9999/add`

Body:

I use JSON format for data transmission here

```json
{
    "key": "xxx",
    "value": "xxxx",
}
```

eg:

```shell
curl -X POST -d '{"key": "name", "value": "I am kvs"}' http://localhost:9999/add
```
