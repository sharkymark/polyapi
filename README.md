# An API app written in Go

polyapi retrieves information from many (poly) disparate APIs within a CLI app that you may have historically retrieved through web pages. Go is a cross-platform, statically typed and compiled programming language. A dev container (development container) is a self-contained dev environment with a Docker container allowing portability and instant creation of a dev environment without manual installation.

## Lastest update

This is a new project and actively under development.

## Currently implemented functionality
1. Reads environment variables like API keys
1. Shows weather by geocoding an address entered
1. Shows stock ticker data
1. Stores validated addresses and ticker symbols in a local SQLite3 database for re-use or deletion
1. Also stores the last temperature and last ticker price on each API call with an `updated_at` timestamp
1. Retrieves latest average US Treasury bond, note and bill rates and app calculates spreads. 
1. Retrieves latest PPI and CPI from U.S. Bureau of Labor Statistics


![screenshot of main menu and retrieving weather](./docs/images/polyapi-address-weather.png)

## SQLite3 local database

At program start, a db directory and `polyapi.db` are created. `db/polyapi.db` is added to a `.gitignore` file so it will not be included in the code repository.

## dev container

I'm using a dev container so I don't have to install Go on my Mac. All I need a is a Docker daemon, which in my case is `colima` and VS Code with the dev container extension.

## API providers

### Weather for Addresses (USGOV)

NOAA's API provides weather forecast information but requires latitude and longitude coordinates. The U.S. Census bureau has a geocoding API that returns coordinates based on a valid address. No API keys required for both APIs.

### Stock Quotes (Alpha Vantage)

Get a [free API key](https://www.alphavantage.co/support/#api-key) to retrieve stock quote data. Add an environment variable in your configuration script e.g., `.zshrc` or `bashrc` that the dev container reads. 

```sh
export ALPHAVANTAGE_API_KEY=""
```

### US Treasury Rates (USGOV)

The U.S. Treasury has a [public API](https://fiscaldata.treasury.gov/api-documentation/) to retrieve financial data including their [rate API](https://fiscaldata.treasury.gov/datasets/average-interest-rates-treasury-securities/average-interest-rates-on-u-s-treasury-securities#api-quick-guide) for [average treasury rates](https://api.fiscaldata.treasury.gov/services/api/fiscal_service/v2/accounting/od/avg_interest_rates?sort=-record_date).  No API key required.

### US PPI and CPI (USGOV)

The U.S. Bureau of Labor Statistics has a [public API](https://www.bls.gov/developers/api_faqs.htm) to retrieve the producer price index "PPI" and consumer price index "CPI"

### US Federal Reserve (USGOV)

The U.S. Federal Reserve has a [public API called FRED](https://www.bls.gov/developers/home.htm#) to retrieve economic statistics. It does require an [API key](https://fred.stlouisfed.org/docs/api/api_key.html)

## Creating a binary

`.gitignore` is set to ignore `main` and `polyapi` binaries to reduce repository size. `git build .` uses the OS and architecture of the development machine. To create other binaries, use examples like:

```sh
GOOS=darwin GOARCH=arm64 go build -o polyapi_darwin_arm64
GOOS=linux GOARCH=amd64 go build -o polyapi_linux_amd64
GOOS=windows GOARCH=amd64 go build -o polyapi_windows_amd64.exe
```

## Resources

[Go](https://go.dev/)

[Dev Container specification](https://containers.dev/implementors/spec/)

## License

This project is licensed under the [MIT License](LICENSE)

## Contributing

### Disclaimer: Unmaintained and Untested Code

Please note that this program is not actively maintained or tested. While it may work as intended, it's possible that it will break or behave unexpectedly due to changes in dependencies, environments, or other factors.

Use this program at your own risk, and be aware that:
1. Bugs may not be fixed
1. Compatibility issues may arise
1. Security vulnerabilities may exist

If you encounter any issues or have concerns, feel free to open an issue or submit a pull request.