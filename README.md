# addressdb

## Status Of The Project

Ready to be used in prod

## Overview

addressdb demonstrates the use of Bloom filters to share sets of blockchain addresses while preserving privacy. It provides tools for generating Ethereum addresses, encoding them into a Bloom filter, and checking addresses against the filter.

## Key Features

- Generate large sets of Ethereum addresses
- Encode addresses into a space-efficient Bloom filter
- Check if an address is potentially in the set
- Preserve privacy while sharing address sets

## How It Works

The project uses a Bloom filter, stored in a `.gob` file. This data structure allows for efficient storage and querying of large sets of data with a controllable false-positive rate.

Example: A set of 1 million Ethereum addresses results in a `.gob` file of approximately 3.5MB (uncompressed). The actual size depends on the number of addresses and the chosen false-positive rate.

The `.gob` file can be easily shared across data pipelines. Adding key-pair encryption (not implemented in this version) would further enhance confidentiality.

## Installation

```bash
git clone https://github.com/your-username/addressdb.git
cd addressdb
go mod tidy
```

## Library Usage

Create a new Bloom filter store and add addresses to it:
```go
// Create an EVM address handler that will be used to validate and encode addresses
addressHandler := &address.EVMAddressHandler{}

// Create a new BloomFilterStore to store 10000 addresses with a false positive rate of 0.0000001
store, _ := NewBloomFilterStore(addressHandler)

store.AddAddress("0x1234567890123456789012345678901234567890")

if ok, _ := store.CheckAddress("0x1234567890123456789012345678901234567890"); ok {
    fmt.Println("Address found in the Bloom filter")
} else {
    fmt.Println("Address not found in the Bloom filter")
}

store.SaveToFile(filePath)
```

Or load the bloom filter from a file:
```go
// Create a new BloomFilterStore and load from the file
store, _ := NewBloomFilterStoreFromFile(filePath, addressHandler)

if ok, _ := store.CheckAddress("0x1234567890123456789012345678901234567890"); ok {
    fmt.Println("Address found in the Bloom filter")
} else {
    fmt.Println("Address not found in the Bloom filter")
}
```

### Auto-reloading from file when the file changes
```go

// Create a file watcher notifier, that will reload the Bloom filter when the file changes.  
// But never more than once every 2 seconds.
addressHandler := &address.EVMAddressHandler{}
store, _ := NewBloomFilterStoreFromFile(filePath, addressHandler)
notifier, _ := reload.NewFileWatcherNotifier(filePath, 2*time.Second)

// Create the ReloadManager with the notifier.
manager := reload.NewReloadManager(filter, notifier)
manager.Start(context.Background())
defer manager.Stop()

```

### Using pgp encrypted and sign files
```go
// Create the pgp secure data handler
pgpHandler := securedata.NewPGPSecureDataHandler(
	withPublicKeyPath(publicKeyPath),
	withPrivateKeyPath(privateKeyPath, passphrase),
)

// Create a new Bloom filter store from a pgp encrypted file
addressHandler := &address.EVMAddressHandler{}
store, _ := NewBloomFilterStoreFromFile(filePath, addressHandler, WithSecureDataHandler(pgpHandler))
````

## CLI Usage

### Step 1: Generate Ethereum Addresses (Optional)

```bash
go run cmd/cli/main.go generate-addresses -n 1000000
```

Generates 1 million Ethereum addresses and stores them in `address.txt`.

### Step 2: Build the Bloom Filter

```bash
go run cmd/cli/main.go encode -n 1000000 -p 0.000001
```

- `-n`: Number of entries (should match the number of generated addresses)
- `-p`: False positive rate

Creates a `bloomfilter.gob` file containing the Bloom filter.

### Step 3: Use the Filter

Interactive mode:

```bash
go run cmd/cli/main.go check -f bloomfilter.gob
```

Batch mode:

```bash
cat my_addresses.txt | go run cmd/cli/main.go batch-check -f bloomfilter.gob
```

## Large-Scale Example

Building a Bloom filter with 24 million Ethereum addresses:

```bash
go run cmd/cli/main.go encode -n 1000000000 -p 0.000001 -input ~/path/to/eth_all.csv

# Check addresses
go run cmd/cli/main.go check -f bloomfilter.gob
```

Result: `bloomfilter.gob` file of about 3.3GB.

## Performance

- Encoding 24M addresses: ~20 seconds on a standard machine.
- Constant-time complexity for adding and checking addresses.

### Benchmarks

```bash
# Debug build
time target/debug/pa-cli check -input ~/Downloads/eth_all.csv
# 13.59s user 0.33s system 97% cpu 14.211 total

# Release build
time target/release/pa-cli encode -input ~/Downloads/eth_all.csv
# 4.32s user 0.26s system 94% cpu 4.835 total
```

#### Large-scale performance

- Loading a 51GB Bloom filter (1.4 billion addresses): 18 seconds
- Checking 1173 addresses: ~2ms

#### Small-scale performance

```bash
go run cmd/cli/main.go encode -n 100000 -p 0.0000001
```

Results in a ~450KB filter.

#### Micro-benchmarks

```bash
go test -bench=. -benchmemstore % go test -bench=. -test.benchmem        
goos: darwin
goarch: arm64
pkg: addressdb/store
BenchmarkAddAddress-16                  11823907               103.5 ns/op            48 B/op          1 allocs/op
BenchmarkBloomFilterTestNaive-16         5685646               212.2 ns/op            95 B/op          1 allocs/op
BenchmarkCheckAddress-16                14392717                82.72 ns/op           48 B/op          1 allocs/op
BenchmarkBloomFilterNaiveCheck-16        5969546               214.6 ns/op            95 B/op          1 allocs/op
```

## Limitations

- False-positive rate, but no false negatives
- No encryption for the `.gob` file in current implementation
- Potential for brute-force recovery of entries if the value space is limited

## Future Improvements

- Implement encryption for `.gob` files
- Optimize Bloom filter parameters for various use cases
- Add a web interface for easier interaction

## Contributing

Contributions are welcome! Please submit a Pull Request.​​​​​​​​​​​​​​​​

Yes we are welcome to any suggestion including a better name for the project.
