# Skywatch

## Features

- serialize json string to message pack
- deserialize message pack to json string

## Installation

```bash
go install github.com/blackhorseya/skywatch
```

## Usage

```bash
skywatch -h
```

### Encode json to message pack

```bash
skywatch encode json <input json string>
```

### Decode message pack to json

```bash
skywatch decode <input msgpack hex string>
```
