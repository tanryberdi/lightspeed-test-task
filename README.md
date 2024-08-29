# Unique IP Address Counter

This Go program efficiently counts the number of unique IPv4 addresses in a large text file. It's designed to handle files of unlimited size, potentially occupying tens or hundreds of gigabytes, while using minimal memory and processing time.

## Features

- Memory-efficient: Uses a bitmap to represent the entire IPv4 address space.
- Fast processing: Employs parallel processing with multiple worker goroutines.
- Large file handling: Processes the file in chunks, allowing for very large inputs.
- Optimized a bit counting: Uses an efficient algorithm for counting set bits.

## Requirements

- Go 1.13 or later

## Usage

1. Clone this repository:
```shell
git clone https://github.com/tanryberdi/lightspeed-test-task.git
cd lightspeed-test-task
```

2. Build the program:
```shell
go build main.go
```

3. Run the program with the path to the input file:
```shell
./main path/to/your/ip_list.txt
```

## Input File Format

The input file should contain IPv4 addresses, one per line. For example:
```text
192.168.1.1
10.0.0.1
172.16.0.1
192.168.1.1
```

## Implementation Details

- The program uses a custom bitmap implementation with a slice of uint64 for efficient memory usage and fast bit operations.
- It processes the file in parallel using multiple worker goroutines.
- IP addresses are parsed byte-by-byte for maximum efficiency.
- Atomic operations are used when merging results from different workers to ensure thread safety.
- A worker pool is employed to reuse memory and reduce allocations.

## Performance

This implementation is designed to be more efficient than a naive approach using a HashSet. It should handle very large files with billions of IP addresses efficiently.

## Limitations

- The program is specifically designed for IPv4 addresses and won't work with IPv6 addresses.
- The count of unique addresses is limited to 2^32 (the size of the IPv4 address space).
- Didn't check with a big file that you shared in the task description. Really sorry for that.
