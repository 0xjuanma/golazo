<p align="center">
  <img src="assets/golazo-demo.gif" alt="Golazo demo" width="600">
</p>

# Golazo 

> [!IMPORTANT]
> Tool is in active development.

A minimalist terminal user interface (TUI) for following football matches in real-time. Get live match updates, finished match statistics, and minute-by-minute events directly in your terminal.

## Installation

### Using the install script (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/0xjuanma/golazo/main/scripts/install.sh | bash
```

### Manual installation

1. Clone the repository:
```bash
git clone https://github.com/0xjuanma/golazo.git
cd golazo
```

2. Build the binary:
```bash
go build -o golazo ./cmd/golazo
```

3. (Optional) Move to your PATH:
```bash
sudo mv golazo /usr/local/bin/
```

## Usage

Run the application:
```bash
golazo
```

### Options

- `--mock`: Use mock data for all views instead of real API data
  ```bash
  golazo --mock
  ```

### Environment Variables

For the Stats view (finished matches), set your API-Sports.io API key:
```bash
export FOOTBALL_DATA_API_KEY="your-api-key-here"
```

Get your free API key at: https://www.api-sports.io/


