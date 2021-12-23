# Small test suite

## Installing Requirements

```npm install```

## Testing

```npm test```

if the taget testing is mongodb, you should define a different invalid id since mongo use objectid:

```CYPRESS_INVALID_ID=507f1f77bcf86cd799439011 npm test```


Other env variables:

  - CYPRESS_API_PORT: Port which the server is listening
