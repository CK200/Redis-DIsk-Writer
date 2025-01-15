# Redis to Disk Writer

## Introduction
Redis to Disk Writer is a utility tool designed to efficiently transfer data from Redis to persistent storage. This tool is ideal for backup purposes, data migration, or simply safeguarding data against potential data loss in volatile memory.

## Prerequisites
Before you begin, ensure you have the following installed:
- Redis server (version x.x or higher)
- Go (version 1.x or higher)

## Installation
To install the necessary components for this project, follow these steps:
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/redis-disk-writer.git
   ```
2. Navigate to the project directory:
   ```bash
   cd redis-disk-writer
   ```

## Usage
To use the Redis to Disk Writer, perform the following steps:
1. Ensure your Redis server is running.
2. Execute the program with the necessary parameters:
   ```bash
   go run cmd/main.go <queue_name_1> <queue_name_2> ...
   ```
   Replace `<queue_name_1>`, `<queue_name_2>`, etc. with the names of the Redis queues you want to process. Multiple queue names can be provided as separate arguments.

## Contributing
Contributions to the Redis to Disk Writer are welcome! Here's how you can contribute:
1. Fork the repository.
2. Create a new branch for your feature (`git checkout -b feature/your_feature_name`).
3. Commit your changes (`git commit -am 'Add some feature'`).
4. Push to the branch (`git push origin feature/your_feature_name`).
5. Open a new Pull Request.

## License
This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.