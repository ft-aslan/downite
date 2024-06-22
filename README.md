# Downite Self-Hostable Torrent and URL Download Client
![downite-torrents](https://github.com/ft-aslan/downite/assets/13184550/0ade4fc1-798b-4164-85e8-51f40f28ca20)



![downite-home](https://github.com/ft-aslan/downite/assets/13184550/eb9184ed-ec40-4fb1-9dd3-e60ffa09148a)

## About

This project is a self-hostable torrent and URL download client designed to run on the same server. It provides a robust solution for managing and downloading torrents and files directly from URLs, all from a unified interface. Future updates will include seamless integration with popular media management applications such as Sonarr and Radarr, allowing for automated media acquisition and organization.

## Features

- **Torrent Client**: Efficiently download and manage torrents.
- **URL Download Client**: Directly download files from URLs.
- **Unified Interface**: Manage both torrents and URL downloads from a single interface.
- **Future Integration**: Planned support for Sonarr and Radarr to automate and streamline media management.

## Planned Integrations

- **Sonarr**: Automatically download and manage TV shows.
- **Radarr**: Automatically download and manage movies.

This project aims to provide a comprehensive, all-in-one solution for media downloading and management, making it an essential tool for any media enthusiast.

## How To Install

### Prerequisites for building from source code

- [Bun](https://github.com/oven-sh/bun) package manager
- [Go](https://go.dev/)

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/ft-aslan/downite.git
   ```
2. Navigate to the project directory:
   ```sh
   cd downite
   ```
3. Install dependencies:
   ```sh
   bun install
   ```
4. Build server:
   ```sh
   bun run build:server
   ```
5. Build web client:
   ```sh
   bun run build:web
   ```
6. Run server:
   ```sh
   bun run start:server
   ```
7. Run server:
   ```sh
   bun run start:web
   ```

- Web client is running on port 4173 by default. http://localhost:4173
- Server is running on port 9999 by default. http://localhost:9999
- Documentation link. http://localhost:9999/docs

## How to Develop

### Prerequisites for developing

- [Goose](https://github.com/pressly/goose) database migration tool
- [Bun](https://github.com/oven-sh/bun) package manager
- [Go](https://go.dev/)
- (Optional) [Air](https://github.com/air-verse/air) live reload tool

### Running

1. Clone the repository:
   ```sh
   git clone https://github.com/ft-aslan/downite.git
   ```
2. Navigate to the project directory:
   ```sh
   cd downite
   ```
3. Install dependencies:
   ```sh
   bun install
   ```
4. Start the server in development mode with Air or without Air:
   - air
   ```sh
   bun run dev:server
   ```
   - vanilla
   ```sh
   bun run dev:server:nohot
   ```
5. Start the web client in development mode:
   ```sh
   bun run dev:web
   ```

- Web client is running on port 4173 by default. http://localhost:4173
- Server is running on port 9999 by default. http://localhost:9999
- Documentation. http://localhost:9999/docs

### Migrations

- Migration down
  ```sh
  bun run db down
  ```
- Migration up
  ```sh
  bun run db up
  ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
